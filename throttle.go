package interceptor

import (
	"net/http"
	"time"
)

// Throttle provides an interceptor that delays client request waiting for a value
// from an input channel.
// Use `time.NewTicker(duration).C` or `time.Tick(duration)` as an input parameter.
//
// Example:
//
//	var (
//		ticker = time.NewTicker(duration)
//		icp = interceptor.New(interceptor.Throttle(ticker.C)).Then(http.DefaultClient)
//	)
//
//	icp.Do(req)
//	icp.Do(req)
//	. . .
//	icp.Do(req)
func Throttle(tick <-chan time.Time) Interceptor {
	return func(next Doer) Doer {
		return DoerFunc(
			func(req *http.Request) (*http.Response, error) {
				<-tick

				return next.Do(req)
			},
		)
	}
}
