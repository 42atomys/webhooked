package fasthttpz

import "github.com/valyala/fasthttp"

type RequestCtx struct {
	*fasthttp.RequestCtx
}

func (r *RequestCtx) TemplateContext() map[string]any {
	return map[string]any{
		"ConnID":      r.ConnID(),
		"ConnTime":    r.ConnTime(),
		"Host":        string(r.Host()),
		"IsTLS":       r.IsTLS(),
		"Method":      string(r.Method()),
		"QueryArgs":   r.QueryArgs(),
		"RemoteAddr":  r.RemoteAddr(),
		"RemoteIP":    r.RemoteIP(),
		"RequestTime": r.Time(),
		"URI":         r.URI(),
		"UserAgent":   string(r.UserAgent()),
		"Request":     &r.Request,
		"Payload":     string(r.Request.Body()),
	}
}
