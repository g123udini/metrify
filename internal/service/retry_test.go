package service

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestRetry(t *testing.T) {
	type args[T any] struct {
		attempts int
		base     time.Duration
		step     time.Duration
		fn       func() (T, error)
	}
	type testCase[T any] struct {
		name    string
		args    args[T]
		want    T
		wantErr bool
	}

	tests := []testCase[int]{
		{
			name: "success on first try",
			args: args[int]{
				attempts: 3,
				base:     0,
				step:     0,
				fn: func() (int, error) {
					return 42, nil
				},
			},
			want:    42,
			wantErr: false,
		},
		{
			name: "success after a few failures",
			args: args[int]{
				attempts: 3,
				base:     0,
				step:     0,
				fn: func() func() (int, error) {
					calls := 0
					return func() (int, error) {
						calls++
						if calls < 2 {
							return 0, errors.New("temporary error")
						}
						return 7, nil
					}
				}(),
			},
			want:    7,
			wantErr: false,
		},
		{
			name: "all attempts failed",
			args: args[int]{
				attempts: 2,
				base:     0,
				step:     0,
				fn: func() (int, error) {
					return 0, errors.New("permanent error")
				},
			},
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := Retry(tt.args.attempts, tt.args.base, tt.args.step, tt.args.fn)
			if (err != nil) != tt.wantErr {
				t.Errorf("Retry() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Retry() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRetryDB(t *testing.T) {
	type args[T any] struct {
		attempts int
		base     time.Duration
		step     time.Duration
		fn       func() (T, error)
	}
	type testCase[T any] struct {
		name      string
		args      args[T]
		want      T
		wantErr   bool
		wantCalls int
		calls     *int
	}

	tests := []testCase[int]{
		{
			name: "retry on connection failure then success",
			args: func() args[int] {
				calls := 0
				return args[int]{
					attempts: 3,
					base:     0,
					step:     0,
					fn: func() (int, error) {
						calls++
						if calls < 2 {
							return 0, &pgconn.PgError{Code: pgerrcode.ConnectionFailure}
						}
						return 100, nil
					},
				}
			}(),
			want:      100,
			wantErr:   false,
			wantCalls: 2,
			calls: func() *int {
				v := 0
				return &v
			}(),
		},
		{
			name: "non pg error - no retry",
			args: func() args[int] {
				calls := 0
				return args[int]{
					attempts: 5,
					base:     0,
					step:     0,
					fn: func() (int, error) {
						calls++
						return 0, errors.New("some error")
					},
				}
			}(),
			want:      0,
			wantErr:   true,
			wantCalls: 1,
			calls: func() *int {
				v := 0
				return &v
			}(),
		},
		{
			name: "pg error but non-retryable code - no retry",
			args: func() args[int] {
				calls := 0
				return args[int]{
					attempts: 5,
					base:     0,
					step:     0,
					fn: func() (int, error) {
						calls++
						return 0, &pgconn.PgError{Code: "XXXXX"} // не тот код
					},
				}
			}(),
			want:      0,
			wantErr:   true,
			wantCalls: 1,
			calls: func() *int {
				v := 0
				return &v
			}(),
		},
		{
			name: "retryable pg error but never succeeds",
			args: func() args[int] {
				calls := 0
				return args[int]{
					attempts: 3,
					base:     0,
					step:     0,
					fn: func() (int, error) {
						calls++
						return 0, &pgconn.PgError{Code: pgerrcode.DeadlockDetected}
					},
				}
			}(),
			want:      0,
			wantErr:   true,
			wantCalls: 2,
			calls: func() *int {
				v := 0
				return &v
			}(),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			calls := 0
			tt.calls = &calls

			origFn := tt.args.fn
			tt.args.fn = func() (int, error) {
				calls++
				return origFn()
			}

			got, err := RetryDB(tt.args.attempts, tt.args.base, tt.args.step, tt.args.fn)
			if (err != nil) != tt.wantErr {
				t.Errorf("RetryDB() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RetryDB() got = %v, want %v", got, tt.want)
			}
			if tt.calls != nil && *tt.calls != tt.wantCalls {
				t.Errorf("RetryDB() calls = %d, want %d", *tt.calls, tt.wantCalls)
			}
		})
	}
}
