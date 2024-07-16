package interceptor

import "net/http"

func Log(
	debugfFunc, errorfFunc func(string, ...interface{}),
) Interceptor {
	return func(next Doer) Doer {
		return DoerFunc(
			func(req *http.Request) (*http.Response, error) {
				debugfFunc("Sending [%s] request to %s", req.Method, req.URL)

				res, err := next.Do(req)
				if err != nil {
					errorfFunc("Error calling [%s] %s: %s", req.Method, req.URL, err)

					return res, err
				}

				debugfFunc("Received the response: %s", res.Status)

				return res, err
			},
		)
	}
}
