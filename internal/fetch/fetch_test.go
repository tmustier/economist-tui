package fetch

import (
	"errors"
	"testing"

	appErrors "github.com/tmustier/economist-tui/internal/errors"
)

func TestNormalizeErrorPaywall(t *testing.T) {
	err := normalizeError(appErrors.PaywallError{})
	if !appErrors.IsUserError(err) {
		t.Fatalf("expected user error, got %v", err)
	}
	expected := "paywall detected - run 'economist login' to read full articles"
	if err.Error() != expected {
		t.Fatalf("expected %q, got %q", expected, err.Error())
	}
}

func TestNormalizeErrorPassThrough(t *testing.T) {
	base := errors.New("boom")
	if normalizeError(base) != base {
		t.Fatalf("expected same error instance")
	}

	userErr := appErrors.NewUserError("no content")
	if normalizeError(userErr) != userErr {
		t.Fatalf("expected user error pass-through")
	}
}
