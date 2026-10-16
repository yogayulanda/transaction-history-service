package grpc

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/yogayulanda/go-core/security"
	historyv1 "github.com/yogayulanda/transaction-history-service/gen/go/history/v1"
	"github.com/yogayulanda/transaction-history-service/internal/domain"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestGatewayUsesGRPCInterceptorsAndDoesNotExposeBackoffice(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	server := grpc.NewServer(grpc.UnaryInterceptor(func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		incoming, _ := metadata.FromIncomingContext(ctx)
		if values := incoming.Get("authorization"); len(values) == 0 || values[0] != "Bearer test-token" {
			return nil, status.Error(codes.Unauthenticated, "authentication required")
		}
		ctx = security.Inject(ctx, &security.Claims{Subject: "user-1", Attributes: map[string]string{"app_name": "app_a"}})
		_ = grpc.SetHeader(ctx, metadata.Pairs("x-request-id", "gateway-request-1"))
		return handler(ctx, req)
	}))
	historyv1.RegisterHistoryServiceServer(server, newTestHistoryHandler(&handlerRepository{
		item: domain.TransactionSnapshot{ID: "tx-1", Status: "SUCCESS"},
	}))
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(server.Stop)

	ctx, cancel := context.WithCancel(context.Background())
	mux := runtime.NewServeMux(runtime.WithOutgoingHeaderMatcher(func(key string) (string, bool) {
		if strings.EqualFold(key, "x-request-id") {
			return "X-Request-Id", true
		}
		return runtime.DefaultHeaderMatcher(key)
	}))
	if err := historyv1.RegisterHistoryServiceHandlerFromEndpoint(
		ctx,
		mux,
		listener.Addr().String(),
		[]grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())},
	); err != nil {
		cancel()
		t.Fatal(err)
	}
	defer cancel()

	unauthorized := httptest.NewRecorder()
	mux.ServeHTTP(unauthorized, httptest.NewRequest(http.MethodGet, "/v1/transactions", nil))
	if unauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("unauthorized gateway response = %d, want 401", unauthorized.Code)
	}

	authorizedRequest := httptest.NewRequest(http.MethodGet, "/v1/transactions", nil)
	authorizedRequest.Header.Set("Authorization", "Bearer test-token")
	authorized := httptest.NewRecorder()
	mux.ServeHTTP(authorized, authorizedRequest)
	if authorized.Code != http.StatusOK || authorized.Header().Get("X-Request-Id") != "gateway-request-1" {
		t.Fatalf("authorized gateway response = %d, request-id=%q, body=%s", authorized.Code, authorized.Header().Get("X-Request-Id"), authorized.Body.String())
	}

	backoffice := httptest.NewRecorder()
	mux.ServeHTTP(backoffice, httptest.NewRequest(http.MethodPost, "/history.v1.HistoryBackofficeService/SearchTransactions", strings.NewReader(`{}`)))
	if backoffice.Code != http.StatusNotFound {
		t.Fatalf("backoffice route returned %d, want 404", backoffice.Code)
	}
}
