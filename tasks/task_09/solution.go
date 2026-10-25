package main

import (
	"context"
	"errors"
	"sync"
)

type Result[R any] struct {
	Value R
	Err   error
}

var ErrInvalidWorkers = errors.New("workers must be positive")

func ParallelMap[T any, R any](ctx context.Context, workers int, in []T, fn func(context.Context, T) (R, error)) ([]Result[R], error) {
	if workers <= 0 {
		return nil, ErrInvalidWorkers
	}

	result := make([]Result[R], len(in))
	jobs := make(chan int)
	wg := sync.WaitGroup{}
	for range min(len(in), workers) {
		wg.Go(func() {
			for i := range jobs {
				res, err := fn(ctx, in[i])
				if err != nil {
					result[i] = Result[R]{Err: err}
				} else {
					result[i] = Result[R]{Value: res}
				}
			}
		})
	}

feed:
	for i := range in {
		if ctx.Err() != nil {
			break
		}
		select {
		case jobs <- i:
		case <-ctx.Done():
			break feed
		}
	}
	close(jobs)
	wg.Wait()

	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return result, ctx.Err()
}
