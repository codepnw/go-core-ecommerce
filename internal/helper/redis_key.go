package helper

import "fmt"

func RedisProductStockKey(productID int64) string {
	return fmt.Sprintf("product:stock:%d", productID)
}

func RedisProductItemKey(productID int64) string {
	return fmt.Sprintf("product:%d", productID)
}

func RedisProductListKey(page, limit int) string {
	return fmt.Sprintf("products:page:%d:limit:%d", page, limit)
}

