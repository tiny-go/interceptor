package interceptor

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"testing"
)

type logger struct {
	mu sync.Mutex

	debugs []string
	errors []string
}

func (l *logger) Debugf(format string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.debugs = append(l.debugs, fmt.Sprintf(format, args...))
}

func (l *logger) Errorf(format string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.errors = append(l.errors, fmt.Sprintf(format, args...))
}

func Test_Log_Retry(t *testing.T) {
	calls := make(callChain, 4)
	calls <- doReadBodyAndFail
	calls <- doReadBodyAndFail
	calls <- doReadBodyAndFail
	calls <- http.DefaultClient.Do

	ts := httptest.NewServer(serveWithCode(http.StatusOK))
	defer ts.Close()

	var (
		// wrap `*strings.Reader` in order not to pass the type cast check in `http.NewRequest()`.
		body = readSeekCloser{strings.NewReader("foo bar baz")}

		logger = new(logger)

		expectedDebugLogs = []string{
			"Sending the request: [POST] " + ts.URL,
			"Sending the request: [POST] " + ts.URL,
			"Sending the request: [POST] " + ts.URL,
			"Sending the request: [POST] " + ts.URL,
			"Received the response: 200 OK",
		}

		expectedErrorLogs = []string{
			"Error calling [POST] " + ts.URL + ": failure",
			"Error calling [POST] " + ts.URL + ": failure",
			"Error calling [POST] " + ts.URL + ": failure",
		}
	)

	req, err := http.NewRequest(http.MethodPost, ts.URL, body)
	if err != nil {
		t.Fatalf("cannot instantiate a client: %s", err)
	}

	res, err := New(Retry(3), Log(logger)).Then(calls).Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	defer req.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("unexpected status code: %d", res.StatusCode)
	}

	if !reflect.DeepEqual(logger.debugs, expectedDebugLogs) {
		t.Errorf(
			"logger debug output:\n%s\ndoes not match expected:\n%s",
			strings.Join(logger.debugs, "\n"),
			strings.Join(expectedDebugLogs, "\n"),
		)
	}

	if !reflect.DeepEqual(logger.errors, expectedErrorLogs) {
		t.Errorf(
			"logger error output:\n%s\ndoes not match expected:\n%s",
			strings.Join(logger.errors, "\n"),
			strings.Join(expectedDebugLogs, "\n"),
		)
	}
}
