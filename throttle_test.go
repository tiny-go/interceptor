package interceptor_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/tiny-go/interceptor"
)

func Test_Throttle(t *testing.T) {
	handler := make(serveChain, 3)
	handler <- serveWithCode(http.StatusOK)
	handler <- serveWithCode(http.StatusOK)
	handler <- serveWithCode(http.StatusOK)

	ts := httptest.NewServer(handler)
	defer ts.Close()

	req, err := http.NewRequest(http.MethodPost, ts.URL, http.NoBody)
	if err != nil {
		t.Fatalf("cannot instantiate a client: %s", err)
	}

	defer req.Body.Close()

	tick := make(chan time.Time)

	icp := interceptor.New(
		interceptor.Throttle(tick),
	).Then(http.DefaultClient)

	go func() {
		tick <- time.Now()
		tick <- time.Now()
		tick <- time.Now()
	}()

	icp.Do(req)
	icp.Do(req)
	icp.Do(req)
}
