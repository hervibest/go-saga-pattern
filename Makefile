include .env

PRODUCT_MIGRATIONS_DIR= product-svc/db/migrations
TRANSACTION_MIGRATIONS_DIR= transcation-svc/db/migrations

PRODUCT_DB_URL = "postgres://${DB_USERNAME}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${PRODUCT_DB_NAME}?sslmode=disable&TimeZone=Asia/Jakarta"
TRANSACTION_DB_URL = "postgres://${DB_USERNAME}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${TRANSACTION_DB_NAME}?sslmode=disable&TimeZone=Asia/Jakarta"

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