//go:build integration

package integration

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"github.com/yogayulanda/go-core/logger"
	historyv1 "github.com/yogayulanda/transaction-history-service/gen/go/history/v1"
	"github.com/yogayulanda/transaction-history-service/internal/app"
	"github.com/yogayulanda/transaction-history-service/internal/domain"
	"github.com/yogayulanda/transaction-history-service/internal/repository"
	"github.com/yogayulanda/transaction-history-service/internal/utils"
)

func TestPredeployGate(t *testing.T) {
	recreateDatabase(t)
	db := openSQL(t, "transaction_history")
	if !t.Run("a_migrations", func(t *testing.T) { testMigrations(t, db) }) {
		return
	}
	ensureKafkaTopics(t, statusTopic, statusTopic+".dlq")

	local := startRuntime(t, "local", "", "", "", true)
	if !t.Run("b_consumer_products_and_pii", func(t *testing.T) { testConsumerProducts(t, local, db) }) {
		return
	}
	if !t.Run("c_ordering_and_deduplication", func(t *testing.T) { testOrdering(t, local, db) }) {
		return
	}
	if !t.Run("d_concurrent_first_events", func(t *testing.T) { testConcurrency(t, local, db) }) {
		return
	}
	if !t.Run("e_invalid_event_dlq", func(t *testing.T) { testDLQ(t, local, db) }) {
		return
	}
	if !t.Run("f_grpc_gateway_and_backoffice", func(t *testing.T) { testAPI(t, local) }) {
		return
	}
	local.stop()
	t.Run("g_nonlocal_jwt_startup_and_auth", func(t *testing.T) { testNonlocalJWT(t, db) })
}

func testMigrations(t *testing.T, db *sql.DB) {
	runMigrations(t, db, "up")
	assertSchema(t, db)
	for range 5 {
		runMigrations(t, db, "down")
	}
	var applied int
	if err := db.QueryRow(`SELECT COUNT(*) FROM dbo.goose_db_version WHERE is_applied = 1 AND version_id > 0`).Scan(&applied); err != nil {
		t.Fatal(err)
	}
	if applied != 0 {
		t.Fatalf("applied goose migrations after down-to-zero = %d, want 0", applied)
	}
	for _, table := range []string{"transaction_histories", "transaction_history_details", "transaction_history_status_events", "transaction_error_definitions"} {
		var exists int
		if err := db.QueryRow(`SELECT COUNT(*) FROM sys.tables WHERE schema_id = SCHEMA_ID('dbo') AND name = @p1`, table).Scan(&exists); err != nil || exists != 0 {
			t.Fatalf("table %s remains after down-to-zero: count=%d err=%v", table, exists, err)
		}
	}
	runMigrations(t, db, "up")
	assertSchema(t, db)
}

