include .env

USER_MIGRATIONS_DIR= user-svc/db/migrations
PRODUCT_MIGRATIONS_DIR= product-svc/db/migrations
TRANSACTION_MIGRATIONS_DIR= transaction-svc/db/migrations

USER_DB_URL = "postgres://${DB_USERNAME}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${USER_DB_NAME}?sslmode=disable&TimeZone=Asia/Jakarta"
PRODUCT_DB_URL = "postgres://${DB_USERNAME}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${PRODUCT_DB_NAME}?sslmode=disable&TimeZone=Asia/Jakarta"
TRANSACTION_DB_URL = "postgres://${DB_USERNAME}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${TRANSACTION_DB_NAME}?sslmode=disable&TimeZone=Asia/Jakarta"

user-migrate-up :
	goose -dir ${USER_MIGRATIONS_DIR} postgres ${USER_DB_URL} up
	
user-migrate-down :
	goose -dir ${USER_MIGRATIONS_DIR} postgres ${USER_DB_URL} down
	
user-migrate-down-to-zero :
	goose -dir ${USER_MIGRATIONS_DIR} postgres ${USER_DB_URL} down-to 0
	
product-migrate-up :
	goose -dir ${PRODUCT_MIGRATIONS_DIR} postgres ${PRODUCT_DB_URL} up

product-migrate-down :
	goose -dir ${PRODUCT_MIGRATIONS_DIR} postgres ${PRODUCT_DB_URL} down

product-migrate-down-to-zero :
	goose -dir ${PRODUCT_MIGRATIONS_DIR} postgres ${PRODUCT_DB_URL} down-to 0

transaction-migrate-up :
	goose -dir ${TRANSACTION_MIGRATIONS_DIR} postgres ${TRANSACTION_DB_URL} up

transaction-migrate-down :
	goose -dir ${TRANSACTION_MIGRATIONS_DIR} postgres ${TRANSACTION_DB_URL} down

transaction-migrate-down-to-zero :
	goose -dir ${TRANSACTION_MIGRATIONS_DIR} postgres ${TRANSACTION_DB_URL} down-to 0

generate-proto-product:
	cd proto && protoc --go_out=. --go-grpc_out=. product.proto

generate-proto-user:
	cd proto && protoc --go_out=. --go-grpc_out=. user.proto

start-user-svc:
	cd user-svc/cmd/web && go run main.gomockgen -version

start-product-svc:
	cd product-svc/cmd/web && go run main.go

start-transaction-svc:
	cd transaction-svc/cmd/web && go run main.go

start-transaction-worker:
	cd transaction-svc/cmd/worker && go run main.go

start-listener-svc:
	cd transaction-svc/cmd/listener && go run main.go

mockgen-user-svc:
	cd user-svc/internal && \
	mockgen -source=./repository/store/db.go \
	        -destination=./mocks/store/mock_db.go \
	        -package=mockstore

mockgen-log:
	cd commoner && \
	mockgen -source=./helper/validate_helper.go \
	        -destination=./mocks/commoner/helper/mock_validator.go \
	        -package=mockhelper