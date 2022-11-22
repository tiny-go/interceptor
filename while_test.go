package interceptor_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tiny-go/interceptor"
)

func Test_Retry_WhileFails(t *testing.T) {
	// 1 try + 3 retries = 4 calls in total
	handler := make(serveChain, 4)
	handler <- serveWithCode(http.StatusServiceUnavailable)
	handler <- serveWithCode(http.StatusInternalServerError)
	handler <- serveWithCode(http.StatusInternalServerError)
	handler <- serveWithCode(http.StatusNotFound)

	ts := httptest.NewServer(handler)
	defer ts.Close()

	req, err := http.NewRequest(http.MethodPost, ts.URL, http.NoBody)
	if err != nil {
		t.Fatalf("cannot instantiate a client: %s", err)
	}

	res, err := interceptor.New(
		interceptor.Retry(3),
		interceptor.WhileFails(func(r *http.Response) error {
			if r.StatusCode >= 500 {
				return fmt.Errorf("unexpected status code: %d", r.StatusCode)
			}

			return nil
		}),
	).Then(http.DefaultClient).Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	defer req.Body.Close()

	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("unexpected status code: %d", res.StatusCode)
	}
}
