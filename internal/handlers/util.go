package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

func fmtUint(u uint) string {
	return strconv.FormatUint(uint64(u), 10)
}

func parseUintParam(c *gin.Context, key string) (uint, error) {
	val := c.Param(key)
	u64, err := strconv.ParseUint(val, 10, 64)
	return uint(u64), err
}

// getRestaurantIDFromContext extracts restaurant_id from auth context
func getRestaurantIDFromContext(c *gin.Context) uint {
	if v, ok := c.Get("restaurant_id"); ok {
		if id, ok := v.(uint); ok {
			return id
		}
		if id64, ok := v.(int64); ok {
			return uint(id64)
		}
		if idf, ok := v.(float64); ok { // when claims decode numbers as float64
			return uint(idf)
		}
		if s, ok := v.(string); ok {
			if u64, err := strconv.ParseUint(s, 10, 64); err == nil {
				return uint(u64)
			}
		}
	}
	return 0
}
