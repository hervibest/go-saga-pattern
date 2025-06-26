package usecase

import (
	"context"
	"crypto/sha512"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"go-saga-pattern/commoner/constant/enum"
	errorcode "go-saga-pattern/commoner/constant/errcode"
	"go-saga-pattern/commoner/constant/message"
	"go-saga-pattern/commoner/helper"
	"go-saga-pattern/commoner/helper/nullable"
	"go-saga-pattern/commoner/logs"
	"go-saga-pattern/commoner/web"
	"go-saga-pattern/transaction-svc/internal/adapter"
	"go-saga-pattern/transaction-svc/internal/entity"
	"go-saga-pattern/transaction-svc/internal/gateway/task"
	"go-saga-pattern/transaction-svc/internal/model"
	"go-saga-pattern/transaction-svc/internal/model/converter"
	"go-saga-pattern/transaction-svc/internal/model/event"
	"go-saga-pattern/transaction-svc/internal/repository"
	"go-saga-pattern/transaction-svc/internal/repository/store"
	"go-saga-pattern/transaction-svc/internal/usecase/contract"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
	"github.com/sony/gobreaker"
	"go.uber.org/zap"
)

type transactionUseCase struct {
	transactionRepo       repository.TransactionRepository
	transactionDetailRepo repository.TransactionDetailRepository
	databaseStore         store.DatabaseStore
	productAdapter        adapter.ProductAdapter
	messagingAdapter      adapter.MessagingAdapter
	paymentAdapter        adapter.PaymentAdapter
	cacheAdapter          adapter.CacheAdapter
	expireTask            task.TransactionTask
	timeParserHelper      helper.TimeParserHelper
	validator             helper.CustomValidator
	log                   logs.Log
}

func NewTransactionUseCase(transactionRepo repository.TransactionRepository, transactionDetailRepo repository.TransactionDetailRepository,
	databaseStore store.DatabaseStore, productAdapter adapter.ProductAdapter, messagingAdapter adapter.MessagingAdapter, paymentAdapter adapter.PaymentAdapter,
	cacheAdapter adapter.CacheAdapter, expireTask task.TransactionTask, timeParserHelper helper.TimeParserHelper, validator helper.CustomValidator,
	log logs.Log) contract.TransactionUseCase {
	return &transactionUseCase{
		transactionRepo:       transactionRepo,
		transactionDetailRepo: transactionDetailRepo,
		databaseStore:         databaseStore,
		productAdapter:        productAdapter,
		messagingAdapter:      messagingAdapter,
		paymentAdapter:        paymentAdapter,
		cacheAdapter:          cacheAdapter,
		expireTask:            expireTask,
		timeParserHelper:      timeParserHelper,
		validator:             validator,
		log:                   log,
	}
}

