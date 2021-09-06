package interceptor

import (
	"fmt"
	"io"
	"net/http"
)

func UntilStatusCodeIs(status int) Interceptor {
	return func(next Doer) Doer {
		return DoerFunc(
			func(req *http.Request) (*http.Response, error) {
				res, err := next.Do(req)
				if err != nil {
					return nil, err
				}

				if res.StatusCode != status {
					io.Copy(io.Discard, res.Body)
					res.Body.Close()

					return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
				}

				return res, nil
			},
		)
	}
}
