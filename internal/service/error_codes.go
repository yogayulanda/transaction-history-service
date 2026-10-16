package service

import (
	coreerrors "github.com/yogayulanda/go-core/errors"
	"google.golang.org/grpc/codes"
)

const (
	CodeDuplicateReferenceID = "TRH-VAL-002"
	CodeTransactionNotFound  = "TRH-DB-001"
	CodeInternal             = "TRH-REC-001"
)

func RegisterErrorCodes() error {
	return coreerrors.Register(
		coreerrors.Def{
			Code:           CodeDuplicateReferenceID,
			Class:          coreerrors.ClassFailed,
			HTTPStatus:     400,
			GRPCCode:       codes.InvalidArgument,
			Finality:       coreerrors.FinalityBusiness,
			Retryable:      false,
			DefaultMessage: "Transaksi dengan referensi ini sudah pernah diproses.",
		},
		coreerrors.Def{
			Code:           CodeTransactionNotFound,
			Class:          coreerrors.ClassFailed,
			HTTPStatus:     404,
			GRPCCode:       codes.NotFound,
			Finality:       coreerrors.FinalityBusiness,
			Retryable:      false,
			DefaultMessage: "Riwayat transaksi tidak ditemukan.",
		},
		coreerrors.Def{
			Code:           CodeInternal,
			Class:          coreerrors.ClassError,
			HTTPStatus:     500,
			GRPCCode:       codes.Internal,
			Finality:       coreerrors.FinalityTechnicalNonRecoverable,
			Retryable:      false,
			DefaultMessage: "Sistem sedang sibuk. Silakan coba beberapa saat lagi.",
		},
	)
}
