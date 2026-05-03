package controllers

import (
	"context"
	"errors"
	"testing"
	"time"

	"google.golang.org/grpc"
)

func TestDefaultDeadlineUnaryInterceptor(t *testing.T) {
	const fallback = 30 * time.Minute

	captureCtxInvoker := func(captured *context.Context) grpc.UnaryInvoker {
		return func(ctx context.Context, _ string, _, _ any, _ *grpc.ClientConn, _ ...grpc.CallOption) error {
			*captured = ctx
			return nil
		}
	}

	t.Run("applies fallback deadline when caller has none", func(t *testing.T) {
		var got context.Context
		err := defaultDeadlineUnaryInterceptor(fallback)(context.Background(), "/x", nil, nil, nil, captureCtxInvoker(&got))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		dl, ok := got.Deadline()
		if !ok {
			t.Fatal("expected interceptor to attach a deadline")
		}
		if remaining := time.Until(dl); remaining <= 0 || remaining > fallback {
			t.Fatalf("deadline outside expected window: remaining=%v", remaining)
		}
	})

	t.Run("preserves caller deadline when one is already set", func(t *testing.T) {
		callerDeadline := time.Now().Add(2 * time.Second)
		ctx, cancel := context.WithDeadline(context.Background(), callerDeadline)
		defer cancel()

		var got context.Context
		err := defaultDeadlineUnaryInterceptor(fallback)(ctx, "/x", nil, nil, nil, captureCtxInvoker(&got))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		dl, ok := got.Deadline()
		if !ok {
			t.Fatal("expected caller deadline to be preserved")
		}
		if !dl.Equal(callerDeadline) {
			t.Fatalf("interceptor overrode caller deadline: got=%v want=%v", dl, callerDeadline)
		}
	})

	t.Run("zero timeout disables the fallback", func(t *testing.T) {
		var got context.Context
		err := defaultDeadlineUnaryInterceptor(0)(context.Background(), "/x", nil, nil, nil, captureCtxInvoker(&got))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, ok := got.Deadline(); ok {
			t.Fatal("expected no deadline when timeout is zero")
		}
	})

	t.Run("propagates invoker error", func(t *testing.T) {
		want := errors.New("boom")
		invoker := func(_ context.Context, _ string, _, _ any, _ *grpc.ClientConn, _ ...grpc.CallOption) error {
			return want
		}
		got := defaultDeadlineUnaryInterceptor(fallback)(context.Background(), "/x", nil, nil, nil, invoker)
		if !errors.Is(got, want) {
			t.Fatalf("expected error to propagate: got=%v want=%v", got, want)
		}
	})
}
