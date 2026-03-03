package web

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	. "github.com/infrago/base"
	"github.com/infrago/infra"
)

type (
	httpGotoBody struct {
		url string
	}
	httpTextBody struct {
		text string
	}
	httpHtmlBody struct {
		html string
	}
	httpJsonBody struct {
		json Any
	}
	httpJsonpBody struct {
		json     Any
		callback string
	}
	httpEchoBody struct {
		code int
		text string
		data Map
	}
	httpFileBody struct {
		file string
		name string
	}
	httpBinaryBody struct {
		bytes []byte
		name  string
	}
	httpBufferBody struct {
		buffer io.ReadCloser
		size   int64
		name   string
	}
	httpStatusBody string
)

func (site *Site) bodyFail(ctx *Context, err error) {
	if err == nil {
		return
	}

	ctx.Result(infra.Fail.With(err.Error()))

	if ctx.output != nil && ctx.output.Committed() {
		if ctx.Code <= 0 {
			ctx.Code = ctx.output.Status()
		}
		return
	}

	if ctx.failedBody {
		if ctx.Code <= 0 {
			ctx.Code = StatusInternalServerError
		}
		ctx.writer.Header().Set("Content-Type", "text/plain; charset="+ctx.Charset())
		ctx.writer.WriteHeader(ctx.Code)
		_, _ = fmt.Fprint(ctx.writer, StatusText(ctx.Code))
		return
	}

	ctx.failedBody = true
	if ctx.Code <= 0 || ctx.Code < StatusInternalServerError {
		ctx.Code = StatusInternalServerError
	}
	ctx.handling = "error"
	site.handle(ctx)
	site.body(ctx)
}

func (site *Site) body(ctx *Context) {
	if ctx.Code <= 0 {
		ctx.Code = StatusOK
	}

	// Write headers
	for k, v := range ctx.headers {
		ctx.writer.Header().Set(k, v)
	}

	// Write cookies
	for _, cookie := range ctx.cookies {
		cookie.Path = "/"
		cookie.HttpOnly = ctx.site.Config.HttpOnly
		if ctx.site.Config.Domain != "" {
			cookie.Domain = ctx.site.Config.Domain
		}
		if ctx.Domain != "" {
			cookie.Domain = ctx.Domain
		}
		if ctx.site.Config.MaxAge > 0 {
			cookie.MaxAge = int(ctx.site.Config.MaxAge.Seconds())
		}
		http.SetCookie(ctx.writer, &cookie)
	}

	// Issue latest token into configured cookie when ctx.Sign/NewSign is called.
	if ctx.issue && ctx.site.Config.Cookie != "" {
		if token := ctx.Token(); token != "" {
			cookie := http.Cookie{
				Name:     ctx.site.Config.Cookie,
				Value:    token,
				Path:     "/",
				HttpOnly: ctx.site.Config.HttpOnly,
			}
			if ctx.site.Config.Domain != "" {
				cookie.Domain = ctx.site.Config.Domain
			}
			if ctx.Domain != "" {
				cookie.Domain = ctx.Domain
			}
			if ctx.site.Config.MaxAge > 0 {
				cookie.MaxAge = int(ctx.site.Config.MaxAge.Seconds())
			}
			http.SetCookie(ctx.writer, &cookie)
		}
	}

	switch body := ctx.Body.(type) {
	case string:
		site.bodyText(ctx, httpTextBody{body})
	case Map:
		site.bodyJson(ctx, httpJsonBody{body})
	case httpGotoBody:
		site.bodyGoto(ctx, body)
	case httpTextBody:
		site.bodyText(ctx, body)
	case httpHtmlBody:
		site.bodyHtml(ctx, body)
	case httpJsonBody:
		site.bodyJson(ctx, body)
	case httpJsonpBody:
		site.bodyJsonp(ctx, body)
	case httpEchoBody:
		site.bodyEcho(ctx, body)
	case httpFileBody:
		site.bodyFile(ctx, body)
	case httpBinaryBody:
		site.bodyBinary(ctx, body)
	case httpBufferBody:
		site.bodyBuffer(ctx, body)
	case httpStatusBody:
		site.bodyStatus(ctx, body)
	default:
		site.bodyDefault(ctx)
	}
}

func (site *Site) bodyDefault(ctx *Context) {
	if ctx.Code <= 0 {
		ctx.Code = StatusNotFound
		http.NotFound(ctx.writer, ctx.reader)
	} else {
		ctx.writer.WriteHeader(ctx.Code)
		fmt.Fprint(ctx.writer, StatusText(ctx.Code))
	}
}

func (site *Site) bodyStatus(ctx *Context, body httpStatusBody) {
	if ctx.Code <= 0 {
		ctx.Code = StatusNotFound
		http.NotFound(ctx.writer, ctx.reader)
	} else {
		if body == "" {
			body = httpStatusBody(StatusText(ctx.Code))
		}
		ctx.writer.WriteHeader(ctx.Code)
		fmt.Fprint(ctx.writer, body)
	}
}

