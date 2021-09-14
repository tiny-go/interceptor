package interceptor

import "net/http"

func Log(logger interface {
	Debugf(string, ...interface{})
	Errorf(string, ...interface{})
}) Interceptor {
	return func(next Doer) Doer {
		return DoerFunc(
			func(req *http.Request) (*http.Response, error) {
				logger.Debugf("Sending the request: [%s] %s", req.Method, req.URL)

				res, err := next.Do(req)
				if err != nil {
					logger.Errorf("Error calling [%s] %s: %s", req.Method, req.URL, err)

					return res, err
				}

				logger.Debugf("Received the response: %s", res.Status)

				return res, err
			},
		)
	}
}
