package web

import (
	"errors"

	. "github.com/infrago/base"
	"github.com/infrago/ws"
)

func (ctx *Context) Upgrade() error {
	if ctx == nil {
		return errors.New("invalid web context")
	}
	if ctx.upgraded {
		return nil
	}
	if ctx.site == nil || ctx.site.Config.Driver == "" || module.instance == nil || module.instance.connect == nil {
		return errors.New("invalid web connection")
	}

	conn, err := module.instance.connect.Upgrade(ctx.writer, ctx.reader)
	if err != nil {
		return err
	}

	ctx.upgraded = true
	ctx.Code = StatusSwitchingProtocols
	ctx.clearBody()
	ctx.Body = nil

	return ws.Accept(ws.AcceptOptions{
		Conn:       conn,
		Meta:       ctx.Meta,
		Name:       ctx.Name,
		Site:       ctx.Site,
		Host:       ctx.Host,
		Domain:     ctx.Domain,
		RootDomain: ctx.RootDomain,
		Path:       ctx.Path,
		Uri:        ctx.Uri,
		Setting:    cloneContextMap(ctx.Setting),
		Params:     cloneContextMap(ctx.Params),
		Query:      cloneContextMap(ctx.Query),
		Form:       cloneContextMap(ctx.Form),
		Value:      cloneContextMap(ctx.Value),
		Args:       cloneContextMap(ctx.Args),
		Locals:     cloneContextMap(ctx.Locals),
	})
}

func cloneContextMap(src Map) Map {
	if len(src) == 0 {
		return Map{}
	}

	dst := make(Map, len(src))
	for key, value := range src {
		dst[key] = value
	}
	return dst
}
