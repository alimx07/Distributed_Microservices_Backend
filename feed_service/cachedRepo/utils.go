package cachedrepo

import (
	"encoding/base64"
	"log"
	"strconv"
	"strings"
)

func DecodeCursor(c string) (string, int64, error) {
	decode, err := base64.StdEncoding.DecodeString(c)
	var cursor string
	var PageSize int64
	if err != nil {
		log.Println("Einvalid Cursor", err.Error())
		// continue with null cursor
		cursor = "+inf"
		PageSize = 50 // default value
	} else {
		// decoded successfully
		parts := strings.Split(string(decode), ":")
		if len(parts) != 2 {
			log.Println("Invalid Cursor")
			cursor = "+inf"
			PageSize = 50 // default value
		} else {
			cursor = parts[0]
			if v, err := strconv.ParseInt(parts[1], 10, 64); err != nil {
				PageSize = v
			} else {
				PageSize = 50
			}
		}
	}
	return cursor, PageSize, nil
}