func (uc *transactionUseCase) CreateTransaction(ctx context.Context, request *model.CreateTransactionRequest) (*model.CreateTransactionResponse, error) {
	if validatonErrs := uc.validator.ValidateUseCase(request); validatonErrs != nil {
		return nil, validatonErrs
	}

	var transaction *entity.Transaction
	var transactionDetails []*entity.TransactionDetail
	productReqs := make([]*model.CheckProductQuantity, 0, len(request.Products))
	for _, productReq := range request.Products {
		productReqs = append(productReqs, &model.CheckProductQuantity{
			ProductID: productReq.ProductID,
			Quantity:  productReq.Quantity,
			Price:     productReq.Price,
		})
	}

	uc.log.Info("Creating transaction", zap.String("user_id", request.UserID.String()), zap.Any("products", productReqs))
	transactionID := uuid.New()

	products, err := uc.productAdapter.CheckProductAndReserve(ctx, transactionID, productReqs)
	if err != nil {
		uc.log.Warn("failed to check product and reserve", zap.Error(err), zap.String("transaction_id", transactionID.String()))
		return nil, err
	}

	if err := store.BeginTransaction(ctx, uc.log, uc.databaseStore, func(tx store.Transaction) error {
		//test purpose
		// return helper.NewUseCaseError(errorcode.ErrInvalidArgument, message.PriceChanged)

		var totalPrice float64
		for _, product := range productReqs {
			totalPrice += product.Price * float64(product.Quantity)
		}

		transaction = &entity.Transaction{
			ID:                transactionID,
			UserID:            request.UserID,
			TotalPrice:        totalPrice,
			TransactionStatus: enum.TransactionStatusPending,
			InternalStatus:    enum.TrxInternalStatusPending,
		}

		transaction, err = uc.transactionRepo.Insert(ctx, uc.databaseStore, transaction)
		if err != nil {
			return helper.WrapInternalServerError(uc.log, "failed to insert transaction", err)
		}

		transactionDetails = make([]*entity.TransactionDetail, 0, len(products))
		for _, product := range productReqs {
			transactionDetails = append(transactionDetails, &entity.TransactionDetail{
				TransactionID: transaction.ID,
				ProductID:     product.ProductID,
				Quantity:      product.Quantity,
				Price:         product.Price * float64(product.Quantity),
			})
		}

		_, err = uc.transactionDetailRepo.InsertMany(ctx, uc.databaseStore, transactionDetails)
		if err != nil {
			return helper.WrapInternalServerError(uc.log, "failed to insert transaction details", err)
		}

		return nil
	}); err != nil {
		event := &event.TransactionEvent{
			TransactionID: transactionID.String(),
			Status:        enum.TransactionEventCancelled,
		}

		if err := uc.messagingAdapter.Publish(ctx, "transaction.canceled", event); err != nil {
			uc.log.Error("failed to create transaction", zap.Error(err), zap.String("transaction_id", transactionID.String()))
			return nil, helper.WrapInternalServerError(uc.log, "failed to publish transaction canceled event", err)
		}

		return nil, err
	}

	event := &event.TransactionEvent{
		TransactionID: transaction.ID.String(),
		Status:        enum.TransactionEventCommited,
	}

	if err := uc.messagingAdapter.Publish(ctx, "transaction.committed", event); err != nil {
		uc.log.Error("failed to publish transaction committed event", zap.Error(err), zap.String("transaction_id", transactionID.String()))
		return nil, helper.WrapInternalServerError(uc.log, "failed to publish transaction committed event", err)
	}

	token, redirectUrl, err := uc.getPaymentToken(ctx, transaction)
	if err != nil {
		return converter.TransactionToCreateResponse(transaction, ""), nil
	}

	if err := uc.updateTransactionToken(ctx, token, transaction.ID); err != nil {
		uc.log.Error("failed to update transaction token", zap.Error(err), zap.String("transaction_id", transaction.ID.String()))
		return nil, helper.WrapInternalServerError(uc.log, "failed to update transaction token", err)
	}

	transaction.SnapToken = nullable.ToSQLString(&token)

	uc.log.Info("transaction created successfully", zap.Any("transaction", transaction))
	return converter.TransactionToCreateResponse(transaction, redirectUrl), nil
}

func (uc *transactionUseCase) updateTransactionToken(ctx context.Context, token string, transactionID uuid.UUID) error {
	uc.log.Info("Updating transaction token",
		zap.Any("transaction_id", transactionID),
		zap.String("token", token))

	now := time.Now()
	transaction := &entity.Transaction{
		ID:             transactionID,
		InternalStatus: enum.TrxInternalStatusTokenReady,
		SnapToken:      sql.NullString{String: token, Valid: true},
		UpdatedAt:      &now,
	}

	uc.log.Info("Updating transaction token in database",
		zap.Any("transaction_id", transaction.ID),
		zap.String("token", transaction.SnapToken.String))

	err := uc.transactionRepo.UpdateToken(ctx, uc.databaseStore, transaction)
	if err != nil {
		return helper.WrapInternalServerError(uc.log, "failed to update snapshot token in database", err)
	}

	if err = uc.expireTask.EnqueueTransactionExpire(transactionID); err != nil {
		uc.log.Error("failed to enqueue transaction expire task", zap.Error(err), zap.String("transaction_id", transaction.ID.String()))
		return helper.WrapInternalServerError(uc.log, "failed to enqueue transaction expire task", err)
	}

	if err = uc.expireTask.EnqueueTransactionExpireFinal(transactionID); err != nil {
		uc.log.Error("failed to enqueue transaction expire final task", zap.Error(err), zap.String("transaction_id", transaction.ID.String()))
		return helper.WrapInternalServerError(uc.log, "failed to enqueue transaction expire task", err)
	}

	return nil
}