func testConsumerProducts(t *testing.T, runtime *runningRuntime, db *sql.DB) {
	writer := newKafkaWriter(t)
	products := []struct {
		typeName string
		category string
		detail   map[string]any
	}{
		{domain.ProductTransferBiFast, domain.CategoryTransfer, map[string]any{"bank_code": "014", "note": "sample transfer note"}},
		{domain.ProductPulsa, domain.CategoryPPOB, map[string]any{"serial_number": "SAMPLE-SERIAL-1"}},
		{domain.ProductDataPackage, domain.CategoryPPOB, map[string]any{"product_name": "Sample data", "serial_number": "SAMPLE-SERIAL-2"}},
		{domain.ProductPLNPrepaid, domain.CategoryPPOB, map[string]any{"token": "SAMPLE-PLN-TOKEN", "kwh": "2.5", "tariff_power": "R1-1300"}},
		{domain.ProductPLNPostpaid, domain.CategoryPPOB, map[string]any{"period": "2026-10", "stand_meter": "10-20"}},
		{domain.ProductBPJSKesehatan, domain.CategoryPPOB, map[string]any{"period_months": 1, "participant_count": 2, "branch": "Sample City"}},
		{domain.ProductCashIn, domain.CategoryCICO, map[string]any{"code": "SAMPLE-CICO-123456", "warkat_no": "SAMPLE-WR-1"}},
		{domain.ProductCashOut, domain.CategoryCICO, map[string]any{"code": "SAMPLE-CICO-654321", "warkat_no": "SAMPLE-WR-2"}},
	}
	db = runtime.core.SQLByName("transaction_history").DB
	hashKey := mustHashKey(t)
	hasher, err := utils.NewHasher(hashKey)
	if err != nil {
		t.Fatal(err)
	}
	for _, product := range products {
		transactionID := "product-" + product.typeName
		body := eventBody(transactionID, "owner-1", product.typeName, product.category, product.detail)
		publishStatusEvent(t, writer, statusTopic, transactionID, uuid.NewString(), 1, body)
		waitForTransaction(t, db, transactionID, 1, 1, 1)

		var masked, destinationHash, sourceMasked, status string
		var statusVersion int64
		if err := db.QueryRow(`SELECT destination_no_masked, destination_no_hash, source_account_masked, status, status_version FROM dbo.transaction_histories WHERE transaction_id = @p1`, transactionID).
			Scan(&masked, &destinationHash, &sourceMasked, &status, &statusVersion); err != nil {
			t.Fatal(err)
		}
		if masked != "0812****7890" || destinationHash != hasher.Hash("phone", destinationID) || sourceMasked != "****3210" || status != domain.StatusSuccess || statusVersion != 1 {
			t.Fatalf("unexpected persisted header for %s: masked=%q source=%q status=%s version=%d", product.typeName, masked, sourceMasked, status, statusVersion)
		}
		var detailJSON string
		if err := db.QueryRow(`SELECT detail_json FROM dbo.transaction_history_details d JOIN dbo.transaction_histories h ON h.id = d.transaction_history_id WHERE h.transaction_id = @p1`, transactionID).Scan(&detailJSON); err != nil {
			t.Fatal(err)
		}
		var persistedDetail map[string]any
		if err := json.Unmarshal([]byte(detailJSON), &persistedDetail); err != nil {
			t.Fatal(err)
		}
		if len(persistedDetail) == 0 {
			t.Fatalf("no normalized detail persisted for %s", product.typeName)
		}
		if product.typeName == domain.ProductCashIn || product.typeName == domain.ProductCashOut {
			if persistedDetail["code_masked"] != "****456" && persistedDetail["code_masked"] != "****321" {
				t.Fatalf("CICO code was not masked: %+v", persistedDetail)
			}
			if _, rawPresent := persistedDetail["code"]; rawPresent {
				t.Fatal("raw CICO code persisted")
			}
		}
		var eventVersion int64
		if err := db.QueryRow(`SELECT e.status_version FROM dbo.transaction_history_status_events e JOIN dbo.transaction_histories h ON h.id = e.transaction_history_id WHERE h.transaction_id = @p1`, transactionID).Scan(&eventVersion); err != nil || eventVersion != 1 {
			t.Fatalf("status event version for %s = %d, err=%v", product.typeName, eventVersion, err)
		}
		assertNoRawNumbers(t, db, transactionID)
	}
}

func mustHashKey(t *testing.T) []byte {
	t.Helper()
	key, err := hex.DecodeString(strings.Repeat("6b", 32))
	if err != nil {
		t.Fatal(err)
	}
	return key
}

