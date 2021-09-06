package interceptor

import (
	"bytes"
	"io"
	"net/http"
)

// Retry request interceptor, call `Retry(0)` means make a single call with no retry.
func Retry(tries int) Interceptor {
	return func(next Doer) Doer {
		return DoerFunc(
			func(req *http.Request) (*http.Response, error) {
				seeker, ok := req.Body.(io.Seeker)

				switch {
				case tries == 0:
					// no need to copy the body if there are no retries
					break
				case req.Body == nil || req.Body == http.NoBody:
					// nothing to wrap (no body in the request e.g. GET/OPTIONS)
					break
				case ok:
					// no need to wrap if it implements io.Seeker (e.g. os.File)
					break
				default:
					// read the body into the buffer and
					var buff bytes.Buffer
					if _, err := io.Copy(&buff, req.Body); err != nil {
						return nil, err
					}

					newBody := bytes.NewReader(buff.Bytes())
					seeker, req.Body = newBody, io.NopCloser(newBody)
				}

				// do not count the first try
				for try := 0; ; try++ {
					select {
					case <-req.Context().Done():
						// TODO: send expanded error, merge context error with the latest one
						return nil, req.Context().Err()
					default:
					}

					res, err := next.Do(req)
					if err == nil || try == tries {
						return res, err
					}

					if seeker != nil {
						seeker.Seek(0, io.SeekStart)
					}
				}
			},
		)
	}
}
