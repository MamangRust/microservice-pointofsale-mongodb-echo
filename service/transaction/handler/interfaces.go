package handler

import (
	pbtransaction "github.com/MamangRust/microservice-point-of-sale-pb/transaction"
)

type TransactionQueryHandleGrpc interface {
	pbtransaction.TransactionServiceServer
}
