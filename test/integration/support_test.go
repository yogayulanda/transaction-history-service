//go:build integration

package integration

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/microsoft/go-mssqldb"
	"github.com/segmentio/kafka-go"
	coreapp "github.com/yogayulanda/go-core/app"
	coreconfig "github.com/yogayulanda/go-core/config"
	coreErrors "github.com/yogayulanda/go-core/errors"
	"github.com/yogayulanda/go-core/messaging"
	coremigration "github.com/yogayulanda/go-core/migration"
	coreserver "github.com/yogayulanda/go-core/server"
	"github.com/yogayulanda/transaction-history-service/internal/app"
	handlerkafka "github.com/yogayulanda/transaction-history-service/internal/handler/kafka"
	"github.com/yogayulanda/transaction-history-service/internal/repository"
	"github.com/yogayulanda/transaction-history-service/internal/service"
	"github.com/yogayulanda/transaction-history-service/internal/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

const (
	sqlPassword    = "Local!Passw0rd2026"
	statusTopic    = "sample.transaction.status"
	testIssuer     = "https://issuer.example.test"
	testAudience   = "transaction-history-local"
	destinationRaw = "+6281234567890"
	destinationID  = "081234567890"
	sourceRaw      = "9876543210"
)

type runningRuntime struct {
	core     *coreapp.App
	conn     *grpc.ClientConn
	httpURL  string
	cancel   context.CancelFunc
	done     <-chan error
	stopOnce sync.Once
}

type consumerStartable struct {
	consumer messaging.Consumer
	ctx      context.Context
}

func (c consumerStartable) Name() string { return "integration_status_consumer" }
func (c consumerStartable) Start() error { return c.consumer.Start(c.ctx) }

func sqlPort() string {
	if port := os.Getenv("TH_SQL_PORT"); port != "" {
		return port
	}
	return "14333"
}

func kafkaPort() string {
	if port := os.Getenv("TH_KAFKA_PORT"); port != "" {
		return port
	}
	return "29092"
}

func sqlDSN(database string) string {
	address := net.JoinHostPort("127.0.0.1", sqlPort())
	connectionURL := &url.URL{
		Scheme: "sqlserver",
		User:   url.UserPassword("sa", sqlPassword),
		Host:   address,
	}
	query := url.Values{}
	query.Set("database", database)
	query.Set("encrypt", "disable")
	connectionURL.RawQuery = query.Encode()
	return connectionURL.String()
}

func openSQL(t *testing.T, database string) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlserver", sqlDSN(database))
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		t.Fatalf("connect to local SQL Server: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func recreateDatabase(t *testing.T) {
	t.Helper()
	admin := openSQL(t, "master")
	_, err := admin.Exec(`
IF DB_ID(N'transaction_history') IS NOT NULL
BEGIN
    ALTER DATABASE transaction_history SET SINGLE_USER WITH ROLLBACK IMMEDIATE;
    DROP DATABASE transaction_history;
END;
CREATE DATABASE transaction_history;`)
	if err != nil {
		t.Fatalf("recreate transaction_history database: %v", err)
	}
}

func migrationDir(t *testing.T) string {
	t.Helper()
	dir, err := filepath.Abs(filepath.Join("..", "..", "migrations", "transaction"))
	if err != nil {
		t.Fatal(err)
	}
	return dir
}

func runMigrations(t *testing.T, db *sql.DB, action string) {
	t.Helper()
	if err := coremigration.RunGoose(db, "sqlserver", migrationDir(t), action); err != nil {
		t.Fatalf("goose %s: %v", action, err)
	}
}

