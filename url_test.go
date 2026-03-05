package web

import (
	"net/http"
	"testing"

	"github.com/infrago/infra"
)

func TestFindRouteInfoFallback(t *testing.T) {
	site := &webSite{
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

	module.sites = map[string]*webSite{
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
		site: &webSite{Name: "www"},
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

func TestRouteUriAndRouteUrl(t *testing.T) {
	oldSites := module.sites
	oldAliases := module.siteAliases
	oldConfig := module.config
	defer func() {
		module.sites = oldSites
		module.siteAliases = oldAliases
		module.config = oldConfig
	}()

	module.sites = map[string]*webSite{
		infra.DEFAULT: {Name: infra.DEFAULT},
		"www": {
			Name: "www",
			Config: Config{
				Domain: "www.demo.local",
				Port:   8090,
			},
			routerInfos: map[string]Info{
				"home.*": {Router: "home.*", Uri: "/"},
			},
			routerOrder: []string{"home.*"},
		},
	}
	module.siteAliases = map[string]string{
		"www": "www",
	}
	module.config = Config{Port: 8090}

	if got := RouteUri("www.home"); got != "/" {
		t.Fatalf("expected route uri /, got %s", got)
	}

	if got := RouteUrl("www.home"); got != "http://www.demo.local:8090/" {
		t.Fatalf("expected route url http://www.demo.local:8090/, got %s", got)
	}
}

func TestContextRouteMethods(t *testing.T) {
	oldSites := module.sites
	oldAliases := module.siteAliases
	oldConfig := module.config
	defer func() {
		module.sites = oldSites
		module.siteAliases = oldAliases
		module.config = oldConfig
	}()

	module.sites = map[string]*webSite{
		"www": {
			Name: "www",
			Config: Config{
				Port: 8090,
			},
			routerInfos: map[string]Info{
				"home.*": {Router: "home.*", Uri: "/home"},
			},
			routerOrder: []string{"home.*"},
		},
		infra.DEFAULT: {Name: infra.DEFAULT},
	}
	module.siteAliases = map[string]string{
		"www": "www",
	}
	module.config = Config{Port: 8090}

	ctx := &Context{
		site: &webSite{Name: "www"},
		Name: "home",
		Host: "www.example.org",
	}

	if got := ctx.RouteUri("home"); got != "/home" {
		t.Fatalf("expected context route uri /home, got %s", got)
	}
	if got := ctx.RouteUrl("home"); got != "http://www.example.org:8090/home" {
		t.Fatalf("expected context route url http://www.example.org:8090/home, got %s", got)
	}
	if got := ctx.SiteUrl("www", "/home"); got != "http://www.example.org:8090/home" {
		t.Fatalf("expected context site url http://www.example.org:8090/home, got %s", got)
	}
}