func (site *Site) bodyGoto(ctx *Context, body httpGotoBody) {
	if ctx.Code <= 0 {
		ctx.Code = StatusFound
	}
	http.Redirect(ctx.writer, ctx.reader, body.url, StatusFound)
}

func (site *Site) bodyText(ctx *Context, body httpTextBody) {
	res := ctx.writer

	if ctx.Type == "" {
		ctx.Type = "text"
	}

	mimeType := infra.Mimetype(ctx.Type, "text/plain")
	res.Header().Set("Content-Type", fmt.Sprintf("%v; charset=%v", mimeType, ctx.Charset()))

	res.WriteHeader(ctx.Code)
	if _, err := fmt.Fprint(res, body.text); err != nil {
		site.bodyFail(ctx, err)
	}
}

func (site *Site) bodyHtml(ctx *Context, body httpHtmlBody) {
	res := ctx.writer

	if ctx.Type == "" {
		ctx.Type = "html"
	}

	mimeType := infra.Mimetype(ctx.Type, "text/html")
	res.Header().Set("Content-Type", fmt.Sprintf("%v; charset=%v", mimeType, ctx.Charset()))

	res.WriteHeader(ctx.Code)
	if _, err := fmt.Fprint(res, body.html); err != nil {
		site.bodyFail(ctx, err)
	}
}

func (site *Site) bodyJson(ctx *Context, body httpJsonBody) {
	res := ctx.writer

	if ctx.Type == "" {
		ctx.Type = "json"
	}

	bytes, err := json.Marshal(body.json)
	if err != nil {
		site.bodyFail(ctx, err)
		return
	}

	mimeType := infra.Mimetype(ctx.Type, "application/json")
	res.Header().Set("Content-Type", fmt.Sprintf("%v; charset=%v", mimeType, ctx.Charset()))
	res.WriteHeader(ctx.Code)
	if _, err := fmt.Fprint(res, string(bytes)); err != nil {
		site.bodyFail(ctx, err)
	}
}

func (site *Site) bodyJsonp(ctx *Context, body httpJsonpBody) {
	res := ctx.writer

	if ctx.Type == "" {
		ctx.Type = "script"
	}

	bytes, err := json.Marshal(body.json)
	if err != nil {
		site.bodyFail(ctx, err)
		return
	}

	mimeType := infra.Mimetype(ctx.Type, "application/javascript")
	res.Header().Set("Content-Type", fmt.Sprintf("%v; charset=%v", mimeType, ctx.Charset()))

	res.WriteHeader(ctx.Code)
	if _, err := fmt.Fprintf(res, "%s(%s);", body.callback, string(bytes)); err != nil {
		site.bodyFail(ctx, err)
	}
}

func (site *Site) bodyEcho(ctx *Context, body httpEchoBody) {
	result := Map{
		"code": body.code,
		"time": time.Now().Unix(),
	}

	if body.text != "" {
		result["text"] = body.text
	}

	if body.data != nil {
		result["data"] = body.data
	}

	site.bodyJson(ctx, httpJsonBody{result})
}

func (site *Site) bodyFile(ctx *Context, body httpFileBody) {
	req, res := ctx.reader, ctx.writer

	if ctx.Type == "" {
		ctx.Type = "file"
	}

	mimeType := infra.Mimetype(ctx.Type, "application/octet-stream")
	res.Header().Set("Content-Type", fmt.Sprintf("%v; charset=%v", mimeType, ctx.Charset()))

	if body.name != "" {
		res.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%v;", url.QueryEscape(body.name)))
	}

	http.ServeFile(res, req, body.file)
}

func (site *Site) bodyBinary(ctx *Context, body httpBinaryBody) {
	res := ctx.writer

	if ctx.Type == "" {
		ctx.Type = "file"
	}

	mimeType := infra.Mimetype(ctx.Type, "application/octet-stream")
	res.Header().Set("Content-Type", fmt.Sprintf("%v; charset=%v", mimeType, ctx.Charset()))

	if body.name != "" {
		res.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%v;", url.QueryEscape(body.name)))
	}

	res.WriteHeader(ctx.Code)
	if _, err := res.Write(body.bytes); err != nil {
		site.bodyFail(ctx, err)
	}
}

func (site *Site) bodyBuffer(ctx *Context, body httpBufferBody) {
	res := ctx.writer

	if ctx.Type == "" {
		ctx.Type = "file"
	}

	mimeType := infra.Mimetype(ctx.Type, "application/octet-stream")
	res.Header().Set("Content-Type", fmt.Sprintf("%v; charset=%v", mimeType, ctx.Charset()))

	if body.name != "" {
		res.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%v;", url.QueryEscape(body.name)))
	}

	if body.size > 0 {
		res.Header().Set("Content-Length", fmt.Sprintf("%d", body.size))
	}

	res.WriteHeader(ctx.Code)
	if _, err := io.Copy(res, body.buffer); err != nil {
		site.bodyFail(ctx, err)
	}
	body.buffer.Close()
}
