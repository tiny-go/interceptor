package interceptor

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type serveChain chan func(http.ResponseWriter, *http.Request)

func (next serveChain) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler := <-next

	handler(w, r)
}

type readSeekCloser struct {
	*strings.Reader
}

func (readSeekCloser) Close() error {
	return nil
}

func serveWithCode(code int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ioutil.ReadAll(r.Body)

		w.WriteHeader(code)
	}
}

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

	res, err := New(Retry(3), UntilStatusCodeIs(200)).Then(http.DefaultClient).Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	defer req.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status code: %d", res.StatusCode)
	}
}
