package algoliautil

import (
	"errors"
	"net/http"
	"testing"

	"github.com/algolia/algoliasearch-client-go/v4/algolia/errs"
	suggestions "github.com/algolia/algoliasearch-client-go/v4/algolia/query-suggestions"
	"github.com/algolia/algoliasearch-client-go/v4/algolia/search"
)

func TestIsRetryableError(t *testing.T) {
	t.Parallel()

	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "returns true if not found error",
			args: args{
				err: &search.APIError{
					Message: "not found",
					Status:  http.StatusNotFound,
				},
			},
			want: true,
		},
		{
			name: "returns true if no more host error",
			args: args{
				err: errs.ErrNoMoreHostToTry,
			},
			want: true,
		},
		{
			name: "returns false if not retryable error",
			args: args{err: errors.New("test")},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsRetryableError(tt.args.err); got != tt.want {
				t.Errorf("IsRetryableError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsNotFoundError(t *testing.T) {
	t.Parallel()

	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "returns true if not found error",
			args: args{
				err: &search.APIError{
					Message: "not found",
					Status:  http.StatusNotFound,
				},
			},
			want: true,
		},
		{
			name: "returns false if not found error",
			args: args{
				err: &search.APIError{
					Message: "bad request",
					Status:  http.StatusBadRequest,
				},
			},
			want: false,
		},
		{
			name: "returns true if suggestions not found error",
			args: args{
				err: &suggestions.APIError{
					Message: "not found",
					Status:  http.StatusNotFound,
				},
			},
			want: true,
		},
		{
			name: "returns false if not algolia error",
			args: args{err: errors.New("test")},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNotFoundError(tt.args.err); got != tt.want {
				t.Errorf("IsNotFoundError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsForbiddenError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "search 403",
			err:  &search.APIError{Status: http.StatusForbidden},
			want: true,
		},
		{
			name: "suggestions 403",
			err:  &suggestions.APIError{Status: http.StatusForbidden},
			want: true,
		},
		{
			name: "search 404 is not forbidden",
			err:  &search.APIError{Status: http.StatusNotFound},
			want: false,
		},
		{
			name: "non-algolia error",
			err:  errors.New("boom"),
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsForbiddenError(tt.err); got != tt.want {
				t.Errorf("IsForbiddenError() = %v, want %v", got, tt.want)
			}
		})
	}
}