func testOrdering(t *testing.T, runtime *runningRuntime, db *sql.DB) {
	db = runtime.core.SQLByName("transaction_history").DB
	log := &capturedLogger{}
	repo := repository.NewTransactionRepository(db, log)
	transactionID := "ordering-sample"
	transactionAt := time.Now().UTC()
	apply := func(eventID string, version int64, status string) domain.ApplyResult {
		t.Helper()
		transaction := &domain.Transaction{
			TransactionID: transactionID, AppName: "app_a", UserID: "owner-1", ProductCategory: domain.CategoryTransfer,
			ProductType: domain.ProductTransferInternal, Provider: "bca", Direction: "debit", Amount: "100.00",
			FeeAmount: "0.00", TotalAmount: "100.00", Currency: "IDR", Status: status, StatusVersion: version,
			SourceSystem: "sample-producer", TransactionAt: transactionAt,
		}
		result, err := repo.ApplyStatusEvent(context.Background(), transaction, nil, domain.StatusEvent{
			EventID: eventID, EventType: "transaction.status.v1", StatusVersion: version,
			ToStatus: status, ChangedByType: domain.ChangedBySystem, ChangedAt: transactionAt,
		})
		if err != nil {
			t.Fatalf("ApplyStatusEvent(%s, %d) error: %v", eventID, version, err)
		}
		return result
	}
	if result := apply(uuid.NewString(), 2, domain.StatusPending); result != domain.Created {
		t.Fatalf("version 2 result = %s, want Created", result)
	}
	firstEventID := uuid.NewString()
	if result := apply(firstEventID, 1, domain.StatusFailed); result != domain.Stale {
		t.Fatalf("version 1 result = %s, want Stale", result)
	}
	if result := apply(firstEventID, 1, domain.StatusFailed); result != domain.Duplicate {
		t.Fatalf("duplicate event result = %s, want Duplicate", result)
	}
	if result := apply(uuid.NewString(), 2, domain.StatusSuccess); result != domain.Duplicate {
		t.Fatalf("different event with used version result = %s, want Duplicate", result)
	}
	log.mu.Lock()
	conflicts := 0
	for _, entry := range log.serviceLogs {
		if entry.ErrorCode == "version_conflict" {
			conflicts++
		}
	}
	log.mu.Unlock()
	if conflicts != 1 {
		t.Fatalf("version_conflict count = %d, want 1", conflicts)
	}
	var status string
	var version int64
	if err := db.QueryRow(`SELECT status, status_version FROM dbo.transaction_histories WHERE transaction_id = @p1`, transactionID).Scan(&status, &version); err != nil {
		t.Fatal(err)
	}
	if status != domain.StatusPending || version != 2 {
		t.Fatalf("header after ordering matrix = %s/%d, want PENDING/2", status, version)
	}
	headers, _, events, err := transactionCounts(db, transactionID)
	if err != nil {
		t.Fatal(err)
	}
	if headers != 1 || events != 2 {
		t.Fatalf("rows after ordering matrix = headers:%d events:%d, want 1 and 2", headers, events)
	}
}

func testConcurrency(t *testing.T, runtime *runningRuntime, db *sql.DB) {
	db = runtime.core.SQLByName("transaction_history").DB
	repo := repository.NewTransactionRepository(db, nil)
	const workers = 20
	transactionID := "concurrent-first-event"
	start := make(chan struct{})
	errs := make(chan error, workers)
	var wait sync.WaitGroup
	for range workers {
		wait.Add(1)
		go func() {
			defer wait.Done()
			<-start
			when := time.Now().UTC()
			transaction := &domain.Transaction{
				TransactionID: transactionID, AppName: "app_a", UserID: "owner-1", ProductCategory: domain.CategoryTransfer,
				ProductType: domain.ProductTransferInternal, Provider: "bca", Direction: "debit", Amount: "1.00",
				FeeAmount: "0.00", TotalAmount: "1.00", Currency: "IDR", Status: domain.StatusPending, StatusVersion: 1,
				SourceSystem: "sample-producer", TransactionAt: when,
			}
			_, err := repo.ApplyStatusEvent(context.Background(), transaction, nil, domain.StatusEvent{
				EventID: uuid.NewString(), EventType: "transaction.status.v1", StatusVersion: 1,
				ToStatus: domain.StatusPending, ChangedByType: domain.ChangedBySystem, ChangedAt: when,
			})
			errs <- err
		}()
	}
	close(start)
	wait.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Errorf("concurrent ApplyStatusEvent error: %v", err)
		}
	}
	var headers int
	if err := db.QueryRow(`SELECT COUNT(*) FROM dbo.transaction_histories WHERE transaction_id = @p1`, transactionID).Scan(&headers); err != nil {
		t.Fatal(err)
	}
	if headers != 1 {
		t.Fatalf("concurrent header count = %d, want 1", headers)
	}
}

