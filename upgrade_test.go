package web

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/infrago/infra"
)

type testUpgradeSocket struct{}

func (testUpgradeSocket) ReadMessage() (int, []byte, error) { return 0, nil, io.EOF }
func (testUpgradeSocket) WriteMessage(int, []byte) error    { return nil }
func (testUpgradeSocket) Close() error                      { return nil }
func (testUpgradeSocket) Raw() any                          { return nil }

type testUpgradeConnect struct {
	socket   Socket
	upgraded bool
}

func (c *testUpgradeConnect) Open() error                           { return nil }
func (c *testUpgradeConnect) Close() error                          { return nil }
func (c *testUpgradeConnect) Register(string, Info, []string) error { return nil }
func (c *testUpgradeConnect) Start() error                          { return nil }
func (c *testUpgradeConnect) StartTLS(string, string) error         { return nil }
func (c *testUpgradeConnect) Upgrade(http.ResponseWriter, *http.Request) (Socket, error) {
	c.upgraded = true
	return c.socket, nil
}

func TestContextUpgradeUsesDefaultEndpoint(t *testing.T) {
	oldEndpoints := module.endpoints
	oldInstance := module.instance
	defer func() {
		module.endpoints = oldEndpoints
		module.instance = oldInstance
	}()

	socket := testUpgradeSocket{}
	connect := &testUpgradeConnect{socket: socket}
	called := false
	module.endpoints = map[string]Endpoint{
		infra.DEFAULT: {
			Accept: func(ctx *Context, got Socket) error {
				called = true
				if got != socket {
					t.Fatalf("unexpected socket: %#v", got)
				}
				return nil
			},
		},
	}
	module.instance = &Instance{connect: connect}

	ctx := &Context{
		Meta:   infra.NewMeta(),
		site:   &webSite{Config: Config{Driver: DEFAULT}},
		writer: httptest.NewRecorder(),
		reader: httptest.NewRequest(http.MethodGet, "/socket", nil),
	}

	if err := ctx.Upgrade(); err != nil {
		t.Fatalf("upgrade failed: %v", err)
	}
	if !called || !connect.upgraded {
		t.Fatalf("expected default endpoint accept to be called")
	}
	if !ctx.upgraded || ctx.Code != StatusSwitchingProtocols {
		t.Fatalf("expected websocket upgrade state, got upgraded=%v code=%d", ctx.upgraded, ctx.Code)
	}
}

func TestContextUpgradeUsesNamedEndpoint(t *testing.T) {
	oldEndpoints := module.endpoints
	oldInstance := module.instance
	defer func() {
		module.endpoints = oldEndpoints
		module.instance = oldInstance
	}()

	module.endpoints = map[string]Endpoint{
		infra.DEFAULT: {Accept: func(*Context, Socket) error { t.Fatalf("default endpoint should not be used"); return nil }},
		"custom":      {Accept: func(*Context, Socket) error { return nil }},
	}
	module.instance = &Instance{connect: &testUpgradeConnect{socket: testUpgradeSocket{}}}

	ctx := &Context{
		Meta:   infra.NewMeta(),
		site:   &webSite{Config: Config{Driver: DEFAULT}},
		writer: httptest.NewRecorder(),
		reader: httptest.NewRequest(http.MethodGet, "/socket", nil),
	}

	if err := ctx.Upgrade("custom"); err != nil {
		t.Fatalf("upgrade failed: %v", err)
	}
}

func TestContextUpgradeUsesInfraAcceptor(t *testing.T) {
	oldEndpoints := module.endpoints
	oldInstance := module.instance
	defer func() {
		module.endpoints = oldEndpoints
		module.instance = oldInstance
	}()

	module.endpoints = map[string]Endpoint{}
	module.instance = &Instance{connect: &testUpgradeConnect{socket: testUpgradeSocket{}}}
	infra.RegisterUpgradeAcceptor("web_test_accept", func(opts infra.UpgradeAcceptOptions) error {
		if opts.Socket == nil || opts.Name != "demo.socket" {
			t.Fatalf("unexpected upgrade options: %#v", opts)
		}
		return nil
	})

	ctx := &Context{
		Meta:   infra.NewMeta(),
		Name:   "demo.socket",
		site:   &webSite{Config: Config{Driver: DEFAULT}},
		writer: httptest.NewRecorder(),
		reader: httptest.NewRequest(http.MethodGet, "/socket", nil),
	}

	if err := ctx.Upgrade("web_test_accept"); err != nil {
		t.Fatalf("upgrade failed: %v", err)
	}
}
