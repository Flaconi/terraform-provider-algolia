package algoliautil

import (
	"errors"
	"net/http"

	"github.com/algolia/algoliasearch-client-go/v4/algolia/errs"
	suggestions "github.com/algolia/algoliasearch-client-go/v4/algolia/query-suggestions"
	"github.com/algolia/algoliasearch-client-go/v4/algolia/search"
)

func IsRetryableError(err error) bool {
	if IsNotFoundError(err) {
		return true
	}
	if errors.Is(err, errs.ErrNoMoreHostToTry) {
		return true
	}
	return false
}

func IsNotFoundError(err error) bool {
	var apiErr *search.APIError
	if errors.As(err, &apiErr) {
		return apiErr.Status == http.StatusNotFound
	}
	var suggErr *suggestions.APIError
	if errors.As(err, &suggErr) {
		return suggErr.Status == http.StatusNotFound
	}
	return false
}

func IsForbiddenError(err error) bool {
	var apiErr *search.APIError
	if errors.As(err, &apiErr) {
		return apiErr.Status == http.StatusForbidden
	}
	var suggErr *suggestions.APIError
	if errors.As(err, &suggErr) {
		return suggErr.Status == http.StatusForbidden
	}
	return false
}
