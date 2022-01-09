package goretry

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"
)

const testRetrySuccess = "success"

func Test_retry(t *testing.T) {
	ctx := context.Background()
	count := 0

	type args struct {
		fn     RetryFunc
		cRetry *CustomRetry
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name: "Normal test case 1 - no retry",
			args: args{
				fn: func(ctx context.Context) (interface{}, error) {
					return testRetrySuccess, nil
				},
				cRetry: &CustomRetry{
					RetryCount: 1,
					Cond:       defaultCond,
					Timeout:    defaultTimeout,
					apiChan:    make(chan continueStruct),
				},
			},
			want: testRetrySuccess,
		},
		{
			name: "Normal test case 2 - 2 retry",
			args: args{
				fn: func(ctx context.Context) (interface{}, error) {
					if count == 2 {
						return testRetrySuccess, nil
					}
					count++

					return "", errors.New("error")
				},
				cRetry: &CustomRetry{
					RetryCount: 3,
					Cond:       defaultCond,
					Timeout:    defaultTimeout,
					apiChan:    make(chan continueStruct),
				},
			},
			want: testRetrySuccess,
		},
		{
			name: "Normal test case 3 - with timeout",
			args: args{
				fn: func(ctx context.Context) (interface{}, error) {
					return testRetrySuccess, nil
				},
				cRetry: &CustomRetry{
					RetryCount: noRetryCount,
					Timeout:    time.Millisecond * 10,
					Cond:       defaultCond,
					apiChan:    make(chan continueStruct),
				},
			},
			want: testRetrySuccess,
		},
		{
			name: "Failed test case 1 - retry count exceeded",
			args: args{
				fn: func(ctx context.Context) (interface{}, error) {
					return nil, errors.New("error")
				},
				cRetry: &CustomRetry{
					RetryCount: 3,
					Cond:       defaultCond,
					Timeout:    defaultTimeout,
					apiChan:    make(chan continueStruct),
				},
			},
			wantErr: true,
		},
		{
			name: "Failed test case 2 - retry count exceeds",
			args: args{
				fn: func(ctx context.Context) (interface{}, error) {
					return nil, errors.New("error")
				},
				cRetry: &CustomRetry{
					RetryCount: 3,
					Timeout:    time.Millisecond * 5,
					Cond:       defaultCond,
					RetryDelay: time.Second,
					apiChan:    make(chan continueStruct),
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count = 0

			retry(ctx, tt.args.fn, tt.args.cRetry)
			response := <-tt.args.cRetry.apiChan
			got, err := response.resp, response.respErr
			if (err != nil) != tt.wantErr {
				t.Errorf("retry() error = %v, wantErr %v", err, tt.wantErr)

				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("retry() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCustomRetry_Wait(t *testing.T) {
	ctx := context.Background()
	count := 0

	type args struct {
		fn RetryFunc
	}

	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name: "Normal test case 1 - no retry",
			args: args{
				fn: func(ctx context.Context) (interface{}, error) {
					return testRetrySuccess, nil
				},
			},
			want: testRetrySuccess,
		},
		{
			name: "Normal test case 2 - Default retry values",
			args: args{
				fn: func(ctx context.Context) (interface{}, error) {
					if count == defaultRetryCount-1 {
						return testRetrySuccess, nil
					}
					count++

					return "", errors.New("error")
				},
			},
			want: testRetrySuccess,
		},
		{
			name: "Failed test case 1 - Default retry count exceeds",
			args: args{
				fn: func(ctx context.Context) (interface{}, error) {
					return nil, errors.New("error")
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := CustomRetry{}
			count = 0
			a.RetryParallel(ctx, tt.args.fn)
			got, err := a.Wait()
			if (err != nil) != tt.wantErr {
				t.Errorf("CustomRetry.Wait() error = %v, wantErr %v", err, tt.wantErr)

				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CustomRetry.Wait() = %v, want %v", got, tt.want)
			}
		})
	}
}
