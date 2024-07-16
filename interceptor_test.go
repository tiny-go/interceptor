package interceptor_test

import (
	"errors"
	"io"
	"net/http"
	"strings"
)

var errExpected = errors.New("failure")

func doAndFail(*http.Request) (*http.Response, error) {
	return nil, errExpected
}

func doReadBodyAndFail(r *http.Request) (*http.Response, error) {
	io.ReadAll(r.Body)

	return nil, errExpected
}

type readSeekCloser struct {
	*strings.Reader
}

func (readSeekCloser) Close() error {
	return nil
}

func serveWithCode(code int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		io.ReadAll(r.Body)

		w.WriteHeader(code)
	}
}

type callChain chan func(*http.Request) (*http.Response, error)

func (next callChain) Do(req *http.Request) (*http.Response, error) {
	return (<-next)(req)
}

type serveChain chan func(http.ResponseWriter, *http.Request)

func (next serveChain) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	(<-next)(w, r)
}
