package handlers

import (
	"context"
	"fmt"

	"github.com/hibiken/asynq"
)

func HandleUnknown(ctx context.Context, task *asynq.Task) error {
	return fmt.Errorf("%w: unregistered task type %q", asynq.SkipRetry, task.Type())
}
