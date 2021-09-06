package interceptor

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type callChain chan func(*http.Request) (*http.Response, error)

func (next callChain) Do(req *http.Request) (*http.Response, error) {
	call := <-next

	return call(req)
}

func doAndFail(*http.Request) (*http.Response, error) {
	return nil, errExpected
}

func doReadBodyAndFail(r *http.Request) (*http.Response, error) {
	ioutil.ReadAll(r.Body)

	return nil, errExpected
}

func Test_Retry_ContextCancelled(t *testing.T) {
	calls := make(callChain, 1)
	calls <- http.DefaultClient.Do

	ts := httptest.NewServer(serveWithCode(http.StatusOK))
	defer ts.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ts.URL, http.NoBody)
	if err != nil {
		t.Fatalf("cannot instantiate a client: %s", err)
	}

	if _, err := New(Retry(0)).Then(calls).Do(req); !errors.Is(err, context.Canceled) {
		t.Fatalf("unexpected error: %s", err)
	}
}

func Test_Retry_NoRetry(t *testing.T) {
	ts := httptest.NewServer(serveWithCode(http.StatusOK))
	defer ts.Close()

	req, err := http.NewRequest(http.MethodPost, ts.URL, http.NoBody)
	if err != nil {
		t.Fatalf("cannot instantiate a client: %s", err)
	}

	res, err := New(Log(log.Default()), Retry(0)).Then(http.DefaultClient).Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	defer req.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status code: %d", res.StatusCode)
	}
}

func Test_Retry_NoBody(t *testing.T) {
	calls := make(callChain, 2)
	calls <- doAndFail
	calls <- http.DefaultClient.Do

	ts := httptest.NewServer(serveWithCode(http.StatusOK))
	defer ts.Close()

	req, err := http.NewRequest(http.MethodPost, ts.URL, http.NoBody)
	if err != nil {
		t.Fatalf("cannot instantiate a client: %s", err)
	}

	res, err := New(Log(log.Default()), Retry(3)).Then(calls).Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	defer req.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status code: %d", res.StatusCode)
	}
}

func Test_Retry_Seeker(t *testing.T) {
	calls := make(callChain, 4)
	calls <- doReadBodyAndFail
	calls <- doReadBodyAndFail
	calls <- doReadBodyAndFail
	calls <- http.DefaultClient.Do

	ts := httptest.NewServer(serveWithCode(http.StatusOK))
	defer ts.Close()

	// wrap `*strings.Reader` in order not to pass the type cast check in `http.NewRequest()`.
	body := readSeekCloser{strings.NewReader("foo bar baz")}

	req, err := http.NewRequest(http.MethodPost, ts.URL, body)
	if err != nil {
		t.Fatalf("cannot instantiate a client: %s", err)
	}

	res, err := New(Log(log.Default()), Retry(3)).Then(calls).Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	defer req.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status code: %d", res.StatusCode)
	}
}

func Test_Retry_Buffered(t *testing.T) {
	calls := make(callChain, 4)
	calls <- doReadBodyAndFail
	calls <- doReadBodyAndFail
	calls <- doReadBodyAndFail
	calls <- http.DefaultClient.Do

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer ts.Close()

	body := bytes.NewBufferString("foo bar baz")

	req, err := http.NewRequest(http.MethodPost, ts.URL, body)
	if err != nil {
		t.Fatalf("cannot instantiate a client: %s", err)
	}

	res, err := New(Log(log.Default()), Retry(3)).Then(calls).Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	defer req.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status code: %d", res.StatusCode)
	}
}
