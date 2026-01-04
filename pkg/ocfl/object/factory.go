package object

import (
	"context"
)

type Factory interface {
	NewObject(ctx context.Context) Object
}
