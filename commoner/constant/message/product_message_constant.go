package message

const (
	//buyer side
	UnavailableProduct                = "Product is unavailable or already reserved by other transaction"
	ProductOutOfStock                 = "Product is out of stock, please check again"
	RequestedProductMoreThanAvailable = "Requested product quantity is more than available stock"
	PriceChanged                      = "Product price has been changed, please check again"
	ProductNotFound                   = "Product not found for the given id/uuid"

	//owner side
	ProductNotFoundOrAlreadyDeleted = "Product not found or already deleted"
	ProductIsExistsByNameOrSlug     = "Product with the same name or slug already exists"
	ProductTranscationNotFound      = "Product transaction not found for the given id/uuid"
)
