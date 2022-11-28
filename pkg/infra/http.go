package infra

import (
	"net/http"

	httptrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/net/http"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
)

func HTTPTracedTransport(rt http.RoundTripper, serviceName string) http.RoundTripper {
	return httptrace.WrapRoundTripper(rt, []httptrace.RoundTripperOption{
		httptrace.WithBefore(func(req *http.Request, span ddtrace.Span) {
			span.SetTag(ext.ServiceName, serviceName)
			span.SetTag(ext.SpanType, ext.SpanTypeHTTP)
			span.SetTag(ext.HTTPMethod, req.Method)
			span.SetTag(ext.HTTPURL, req.URL.Path)
		}),
	}...)
}
