package web

import (
	"net/http"
	"testing"

	"github.com/infrago/infra"
)

func TestFindRouteInfoFallback(t *testing.T) {
	site := &Site{
		routerInfos: map[string]Info{
			"home.*":     {Router: "home.*", Uri: "/"},
			"user.get":   {Router: "user.get", Uri: "/user/{id}"},
			"user.get.1": {Router: "user.get", Uri: "/users/{id}"},
		},
		routerOrder: []string{"home.*", "user.get", "user.get.1"},
	}

	if info, ok := findRouteInfo(site, "home"); !ok || info.Router != "home.*" {
		t.Fatalf("expected home.* fallback, got ok=%v router=%s", ok, info.Router)
	}

	if info, ok := findRouteInfo(site, "user"); !ok || info.Router != "user.get" {
		t.Fatalf("expected user.get fallback, got ok=%v router=%s", ok, info.Router)
	}

	if _, ok := findRouteInfo(site, "missing"); ok {
		t.Fatalf("expected missing route to fail")
	}
}

func TestResolveSiteHostPrefersCurrentDomainTail(t *testing.T) {
	oldSites := module.sites
	oldAliases := module.siteAliases
	oldConfig := module.config
	defer func() {
		module.sites = oldSites
		module.siteAliases = oldAliases
		module.config = oldConfig
	}()

	module.sites = map[string]*Site{
		infra.DEFAULT: {Name: infra.DEFAULT},
		"api": {
			Name:   "api",
			Config: Config{Domain: "api.config.com"},
		},
	}
	module.siteAliases = map[string]string{
		"api": "api",
	}
	module.config = Config{Port: 8080}

	req, _ := http.NewRequest(http.MethodGet, "http://www.example.org/ping", nil)
	req.Host = "www.example.org:8080"
	ctx := &Context{
		site: &Site{Name: "www"},
		Host: "www.example.org",
	}
	ctx.reader = req

	u := webUrl{ctx: ctx}
	got := u.Site("api", "/ping")
	want := "http://api.example.org:8080/ping"
	if got != want {
		t.Fatalf("expected %s, got %s", want, got)
	}
}
