package interceptor_test

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/tiny-go/interceptor"
)

func Test_Retry_WithInterval_ContextDeadline(t *testing.T) {
	calls := make(callChain, 2)
	calls <- doReadBodyAndFail
	calls <- http.DefaultClient.Do

	ts := httptest.NewServer(serveWithCode(http.StatusOK))
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ts.URL, http.NoBody)
	if err != nil {
		t.Fatalf("cannot instantiate a client: %s", err)
	}

	res, err := interceptor.New(
		interceptor.Retry(3),
		interceptor.WithInterval(time.Second),
	).Then(calls).Do(req)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("unexpected error: %s", err)
	}

	if res != nil {
		t.Error("the response was expected to be nil")
	}
}

func Test_Retry_WithInterval(t *testing.T) {
	calls := make(callChain, 4)
	calls <- doReadBodyAndFail
	calls <- doReadBodyAndFail
	calls <- doReadBodyAndFail
	calls <- http.DefaultClient.Do

	ts := httptest.NewServer(serveWithCode(http.StatusOK))
	defer ts.Close()

	body := bytes.NewBufferString("foo bar baz")

	req, err := http.NewRequest(http.MethodPost, ts.URL, body)
	if err != nil {
		t.Fatalf("cannot instantiate a client: %s", err)
	}

	res, err := interceptor.New(
		interceptor.Retry(3),
		interceptor.WithInterval(100*time.Millisecond),
	).Then(calls).Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	defer req.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status code: %d", res.StatusCode)
	}
}