func assertSchema(t *testing.T, db *sql.DB) {
	t.Helper()
	tables := []string{"transaction_histories", "transaction_history_details", "transaction_history_status_events", "transaction_error_definitions"}
	for _, table := range tables {
		var found int
		if err := db.QueryRow(`SELECT COUNT(*) FROM sys.tables WHERE schema_id = SCHEMA_ID('dbo') AND name = @p1`, table).Scan(&found); err != nil || found != 1 {
			t.Fatalf("table %s exists=%d err=%v", table, found, err)
		}
	}
	constraints := []string{
		"PK_transaction_histories", "CK_th_direction", "CK_th_status", "CK_th_destination_no_type", "CK_th_amounts",
		"PK_transaction_history_details", "FK_thd_transaction_history", "CK_thd_detail_json",
		"PK_transaction_history_status_events", "FK_thse_transaction_history", "CK_thse_to_status",
		"CK_thse_from_status", "CK_thse_changed_by_type",
	}
	for _, constraint := range constraints {
		var found int
		if err := db.QueryRow(`SELECT COUNT(*) FROM sys.objects WHERE name = @p1 AND type IN ('C','F','PK','UQ')`, constraint).Scan(&found); err != nil || found != 1 {
			t.Fatalf("constraint %s exists=%d err=%v", constraint, found, err)
		}
	}
	indexes := []string{
		"UX_th_transaction_id", "IX_th_user_list", "IX_th_reference_no", "IX_th_destination_no_hash", "IX_th_open",
		"UX_thse_event_id", "UX_thse_status_version", "IX_thse_history_time",
	}
	for _, index := range indexes {
		var found int
		if err := db.QueryRow(`SELECT COUNT(*) FROM sys.indexes WHERE name = @p1`, index).Scan(&found); err != nil || found != 1 {
			t.Fatalf("index %s exists=%d err=%v", index, found, err)
		}
	}
}

