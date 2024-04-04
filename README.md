# interceptor
Simple HTTP request interceptor

### The concept
- define the interceptor to use for every outgoing request
- create a request
- provide an HTTP client
- send the request with interceptor

### Examples

- retry while destination service is unavailable
```go
interceptor.New(
    interceptor.Retry(serviceConfig.Client.MaxRetries),
    interceptor.WithInterval(time.Duration(serviceConfig.Client.RetryInterval)),
    interceptor.WhileFails(func(r *http.Response) error {
        if r.StatusCode == 503 {
            return errors.ServiceUnavailable("service is unavailable")
        }

        return nil
    }),
).Then(http.DefaultClient).Do(req)
```
- throttle
```go
var (
    ticker = time.NewTicker(duration)
    icp = interceptor.New(interceptor.Throttle(ticker.C)).Then(http.DefaultClient)
)

icp.Do(req)
icp.Do(req)
. . .
icp.Do(req)
```