func (uc *transactionUseCase) getPaymentToken(ctx context.Context, transaction *entity.Transaction) (string, string, error) {
	snapRequest := &model.PaymentSnapshotRequest{
		OrderID:     transaction.ID.String(),
		GrossAmount: int64(transaction.TotalPrice),
		Email:       "",
	}

	snapResponse, err := uc.paymentAdapter.CreateSnapshot(ctx, snapRequest)
	if err == nil {
		return snapResponse.Token, snapResponse.RedirectURL, nil
	}

	// Jalankan retry async jika terjadi error
	go func() {
		const maxRetry = 5
		retry := 0

		for retry < maxRetry && !errors.Is(err, gobreaker.ErrOpenState) {
			time.Sleep(time.Second * time.Duration(2<<retry)) // exponential backoff: 2s, 4s, 8s, ...

			snapResponse, retryErr := uc.paymentAdapter.CreateSnapshot(context.TODO(), snapRequest)
			if retryErr == nil {
				uc.log.Info("[RetrySuccess] Transaction berhasil update snap token", zap.String("transaction_id", transaction.ID.String()), zap.String("snap_token", snapResponse.Token))
				if err := uc.updateTransactionToken(ctx, snapResponse.Token, transaction.ID); err != nil {
					uc.log.Warn("[UpdateFailed] Gagal update token transaksi setelah retry", zap.Error(err), zap.String("transaction_id", transaction.ID.String()), zap.String("snap_token", snapResponse.Token))
				} else {
					uc.log.Info("[UpdateSuccess] Token transaksi berhasil diupdate setelah retry", zap.String("transaction_id", transaction.ID.String()), zap.String("snap_token", snapResponse.Token))
				}
				return
			}

			if errors.Is(retryErr, gobreaker.ErrOpenState) {

				break
			}

			err = retryErr
			retry++
		}

		// Jika tetap gagal karena breaker open, listen redis stream
		uc.log.Warn("[Recovery] Gagal membuat snapshot, akan mencoba recovery", zap.Error(err), zap.String("transaction_id", transaction.ID.String()))
		uc.subscribeMidtransRecoveryAndRetry(uc.paymentAdapter, snapRequest, transaction.ID)
	}()

	return "", "", fmt.Errorf("midtrans create snapshot error: %w", err)
}

func (uc *transactionUseCase) subscribeMidtransRecoveryAndRetry(adapter adapter.PaymentAdapter, req *model.PaymentSnapshotRequest, transactionID uuid.UUID) {
	for {
		streams, err := uc.cacheAdapter.XRead(context.Background(), &redis.XReadArgs{
			Streams: []string{"midtrans:recovery", "0"},
			Block:   0, // block selamanya sampai ada signal
			Count:   1,
		})

		if err != nil {
			uc.log.Error("[RecoveryFailed] Gagal membaca stream Redis", zap.Error(err), zap.String("transaction_id", transactionID.String()))
			time.Sleep(time.Second * 5)
			continue
		}

		for _, msg := range streams[0].Messages {
			uc.log.Info("[RecoverySignal] Received recovery signal from Redis stream",
				zap.String("transaction_id", transactionID.String()),
				zap.String("message_id", msg.ID),
				zap.String("message_data", msg.Values["data"].(string)))

			ctx := context.Background()
			resp, err := adapter.CreateSnapshot(ctx, req)
			if err != nil {
				uc.log.Error(fmt.Sprintf("[RecoveryFailed] Retry gagal: tx %s dengan error: %v", transactionID, err))
				continue
			}

			uc.log.Info("[RecoverySuccess] Retry berhasil untuk transaksi", zap.String("transaction_id", transactionID.String()))
			if err := uc.updateTransactionToken(ctx, resp.Token, transactionID); err != nil {
				uc.log.Error(fmt.Sprintf("[UpdateFailed] Update transaction token failed with reason : %v", err))
			}

			return
		}
	}
}

func (uc *transactionUseCase) CheckPaymentSignature(signatureKey, transcationId, statusCode, grossAmount string) (bool, string) {
	uc.log.Info("Checking payment signature",
		zap.String("transaction_id", transcationId),
		zap.String("status_code", statusCode),
		zap.String("gross_amount", grossAmount),
		zap.String("payment_server_key", uc.paymentAdapter.GetPaymentServerKey()))

	signatureToCompare := transcationId + statusCode + grossAmount + uc.paymentAdapter.GetPaymentServerKey()
	hash := sha512.New()
	hash.Write([]byte(signatureToCompare))
	hashedSignature := hex.EncodeToString(hash.Sum(nil))
	return hashedSignature == signatureKey, hashedSignature
}