func waitFor(t *testing.T, description string, check func() (bool, error)) {
	t.Helper()
	deadline := time.Now().Add(25 * time.Second)
	var lastErr error
	for time.Now().Before(deadline) {
		ok, err := check()
		if ok && err == nil {
			return
		}
		lastErr = err
		time.Sleep(150 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s: %v", description, lastErr)
}

func freePort(t *testing.T) int {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	_ = listener.Close()
	return port
}

func runtimeConfig(t *testing.T, environment, publicKey, issuer, audience string, grpcPort, httpPort int) *coreconfig.Config {
	t.Helper()
	for key, value := range map[string]string{
		"SERVICE_NAME":                    "transaction-history-predeploy-test",
		"APP_ENV":                         environment,
		"LOG_LEVEL":                       "error",
		"DB_LIST":                         "transaction_history",
		"DB_TRANSACTION_HISTORY_DRIVER":   "sqlserver",
		"DB_TRANSACTION_HISTORY_HOST":     "127.0.0.1",
		"DB_TRANSACTION_HISTORY_PORT":     sqlPort(),
		"DB_TRANSACTION_HISTORY_NAME":     "transaction_history",
		"DB_TRANSACTION_HISTORY_USER":     "sa",
		"DB_TRANSACTION_HISTORY_PASSWORD": sqlPassword,
		"DB_TRANSACTION_HISTORY_PARAMS":   "encrypt=disable",
		"KAFKA_BROKERS":                   "localhost:" + kafkaPort(),
		"KAFKA_INBOUND_ENABLED":           "false",
		"HISTORY_STATUS_TOPIC":            statusTopic,
		"HISTORY_ALLOWED_SOURCES":         "sample-producer",
		"HISTORY_ALLOWED_APPS":            "app_a,app_b",
		"HISTORY_HASH_KEY":                strings.Repeat("6b", 32),
		"HISTORY_BACKOFFICE_ROLE":         "backoffice",
		"HISTORY_JWT_PUBLIC_KEY":          publicKey,
		"HISTORY_JWT_JWKS_URL":            "",
		"HISTORY_JWT_ISSUER":              issuer,
		"HISTORY_JWT_AUDIENCE":            audience,
		"HTTP_PORT":                       fmt.Sprint(httpPort),
		"GRPC_PORT":                       fmt.Sprint(grpcPort),
		"DIAG_ENABLED":                    "false",
	} {
		t.Setenv(key, value)
	}
	cfg, err := coreconfig.Load()
	if err != nil {
		t.Fatal(err)
	}
	if err := app.ConfigureInternalJWT(cfg); err != nil {
		t.Fatal(err)
	}
	cfg.App.DrainDelay = 0
	cfg.App.ShutdownTimeout = 15 * time.Second
	if err := cfg.Validate(); err != nil {
		t.Fatalf("validate runtime config: %v", err)
	}
	return cfg
}

func newCore(t *testing.T, cfg *coreconfig.Config) (*coreapp.App, context.Context, context.CancelFunc) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	core, err := coreapp.New(ctx, cfg)
	if err != nil {
		cancel()
		t.Fatalf("initialize core app: %v", err)
	}
	return core, ctx, cancel
}

func startRuntime(t *testing.T, environment, publicKey, issuer, audience string, withConsumer bool) *runningRuntime {
	t.Helper()
	grpcPort, httpPort := freePort(t), freePort(t)
	cfg := runtimeConfig(t, environment, publicKey, issuer, audience, grpcPort, httpPort)
	core, ctx, cancel := newCore(t, cfg)
	svcApp, err := app.New(ctx, core)
	if err != nil {
		cancel()
		_ = core.Lifecycle().Shutdown(context.Background())
		t.Fatalf("initialize service app: %v", err)
	}
	components := append([]coreserver.Startable{}, svcApp.Components...)
	if withConsumer {
		publisher, err := core.NewKafkaPublisher()
		if err != nil {
			t.Fatalf("create test event publisher: %v", err)
		}
		repo := repository.NewTransactionRepository(core.SQLByName("transaction_history").DB, core.Logger())
		hashKey, _ := hex.DecodeString(strings.Repeat("6b", 32))
		hasher, err := utils.NewHasher(hashKey)
		if err != nil {
			t.Fatal(err)
		}
		validationCatalog := repository.NewStaticCatalog([]string{"app_a", "app_b"}, []string{"sample-producer"})
		eventService := service.NewStatusEventService(repo, validationCatalog, hasher, core.Logger())
		eventHandler := handlerkafka.NewStatusEventHandler(eventService, publisher, core.Logger())
		consumer, err := core.NewKafkaConsumer(
			statusTopic,
			"predeploy-consumer-"+uuid.NewString(),
			eventHandler.Handle,
			messaging.WithConsumerConcurrency(3),
			messaging.WithConsumerRetry(5, 5*time.Second),
			messaging.WithConsumerDLQ(publisher),
		)
		if err != nil {
			t.Fatalf("create test status consumer: %v", err)
		}
		components = append(components, consumerStartable{consumer: consumer, ctx: ctx})
	}
	done := make(chan error, 1)
	go func() { done <- coreserver.Run(ctx, core, components...) }()

	dialCtx, dialCancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer dialCancel()
	conn, err := grpc.DialContext(
		dialCtx,
		net.JoinHostPort("127.0.0.1", fmt.Sprint(grpcPort)),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		cancel()
		_ = core.Lifecycle().Shutdown(context.Background())
		t.Fatalf("connect to in-process gRPC server: %v", err)
	}
	runtime := &runningRuntime{
		core:    core,
		conn:    conn,
		httpURL: fmt.Sprintf("http://127.0.0.1:%d", httpPort),
		cancel:  cancel,
		done:    done,
	}
	t.Cleanup(runtime.stop)
	return runtime
}

func (r *runningRuntime) stop() {
	if r == nil {
		return
	}
	r.stopOnce.Do(func() {
		if r.conn != nil {
			_ = r.conn.Close()
		}
		r.cancel()
		select {
		case <-r.done:
		case <-time.After(20 * time.Second):
			if r.core != nil {
				_ = r.core.Lifecycle().Shutdown(context.Background())
			}
		}
	})
}

func outgoingClaims(subject, appName, role string) context.Context {
	pairs := []string{"x-subject", subject, "x-claim-app_name", appName}
	if role != "" {
		pairs = append(pairs, "x-role", role)
	}
	return metadata.NewOutgoingContext(context.Background(), metadata.Pairs(pairs...))
}

func outgoingJWT(token string) context.Context {
	return metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", "Bearer "+token))
}

func issueToken(t *testing.T, privateKey *rsa.PrivateKey, issuer, audience, subject, appName, role string, expires time.Time) string {
	t.Helper()
	header, err := json.Marshal(map[string]string{"alg": "RS256", "typ": "JWT"})
	if err != nil {
		t.Fatal(err)
	}
	claims, err := json.Marshal(map[string]any{
		"iss":        issuer,
		"aud":        audience,
		"sub":        subject,
		"role":       role,
		"attributes": map[string]string{"app_name": appName},
		"iat":        time.Now().Unix(),
		"exp":        expires.Unix(),
	})
	if err != nil {
		t.Fatal(err)
	}
	unsigned := base64.RawURLEncoding.EncodeToString(header) + "." + base64.RawURLEncoding.EncodeToString(claims)
	digest := sha256.Sum256([]byte(unsigned))
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, digest[:])
	if err != nil {
		t.Fatal(err)
	}
	return unsigned + "." + base64.RawURLEncoding.EncodeToString(signature)
}

