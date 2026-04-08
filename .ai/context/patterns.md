Repository pattern

func (r *TransactionRepository) ListByUser(ctx context.Context, userID string) ([]Transaction, error)

Service pattern

func (s *TransactionService) GetUserHistory(ctx context.Context, userID string) ([]Transaction, error)

Handler pattern

func (h *Handler) GetTransactionHistory(ctx context.Context, req *pb.Request) (*pb.Response, error)

Transaction pattern

dbtx.WithTx(ctx, db, func(txCtx context.Context) error {
  // db operations
})