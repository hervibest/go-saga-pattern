package helper

import (
	"go-saga-pattern/commoner/web"
	"math"
	"net/url"
	"strconv"
)

func CalculatePagination(totalItems int64, page, size int) *web.PageMetadata {
	totalPages := int64(math.Ceil(float64(totalItems) / float64(size)))
	offset := (page - 1) * size

	return &web.PageMetadata{
		Page:        page,
		Size:        size,
		Offset:      offset,
		TotalItem:   totalItems,
		TotalPage:   totalPages,
		HasNext:     page < int(totalPages),
		HasPrevious: page > 1,
	}
}

func GeneratePageURLs(baseURL string, metadata *web.PageMetadata) {
	parsedURL, _ := url.Parse(baseURL)
	q := parsedURL.Query()

	if metadata.HasNext {
		q.Set("page", strconv.Itoa(metadata.Page+1))
		q.Set("size", strconv.Itoa(metadata.Size))
		parsedURL.RawQuery = q.Encode()
		metadata.NextPageURL = parsedURL.String()
	}

	if metadata.HasPrevious {
		q.Set("page", strconv.Itoa(metadata.Page-1))
		q.Set("size", strconv.Itoa(metadata.Size))
		parsedURL.RawQuery = q.Encode()
		metadata.PreviousPageURL = parsedURL.String()
	}
}