func rsaTestKey(t *testing.T) (*rsa.PrivateKey, string) {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	publicKeyDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: publicKeyDER})
	return privateKey, string(publicKeyPEM)
}

func newKafkaWriter(t *testing.T) *kafka.Writer {
	t.Helper()
	writer := &kafka.Writer{
		Addr:                   kafka.TCP("localhost:" + kafkaPort()),
		RequiredAcks:           kafka.RequireAll,
		AllowAutoTopicCreation: true,
	}
	t.Cleanup(func() { _ = writer.Close() })
	return writer
}

func ensureKafkaTopics(t *testing.T, topics ...string) {
	t.Helper()
	conn, err := kafka.Dial("tcp", "localhost:"+kafkaPort())
	if err != nil {
		t.Fatalf("connect to local Kafka for topic setup: %v", err)
	}
	defer conn.Close()
	configs := make([]kafka.TopicConfig, 0, len(topics))
	for _, topic := range topics {
		configs = append(configs, kafka.TopicConfig{Topic: topic, NumPartitions: 3, ReplicationFactor: 1})
	}
	if err := conn.CreateTopics(configs...); err != nil {
		t.Fatalf("create integration Kafka topics: %v", err)
	}
}

func publishStatusEvent(t *testing.T, writer *kafka.Writer, topic, transactionID, eventID string, version int64, body map[string]any) {
	t.Helper()
	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	message := kafka.Message{
		Key:   []byte(transactionID),
		Value: payload,
		Headers: []kafka.Header{
			{Key: "ce_specversion", Value: []byte("1.0")},
			{Key: "ce_id", Value: []byte(eventID)},
			{Key: "ce_type", Value: []byte("transaction.status.v1")},
			{Key: "ce_source", Value: []byte("sample-producer")},
			{Key: "ce_time", Value: []byte(time.Now().UTC().Format(time.RFC3339Nano))},
			{Key: "ce_subject", Value: []byte(transactionID)},
			{Key: "ce_sequence", Value: []byte(fmt.Sprint(version))},
		},
	}
	message.Topic = topic
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := writer.WriteMessages(ctx, message); err != nil {
		t.Fatalf("publish status event %s: %v", transactionID, err)
	}
}

func eventBody(transactionID, userID, productType, productCategory string, detail map[string]any) map[string]any {
	return map[string]any{
		"transaction_id":      transactionID,
		"app_name":            "app_a",
		"user_id":             userID,
		"product_category":    productCategory,
		"product_type":        productType,
		"provider":            "bca",
		"direction":           "debit",
		"denomination":        "5.00",
		"amount":              "100.00",
		"fee_amount":          "1.00",
		"total_amount":        "101.00",
		"currency":            "IDR",
		"status":              "SUCCESS",
		"response_code":       "00",
		"source_account_no":   sourceRaw,
		"destination_name":    "Sample Recipient",
		"destination_no":      destinationRaw,
		"destination_no_type": "phone",
		"reference_no":        "SAMPLE-REF-1",
		"source_system":       "sample-producer",
		"transaction_at":      time.Now().UTC().Format(time.RFC3339Nano),
		"changed_by_type":     "system",
		"detail":              detail,
	}
}

