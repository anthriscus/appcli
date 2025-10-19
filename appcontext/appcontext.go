package appcontext

import (
	"fmt"
	"time"
)

type ContextKey string

const (
	TraceIdKey ContextKey = "TraceID"
)

func GenerateId() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
