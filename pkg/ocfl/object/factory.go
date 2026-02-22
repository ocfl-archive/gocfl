package object

import (
	"context"
)

type Factory interface {
	NewObject(ctx context.Context) Object
	NewChecker(ctx context.Context) Checker
}
