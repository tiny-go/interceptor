package interceptor_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tiny-go/interceptor"
)

func Test_Retry_UntilStatusCodeIs(t *testing.T) {
	handler := make(serveChain, 3)
	handler <- serveWithCode(http.StatusNotFound)
	handler <- serveWithCode(http.StatusBadRequest)
	handler <- serveWithCode(http.StatusOK)

	ts := httptest.NewServer(handler)
	defer ts.Close()

	req, err := http.NewRequest(http.MethodPost, ts.URL, http.NoBody)
	if err != nil {
		t.Fatalf("cannot instantiate a client: %s", err)
	}

	res, err := interceptor.New(
		interceptor.Retry(3),
		interceptor.UntilStatusCodeIs(200),
	).Then(http.DefaultClient).Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	defer req.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status code: %d", res.StatusCode)
	}
}
