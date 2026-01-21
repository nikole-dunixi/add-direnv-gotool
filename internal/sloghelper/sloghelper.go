package sloghelper

import (
	"context"
	"log/slog"
	"os"
)

func FatalContextErr(ctx context.Context, err error, msg string, args ...any) {
	FatalContext(ctx, msg, append(args, slog.Any("err", err))...)
}

func FatalContext(ctx context.Context, msg string, args ...any) {
	slog.ErrorContext(ctx, msg, args...)
	os.Exit(1)
}