func testDLQ(t *testing.T, runtime *runningRuntime, db *sql.DB) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{"localhost:" + kafkaPort()},
		Topic:       statusTopic + ".dlq",
		GroupID:     "predeploy-dlq-reader-" + uuid.NewString(),
		MinBytes:    1,
		MaxBytes:    10e6,
		StartOffset: kafka.FirstOffset,
	})
	defer reader.Close()
	dlqResult := make(chan kafka.Message, 1)
	dlqError := make(chan error, 1)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	go func() {
		message, err := reader.ReadMessage(ctx)
		if err != nil {
			dlqError <- err
			return
		}
		dlqResult <- message
	}()
	writer := newKafkaWriter(t)
	transactionID := "invalid-product-sample"
	body := eventBody(transactionID, "owner-1", "unknown_product", domain.CategoryPPOB, map[string]any{})
	publishStatusEvent(t, writer, statusTopic, transactionID, uuid.NewString(), 1, body)
	select {
	case message := <-dlqResult:
		reason := ""
		for _, header := range message.Headers {
			if header.Key == "x-dlq-reason" {
				reason = string(header.Value)
			}
		}
		if message.Topic != statusTopic+".dlq" || string(message.Key) != transactionID || reason != "invalid_product" {
			t.Fatalf("unexpected DLQ message: topic=%s key=%s reason=%s", message.Topic, message.Key, reason)
		}
	case err := <-dlqError:
		t.Fatalf("read DLQ message: %v", err)
	case <-ctx.Done():
		t.Fatal("timed out waiting for invalid event in DLQ")
	}
	db = runtime.core.SQLByName("transaction_history").DB
	headers, details, events, err := transactionCounts(db, transactionID)
	if err != nil {
		t.Fatal(err)
	}
	if headers != 0 || details != 0 || events != 0 {
		t.Fatalf("invalid event wrote rows: headers=%d details=%d events=%d", headers, details, events)
	}
}

