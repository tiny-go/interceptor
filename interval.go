package interceptor

import (
	"net/http"
	"time"
)

func WithInterval(duration time.Duration) Interceptor {
	return func(next Doer) Doer {
		return DoerFunc(
			func(req *http.Request) (*http.Response, error) {
				res, err := next.Do(req)
				if err != nil {
					timer := time.NewTimer(duration)
					defer timer.Stop()

					select {
					case <-req.Context().Done():
						return nil, err
					case <-timer.C:
						return res, err
					}
				}

				return res, nil
			},
		)
	}
}
