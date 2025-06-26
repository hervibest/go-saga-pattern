package converter

import (
	"go-saga-pattern/transaction-svc/internal/entity"
	"go-saga-pattern/transaction-svc/internal/model"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/midtrans/midtrans-go/coreapi"
)

func TransactionToCreateResponse(transaction *entity.Transaction, redirectUrl string) *model.CreateTransactionResponse {
	return &model.CreateTransactionResponse{
		TransactionId: transaction.ID.String(),
		SnapToken:     transaction.SnapToken.String,
		RedirectURL:   redirectUrl,
	}
}

func TransactionToResponse(transaction *entity.Transaction) *model.TransactionResponse {
	return &model.TransactionResponse{
		ID:                transaction.ID.String(),
		UserID:            transaction.UserID.String(),
		TotalPrice:        transaction.TotalPrice,
		TransactionStatus: transaction.TransactionStatus,
		CheckoutAt:        transaction.CheckoutAt.Format(time.RFC1123),
		PaymentAt:         transaction.PaymentAt.Time.Format(time.RFC1123),
		UpdatedAt:         transaction.UpdatedAt.Format(time.RFC1123),
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

func SchedulerReqToCheckAndUpdate(schedulerReq *coreapi.TransactionStatusResponse, body []byte) *model.CheckAndUpdateTransactionRequest {
	return &model.CheckAndUpdateTransactionRequest{
		MidtransTransactionStatus: schedulerReq.TransactionStatus,
		StatusCode:                schedulerReq.StatusCode,
		SignatureKey:              schedulerReq.SignatureKey,
		SettlementTime:            schedulerReq.SettlementTime,
		// ReferenceID:               schedulerReq.ReferenceID,
		OrderID: schedulerReq.OrderID,
		// Metadata:                  schedulerReq.Metadata,
		GrossAmount: schedulerReq.GrossAmount,
		Body:        body,
	}
}

func TransactionsWithTotalToResponses(productsWithTotal []*entity.TransactionWithTotal) []*model.TransactionResponse {
	responses := make([]*model.TransactionResponse, 0, len(productsWithTotal))
	for _, transationWithTotal := range productsWithTotal {
		transaction := &entity.Transaction{
			ID:                transationWithTotal.ID,
			UserID:            transationWithTotal.UserID,
			TotalPrice:        transationWithTotal.TotalPrice,
			TransactionStatus: transationWithTotal.TransactionStatus,
			CheckoutAt:        transationWithTotal.CheckoutAt,
			PaymentAt:         transationWithTotal.PaymentAt,
			UpdatedAt:         transationWithTotal.UpdatedAt,
		}

		responses = append(responses, TransactionToResponse(transaction))
	}
	return responses
}

func formatTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC1123)
}

func TransactionWithDetailAndTotalToResponse(transactionDetailTotal []*entity.TransactionWithDetailAndTotal, isOwner bool) []*model.TransactionResponse {
	transactionMap := make(map[string]*model.TransactionResponse)

	for _, row := range transactionDetailTotal {
		txID := row.TransactionID.String()
		//Owner dont have to know user's main transaction data
		if isOwner {
			row.TransactionTotalPrice = 0
		}

		// Jika transaksi belum ada di map, buatkan
		if _, exists := transactionMap[txID]; !exists {
			transactionMap[txID] = &model.TransactionResponse{
				ID:                 txID,
				UserID:             row.TransactionUserID.String(),
				TotalPrice:         row.TransactionTotalPrice,
				TransactionStatus:  row.TransactionStatus,
				CheckoutAt:         formatTime(row.TransactionCheckoutAt),
				PaymentAt:          formatTime(row.TransactionPaymentAt),
				UpdatedAt:          formatTime(row.TransactionUpdatedAt),
				TransactionDetails: make([]*model.TransactionDetailResponse, 0),
			}
		}

		// Tambahkan detail jika ada detail yang valid
		if row.TransactionDetailID != uuid.Nil {
			transactionMap[txID].TransactionDetails = append(transactionMap[txID].TransactionDetails, &model.TransactionDetailResponse{
				ID:        row.TransactionDetailID.String(),
				ProductID: row.TransactionDetailProductID.String(),
				Quantity:  row.TransactionDetailQuantity,
				Price:     row.TransactionDetailPrice,
				CreatedAt: formatTime(row.TransactionDetailCreatedAt),
			})
		}
	}

	// Ubah map menjadi slice
	result := make([]*model.TransactionResponse, 0, len(transactionMap))
	for _, tx := range transactionMap {
		result = append(result, tx)
	}

	return result
}