func testAPI(t *testing.T, runtime *runningRuntime) {
	client := historyv1.NewHistoryServiceClient(runtime.conn)
	ownerCtx := outgoingClaims("owner-1", "app_a", "")
	list, err := client.ListTransactions(ownerCtx, &historyv1.ListTransactionsRequest{PageSize: 50})
	if err != nil {
		t.Fatalf("owner list: %v", err)
	}
	if len(list.Items) < 8 {
		t.Fatalf("owner list returned %d items, want at least 8", len(list.Items))
	}
	transactionID := "product-transfer_bifast"
	detail, err := client.GetTransaction(ownerCtx, &historyv1.GetTransactionRequest{TransactionId: transactionID})
	if err != nil || detail.GetTransaction().GetTransactionId() != transactionID {
		t.Fatalf("owner detail response=%+v err=%v", detail, err)
	}
	_, err = client.GetTransaction(outgoingClaims("another-owner", "app_a", ""), &historyv1.GetTransactionRequest{TransactionId: transactionID})
	requireCoreError(t, err, "CORE-1004")
	_, err = client.GetTransaction(context.Background(), &historyv1.GetTransactionRequest{TransactionId: transactionID})
	requireCoreError(t, err, "CORE-1001")

	backofficeClient := historyv1.NewHistoryBackofficeServiceClient(runtime.conn)
	backofficeCtx := outgoingClaims("support-user", "app_a", "backoffice")
	backofficeResponse, err := backofficeClient.SearchTransactions(backofficeCtx, &historyv1.SearchTransactionsRequest{TransactionId: transactionID})
	if err != nil || len(backofficeResponse.GetItems()) != 1 {
		t.Fatalf("backoffice search items=%d err=%v", len(backofficeResponse.GetItems()), err)
	}

	for _, path := range []string{"/v1/transactions", "/v1/transactions/" + transactionID} {
		request, err := http.NewRequest(http.MethodGet, runtime.httpURL+path, nil)
		if err != nil {
			t.Fatal(err)
		}
		request.Header.Set("X-Request-ID", "predeploy-http-request")
		request.Header.Set("X-Subject", "owner-1")
		request.Header.Set("X-Claim-App_Name", "app_a")
		httpResponse, err := (&http.Client{Timeout: 10 * time.Second}).Do(request)
		if err != nil {
			t.Fatalf("HTTP gateway %s: %v", path, err)
		}
		_, _ = io.Copy(io.Discard, httpResponse.Body)
		_ = httpResponse.Body.Close()
		if httpResponse.StatusCode != http.StatusUnauthorized {
			t.Fatalf("HTTP gateway %s status=%d, want 401 because local claim metadata is not forwarded", path, httpResponse.StatusCode)
		}
		if httpResponse.Header.Get("X-Request-ID") != "predeploy-http-request" {
			t.Fatalf("HTTP gateway %s X-Request-ID=%q", path, httpResponse.Header.Get("X-Request-ID"))
		}
	}
	request, err := http.NewRequest(http.MethodPost, runtime.httpURL+"/history.v1.HistoryBackofficeService/SearchTransactions", strings.NewReader(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	httpResponse, err := (&http.Client{Timeout: 10 * time.Second}).Do(request)
	if err != nil {
		t.Fatal(err)
	}
	var routeResponse struct {
		ResponseCode string `json:"response_code"`
	}
	if err := json.NewDecoder(httpResponse.Body).Decode(&routeResponse); err != nil {
		_ = httpResponse.Body.Close()
		t.Fatalf("decode unmapped backoffice route response: %v", err)
	}
	_ = httpResponse.Body.Close()
	if httpResponse.StatusCode != http.StatusInternalServerError || routeResponse.ResponseCode != "CORE-9000" {
		t.Fatalf("backoffice HTTP path status=%d response_code=%q; want the generic unmapped-route response", httpResponse.StatusCode, routeResponse.ResponseCode)
	}
}

func testNonlocalJWT(t *testing.T, db *sql.DB) {
	grpcPort, httpPort := freePort(t), freePort(t)
	cfg := runtimeConfig(t, "sit", "", "", "", grpcPort, httpPort)
	if cfg.Auth.InternalJWT.Enabled {
		t.Fatal("InternalJWT unexpectedly enabled with no service JWT environment")
	}
	core, ctx, cancel := newCore(t, cfg)
	_, err := app.New(ctx, core)
	if err == nil || !strings.Contains(err.Error(), "default claims resolver requires internal JWT") {
		cancel()
		_ = core.Lifecycle().Shutdown(context.Background())
		t.Fatalf("non-local startup error = %v, want default ClaimsResolver guard", err)
	}
	cancel()
	if err := core.Lifecycle().Shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown after expected startup rejection: %v", err)
	}

	privateKey, publicKey := rsaTestKey(t)
	runtime := startRuntime(t, "sit", publicKey, testIssuer, testAudience, false)
	defer runtime.stop()
	client := historyv1.NewHistoryServiceClient(runtime.conn)
	transactionID := "product-transfer_bifast"
	valid := issueToken(t, privateKey, testIssuer, testAudience, "owner-1", "app_a", "backoffice", time.Now().Add(time.Hour))
	if _, err := client.GetTransaction(outgoingJWT(valid), &historyv1.GetTransactionRequest{TransactionId: transactionID}); err != nil {
		t.Fatalf("valid JWT owner detail request: %v", err)
	}

	wrongSigner, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	invalidSignature := issueToken(t, wrongSigner, testIssuer, testAudience, "owner-1", "app_a", "backoffice", time.Now().Add(time.Hour))
	_, err = client.GetTransaction(outgoingJWT(invalidSignature), &historyv1.GetTransactionRequest{TransactionId: transactionID})
	requireCoreError(t, err, "CORE-1001")
	wrongIssuer := issueToken(t, privateKey, "https://wrong-issuer.example.test", testAudience, "owner-1", "app_a", "backoffice", time.Now().Add(time.Hour))
	_, err = client.GetTransaction(outgoingJWT(wrongIssuer), &historyv1.GetTransactionRequest{TransactionId: transactionID})
	requireCoreError(t, err, "CORE-1001")
	expired := issueToken(t, privateKey, testIssuer, testAudience, "owner-1", "app_a", "backoffice", time.Now().Add(-time.Hour))
	_, err = client.GetTransaction(outgoingJWT(expired), &historyv1.GetTransactionRequest{TransactionId: transactionID})
	requireCoreError(t, err, "CORE-1001")
	_ = db
}

type capturedLogger struct {
	mu          sync.Mutex
	serviceLogs []logger.ServiceLog
}

func (*capturedLogger) Info(context.Context, string, ...logger.Field)  {}
func (*capturedLogger) Error(context.Context, string, ...logger.Field) {}
func (*capturedLogger) Debug(context.Context, string, ...logger.Field) {}
func (*capturedLogger) Warn(context.Context, string, ...logger.Field)  {}
func (l *capturedLogger) LogService(_ context.Context, entry logger.ServiceLog) {
	l.mu.Lock()
	l.serviceLogs = append(l.serviceLogs, entry)
	l.mu.Unlock()
}
func (*capturedLogger) LogDB(context.Context, logger.DBLog)                   {}
func (*capturedLogger) LogEvent(context.Context, logger.EventLog)             {}
func (*capturedLogger) LogTransaction(context.Context, logger.TransactionLog) {}
func (l *capturedLogger) WithComponent(string) logger.Logger                  { return l }
