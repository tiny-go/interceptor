package interceptor_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/tiny-go/interceptor"
)

type logItem struct {
	level   string
	message string
}

type logger struct {
	mu   sync.Mutex
	logs []logItem
}

func (l *logger) logf(level, format string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.logs = append(l.logs, logItem{level, fmt.Sprintf(format, args...)})
}

func (l *logger) Debugf(format string, args ...interface{}) {
	l.logf("debug", format, args...)
}

func (l *logger) Infof(format string, args ...interface{}) {
	l.logf("info", format, args...)
}

func (l *logger) Warnf(format string, args ...interface{}) {
	l.logf("warning", format, args...)
}

func (l *logger) Errorf(format string, args ...interface{}) {
	l.logf("error", format, args...)
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

		expectedLogs = []logItem{
			{"info", "Sending [POST] request to " + ts.URL},
			{"debug", "Sending [POST] request to " + ts.URL},
			{"warning", "Error calling [POST] " + ts.URL + ": failure"},
			{"debug", "Sending [POST] request to " + ts.URL},
			{"warning", "Error calling [POST] " + ts.URL + ": failure"},
			{"debug", "Sending [POST] request to " + ts.URL},
			{"warning", "Error calling [POST] " + ts.URL + ": failure"},
			{"debug", "Sending [POST] request to " + ts.URL},
			{"debug", "Received the response: 200 OK"},
			{"info", "Received the response: 200 OK"},
		}
	)

	req, err := http.NewRequest(http.MethodPost, ts.URL, body)
	if err != nil {
		t.Fatalf("cannot instantiate a client: %s", err)
	}

	res, err := interceptor.New(
		// main "error report" logger
		interceptor.Log(logger.Infof, logger.Errorf),
		interceptor.Retry(3),
		// retry "warning" logger
		interceptor.Log(logger.Debugf, logger.Warnf),
	).Then(calls).Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	defer req.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("unexpected status code: %d", res.StatusCode)
	}

	if !reflect.DeepEqual(logger.logs, expectedLogs) {
		t.Errorf(
			"logger debug output:\n%v\ndoes not match expected:\n%v",
			logger.logs,
			expectedLogs,
		)
	}
}
