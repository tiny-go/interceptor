package interceptor

import "net/http"

type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

type DoerFunc func(req *http.Request) (*http.Response, error)

func (fn DoerFunc) Do(req *http.Request) (*http.Response, error) {
	return fn(req)
}

type Interceptor func(next Doer) Doer

func New(interceptors ...Interceptor) Interceptor {
	return Interceptor(func(next Doer) Doer {
		return next
	}).Use(interceptors...)
}

func (icp Interceptor) Use(interceptors ...Interceptor) Interceptor {
	for _, next := range interceptors {
		icp = func(curr, next Interceptor) Interceptor {
			return func(doer Doer) Doer {
				return curr(next(doer))
			}
		}(icp, next)
	}

	return icp
}

func (icp Interceptor) Then(doer Doer) Doer {
	return icp(doer)
}
