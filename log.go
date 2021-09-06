package interceptor

import "net/http"

func Log(logger interface {
	Printf(string, ...interface{})
}) Interceptor {
	return func(next Doer) Doer {
		return DoerFunc(
			func(req *http.Request) (*http.Response, error) {
				logger.Printf("Sending the request to %s\n", req.URL)

				res, err := next.Do(req)
				if err != nil {
					logger.Printf("Error calling %s: %s", req.URL, err)

					return res, err
				}

				logger.Printf("Received the %q response\n", res.Status)

				return res, err
			},
		)
	}
}
