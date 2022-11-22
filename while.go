package interceptor

import (
	"io"
	"net/http"
)

// WhileFails is one of the configurable parts of retry strategy.
// Specify the function to check the response and return descriptive error
// message if needed.
func WhileFails(checkFunc func(*http.Response) error) Interceptor {
	return func(next Doer) Doer {
		return DoerFunc(
			func(req *http.Request) (*http.Response, error) {
				res, err := next.Do(req)
				if err != nil {
					return nil, err
				}

				if err := checkFunc(res); err != nil {
					io.Copy(io.Discard, res.Body)
					res.Body.Close()

					return nil, err
				}

				return res, nil
			},
		)
	}
}