func transactionCounts(db *sql.DB, transactionID string) (headers, details, events int, err error) {
	err = db.QueryRow(`SELECT COUNT(*) FROM dbo.transaction_histories WHERE transaction_id = @p1`, transactionID).Scan(&headers)
	if err != nil {
		return
	}
	err = db.QueryRow(`SELECT COUNT(*) FROM dbo.transaction_history_details d JOIN dbo.transaction_histories h ON h.id = d.transaction_history_id WHERE h.transaction_id = @p1`, transactionID).Scan(&details)
	if err != nil {
		return
	}
	err = db.QueryRow(`SELECT COUNT(*) FROM dbo.transaction_history_status_events e JOIN dbo.transaction_histories h ON h.id = e.transaction_history_id WHERE h.transaction_id = @p1`, transactionID).Scan(&events)
	return
}

func waitForTransaction(t *testing.T, db *sql.DB, transactionID string, wantHeaders, wantDetails, wantEvents int) {
	t.Helper()
	waitFor(t, "transaction rows for "+transactionID, func() (bool, error) {
		headers, details, events, err := transactionCounts(db, transactionID)
		return headers == wantHeaders && details == wantDetails && events == wantEvents, err
	})
}

func assertNoRawNumbers(t *testing.T, db *sql.DB, transactionID string) {
	t.Helper()
	queries := []struct {
		query string
		args  []any
	}{
		{`SELECT * FROM dbo.transaction_histories WHERE transaction_id = @p1`, []any{transactionID}},
		{`SELECT d.* FROM dbo.transaction_history_details d JOIN dbo.transaction_histories h ON h.id = d.transaction_history_id WHERE h.transaction_id = @p1`, []any{transactionID}},
		{`SELECT e.* FROM dbo.transaction_history_status_events e JOIN dbo.transaction_histories h ON h.id = e.transaction_history_id WHERE h.transaction_id = @p1`, []any{transactionID}},
	}
	for _, item := range queries {
		rows, err := db.Query(item.query, item.args...)
		if err != nil {
			t.Fatal(err)
		}
		columns, err := rows.Columns()
		if err != nil {
			_ = rows.Close()
			t.Fatal(err)
		}
		for rows.Next() {
			values := make([]any, len(columns))
			pointers := make([]any, len(columns))
			for i := range values {
				pointers[i] = &values[i]
			}
			if err := rows.Scan(pointers...); err != nil {
				_ = rows.Close()
				t.Fatal(err)
			}
			for _, value := range values {
				text := fmt.Sprint(value)
				if strings.Contains(text, destinationRaw) || strings.Contains(text, destinationID) || strings.Contains(text, sourceRaw) {
					_ = rows.Close()
					t.Fatalf("raw account/phone value found in persisted row")
				}
			}
		}
		if err := rows.Err(); err != nil {
			_ = rows.Close()
			t.Fatal(err)
		}
		_ = rows.Close()
	}
}

func requireCoreError(t *testing.T, err error, code string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected %s, got nil", code)
	}
	info, ok := coreErrors.InfoFromGRPC(err)
	if !ok || info.Code != code {
		t.Fatalf("error info = %+v present=%v; want %s (error %v)", info, ok, code, err)
	}
}

func queryDB(t *testing.T, db *sql.DB, query string, args ...any) *sql.Rows {
	t.Helper()
	rows, err := db.Query(query, args...)
	if err != nil {
		t.Fatal(err)
	}
	return rows
}

func minDuration(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}