func (uc *transactionUseCase) CheckAndUpdateTransaction(ctx context.Context, request *model.CheckAndUpdateTransactionRequest) error {
	if validatonErrs := uc.validator.ValidateUseCase(request); validatonErrs != nil {
		return validatonErrs
	}

	var transaction *entity.Transaction
	var transactionStatus enum.TransactionStatus
	var transactionInternalStatus enum.TrxInternalStatus
	var settlementTimePtr *time.Time
	var err error

	if err := store.BeginTransaction(ctx, uc.log, uc.databaseStore, func(tx store.Transaction) error {
		transaction, err = uc.transactionRepo.FindByID(ctx, tx, request.OrderID, true)
		if err != nil {
			if strings.Contains(err.Error(), pgx.ErrNoRows.Error()) {
				return helper.NewUseCaseError(errorcode.ErrResourceNotFound, message.TransactionNotFound)
			}
			return helper.WrapInternalServerError(uc.log, "failed to update transaction callback in database", err)
		}

		requestIsValid, hashedSignature := uc.CheckPaymentSignature(request.SignatureKey, transaction.ID.String(), request.StatusCode, request.GrossAmount)
		if !requestIsValid {
			uc.log.Warn("Invalid signature key", zap.String("transaction_id", transaction.ID.String()), zap.String("hashed_signature", hashedSignature))
			return helper.NewUseCaseError(errorcode.ErrForbidden, "Invalid signature key")
		}

		//Only check if internal transaction status is "PENDING", "TOKEN_READY", "EXPIRED"
		if transaction.InternalStatus != enum.TrxInternalStatusPending &&
			transaction.InternalStatus != enum.TrxInternalStatusTokenReady &&
			transaction.InternalStatus != enum.TrxInternalStatusExpired {
			uc.log.Warn("Transaction internal status already final", zap.String("transaction_id", transaction.ID.String()), zap.String("internal_status", string(transaction.InternalStatus)))
			return nil
		}

		now := time.Now()

		if transaction.InternalStatus == enum.TrxInternalStatusExpired {
			// IF External Or Midtrans Payment Settled Check the Settlement Time
			if request.MidtransTransactionStatus == string(enum.PaymentStatusSettlement) {
				settlementTime, err := uc.timeParserHelper.TimeParseInDefaultLocation(request.SettlementTime)
				if err != nil {
					uc.log.Error("failed to parse settlement time from external webhook request", zap.String("transaction_id", request.OrderID), zap.Error(err))
					return err
				}

				settlementTimePtr = &settlementTime
				graceDeadline := transaction.UpdatedAt.Add(5 * time.Minute)

				//Checking settlement time from external midtrans with internal expired time
				if settlementTime.After(graceDeadline) {
					transactionStatus = enum.TransactionStatusExpired //User transaction is valid or settled
					transactionInternalStatus = enum.TrxInternalStatusLateSettlement
				} else {
					//User transaction is valid or settled even the user come late but from system not the settled time
					transactionStatus = enum.TransactionStatusSuccess
					transactionInternalStatus = enum.TrxInternalStatusExpiredCheckedValid
					transaction.PaymentAt = nullable.ToSQLTime(now)
				}
				transaction.ExternalSettlementAt = nullable.ToSQLTime(now)
				transaction.SnapToken = sql.NullString{String: "", Valid: true}
			} else {
				transactionInternalStatus = enum.TrxInternalStatusExpiredCheckedInvalid //User doesnt settled even internal status has expired
				transactionStatus = enum.TransactionStatusExpired
			}

		} else {
			switch request.MidtransTransactionStatus {
			case string(enum.PaymentStatusCapture), string(enum.PaymentStatusSettlement):
				settlementTime, err := uc.timeParserHelper.TimeParseInDefaultLocation(request.SettlementTime)
				if err != nil {
					uc.log.Error("failed to parse settlement time from external webhook request", zap.String("transaction_id", request.OrderID), zap.Error(err))
					return err
				}

				settlementTimePtr = &settlementTime
				transactionStatus = enum.TransactionStatusSuccess
				transactionInternalStatus = enum.TrxInternalStatusSettled
				transaction.PaymentAt = nullable.ToSQLTime(now)
				transaction.SnapToken = sql.NullString{String: "", Valid: true}
				transaction.ExternalSettlementAt = nullable.ToSQLTime(*settlementTimePtr)

			case string(enum.PaymentStatusPending):
				transactionStatus = enum.TransactionStatusPending
				transactionInternalStatus = enum.TrxInternalStatusPending

			case string(enum.PaymentStatusExpire):
				transactionStatus = enum.TransactionStatusExpired
				transactionInternalStatus = enum.TrxInternalStatusExpiredCheckedInvalid
				transaction.SnapToken = sql.NullString{String: "", Valid: true}

			case string(enum.PaymentStatusFailure), string(enum.PaymentStatusDeny):
				transactionStatus = enum.TransactionStatusFailed
				transactionInternalStatus = enum.TrxInternalStatusFailed
				transaction.SnapToken = sql.NullString{String: "", Valid: true}

			case string(enum.PaymentStatusCancel):
				transactionStatus = enum.TransactionStatusCancelled
				transactionInternalStatus = enum.TrxInternalStatusCancelledBySystem
				transaction.SnapToken = sql.NullString{String: "", Valid: true}
			}
		}

		transaction.TransactionStatus = transactionStatus
		transaction.InternalStatus = transactionInternalStatus
		transaction.ExternalStatus = sql.NullString{
			String: request.MidtransTransactionStatus,
			Valid:  true,
		}

		externalCallbackResponse := json.RawMessage(request.Body)
		transaction.ExternalCallbackResponse = &externalCallbackResponse

		if err := uc.transactionRepo.UpdateCallback(ctx, tx, transaction); err != nil {
			return helper.WrapInternalServerError(uc.log, "failed to update transaction callback in database", err)
		}

		return nil
	}); err != nil {
		return helper.WrapInternalServerError(uc.log, "failed to update transaction callback in database", err)
	}

	switch transactionStatus {
	case enum.TransactionStatusCancelled:
		event := &event.TransactionEvent{
			TransactionID: transaction.ID.String(),
			Status:        enum.TransactionEventCancelled,
		}

		if err := uc.messagingAdapter.Publish(ctx, "transaction.canceled", event); err != nil {
			return helper.WrapInternalServerError(uc.log, "failed to publish transaction canceled event", err)
		}
	case enum.TransactionStatusExpired, enum.TransactionStatusFailed:
		event := &event.TransactionEvent{
			TransactionID: transaction.ID.String(),
			Status:        enum.TransactionEventExpired,
		}

		if err := uc.messagingAdapter.Publish(ctx, "transaction.expired", event); err != nil {
			return helper.WrapInternalServerError(uc.log, "failed to publish transaction expired event", err)
		}
	}

	if transactionStatus != enum.TransactionStatusSuccess {
		uc.log.Info("Transaction is not successful, no further processing required",
			zap.String("transaction_id", transaction.ID.String()),
			zap.String("transaction_status", string(transaction.TransactionStatus)),
			zap.String("internal_status", string(transaction.InternalStatus)))

		return nil
	}

	event := &event.TransactionEvent{
		TransactionID: transaction.ID.String(),
		Status:        enum.TransactionEventSettled,
	}

	if err := uc.messagingAdapter.Publish(ctx, "transaction.settled", event); err != nil {
		return helper.WrapInternalServerError(uc.log, "failed to publish transaction settled event", err)
	}

	return nil
}

