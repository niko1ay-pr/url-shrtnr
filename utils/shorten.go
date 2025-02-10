package utils

import (
	"encoding/base64"
	"fmt"
	"time"
)

func GetShortCode() string {
	fmt.Println("Shortening URL")

	timestamp := time.Now().UnixNano()
	fmt.Println("Timestamp:	", timestamp)

	key := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d", timestamp)))
	fmt.Println("Key:			", key)

	key = key[:len(key)-2]
	return key[:16]
}
