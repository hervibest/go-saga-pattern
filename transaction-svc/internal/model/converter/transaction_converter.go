package converter

import (
	"go-saga-pattern/transaction-svc/internal/entity"
	"go-saga-pattern/transaction-svc/internal/model"
	"log"
	"time"
)

func TransactionToResponse(transaction *entity.Transaction, redirectUrl string) *model.CreateTransactionResponse {
	return &model.CreateTransactionResponse{
		TransactionId: transaction.ID.String(),
		SnapToken:     transaction.SnapToken.String,
		RedirectURL:   redirectUrl,
	}
}

// // TODO ISSUE
// func TransactionToResponse(transaction *entity.Transaction, transactionDetails []*entity.TransactionDetail) *model.TransactionResponse {
// 	transactionDetailResponses := TransactionDetailToResponses(transactionDetails)
// 	return &model.TransactionResponse{
// 		ID:                 transaction.ID.String(),
// 		UserID:             transaction.UserID.String(),
// 		TotalPrice:         transaction.TotalPrice,
// 		TransactionStatus:  transaction.TransactionStatus,
// 		CheckoutAt:         transaction.CheckoutAt.Local().Format(time.RFC1123),
// 		PaymentAt:          nullable.SQLtoTime(transaction.PaymentAt),
// 		UpdatedAt:          transaction.UpdatedAt.Local().Format(time.RFC1123),
// 		TransactionDetails: transactionDetailResponses,
// 	}
// }

func TransactionDetailToResponses(transactionDetails []*entity.TransactionDetail) []*model.TransactionDetailResponse {
	if len(transactionDetails) == 0 {
		log.Default().Print("Transaction details are empty, returning nil")
		return nil
	}

	transactionDetailResponses := make([]*model.TransactionDetailResponse, 0, len(transactionDetails))
	for _, transactionDetail := range transactionDetails {
		log.Print("Converting transaction detail to response", "ID", transactionDetail.ID.String(), "ProductID", transactionDetail.ProductID.String(), "Quantity", transactionDetail.Quantity, "Price", transactionDetail.Price)
		transactionDetailResponses = append(transactionDetailResponses, &model.TransactionDetailResponse{
			ID:        transactionDetail.ID.String(),
			ProductID: transactionDetail.ProductID.String(),
			Quantity:  transactionDetail.Quantity,
			Price:     transactionDetail.Price,
			CreatedAt: transactionDetail.CreatedAt.Local().Format(time.RFC1123),
		})
	}

	return transactionDetailResponses
}