func (uc *transactionUseCase) UserSearch(ctx context.Context, request *model.UserSearchTransactionRequest) ([]*model.TransactionResponse, *web.PageMetadata, error) {
	products, metadata, err := uc.transactionRepo.FindManyByUserID(ctx, uc.databaseStore, request)
	if err != nil {
		return nil, nil, helper.WrapInternalServerError(uc.log, "failed to find public products", err)
	}

	if products == nil {
		return nil, metadata, nil
	}

	return converter.TransactionsWithTotalToResponses(products), metadata, nil
}

func (uc *transactionUseCase) UserSearchWithDetail(ctx context.Context, request *model.UserSearchTransactionRequest) ([]*model.TransactionResponse, *web.PageMetadata, error) {
	transactions, metadata, err := uc.transactionRepo.FindManyWithDetailByUserID(ctx, uc.databaseStore, request)
	if err != nil {
		return nil, nil, helper.WrapInternalServerError(uc.log, "failed to find public products", err)
	}

	if transactions == nil {
		return nil, metadata, nil
	}

	return converter.TransactionWithDetailAndTotalToResponse(transactions, false), metadata, nil
}

func (uc *transactionUseCase) OwnerSearchWithDetail(ctx context.Context, request *model.OwnerSearchTransactionRequest) ([]*model.TransactionResponse, *web.PageMetadata, error) {
	if _, err := uc.productAdapter.OwnerGetProduct(ctx, request.UserID, request.ProductID); err != nil {
		return nil, nil, err
	}

	transactions, metadata, err := uc.transactionRepo.FindManyWithDetailByProductID(ctx, uc.databaseStore, request)
	if err != nil {
		return nil, nil, helper.WrapInternalServerError(uc.log, "failed to find public products", err)
	}

	if transactions == nil {
		return nil, metadata, nil
	}

	return converter.TransactionWithDetailAndTotalToResponse(transactions, true), metadata, nil
}
