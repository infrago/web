package web

import (
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/infrago/infra"
	. "github.com/infrago/base"
)

type webUrl struct {
	ctx *Context
}

func (m *Module) url() *webUrl {
	return &webUrl{}
}

// Routo forces site base url.
func (u *webUrl) Routo(name string, values ...Map) string {
	vals := Map{}
	if len(values) > 0 {
		vals = values[0]
	}
	vals["[site]"] = true
	return u.Route(name, vals)
}

// Route builds url by route name.
func (u *webUrl) Route(name string, values ...Map) string {
	name = strings.ToLower(name)
	if strings.HasPrefix(name, "http://") || strings.HasPrefix(name, "https://") ||
		strings.HasPrefix(name, "ws://") || strings.HasPrefix(name, "wss://") {
		return name
	}

	currSite := ""
	if u.ctx != nil && u.ctx.site != nil {
		currSite = u.ctx.site.Name
		if name == "" {
			name = u.ctx.Name
		}
	}

	if strings.Contains(name, ".") == false {
		if currSite != "" {
			name = currSite + "." + name
		} else {
			name = infra.DEFAULT + "." + name
		}
	}

	params, querys, options := Map{}, Map{}, Map{}
	if len(values) > 0 {
		for k, v := range values[0] {
			if strings.HasPrefix(k, "{") && strings.HasSuffix(k, "}") {
				params[k] = v
			} else if strings.HasPrefix(k, "[") && strings.HasSuffix(k, "]") {
				options[k] = v
			} else {
				querys[k] = v
			}
		}
	}

	siteName, routeName := splitPrefix(name)
	if siteName == "*" {
		if currSite != "" {
			siteName = currSite
		} else {
			for s := range module.sites {
				siteName = s
				break
			}
		}
	}

	if siteName != "" && siteName != currSite {
		options["[site]"] = siteName
	} else if options["[site]"] != nil && currSite != "" {
		options["[site]"] = currSite
	}

	site := module.sites[siteName]
	if site == nil {
		site = module.sites[infra.DEFAULT]
	}
	if site == nil {
		return name
	}

	info, ok := site.routerInfos[routeName]
	if !ok {
		// try method or index variants
		if v, ok := site.routerInfos[routeName+".get.0"]; ok {
			info = v
			ok = true
		} else if v, ok := site.routerInfos[routeName+".post.0"]; ok {
			info = v
			ok = true
		} else if v, ok := site.routerInfos[routeName+".*.0"]; ok {
			info = v
			ok = true
		}
	}
	if !ok {
		return name
	}

	argsConfig := Vars{}
	if info.Args != nil {
		for k, v := range info.Args {
			argsConfig[k] = v
		}
	}

	dataArgsValues, dataParseValues := Map{}, Map{}
	for k, v := range params {
		if strings.HasPrefix(k, "{") && strings.HasSuffix(k, "}") {
			kk := strings.TrimSuffix(strings.TrimPrefix(k, "{"), "}")
			dataArgsValues[kk] = v
		} else {
			dataArgsValues[k] = v
			querys[k] = v
		}
	}

	zone := time.Local
	if u.ctx != nil && u.ctx.Meta != nil {
		zone = u.ctx.Meta.Timezone()
	}

	_ = infra.Mapping(argsConfig, dataArgsValues, dataParseValues, false, true, zone)

	// merge parsed values
	dataValues := Map{}
	for k, v := range dataParseValues {
		dataValues[k] = v
	}

	uri := info.Uri
	re := regexp.MustCompile(`\{[^}]+\}`)
	uri = re.ReplaceAllStringFunc(uri, func(m string) string {
		key := strings.TrimSuffix(strings.TrimPrefix(m, "{"), "}")
		if v, ok := dataValues[key]; ok {
			return fmt.Sprintf("%v", v)
		}
		if v, ok := params[m]; ok {
			return fmt.Sprintf("%v", v)
		}
		return ""
	})

	if len(querys) > 0 {
		q := url.Values{}
		for k, v := range querys {
			q.Set(k, fmt.Sprintf("%v", v))
		}
		if strings.Contains(uri, "?") {
			uri = uri + "&" + q.Encode()
		} else {
			uri = uri + "?" + q.Encode()
		}
	}

	if siteOpt, ok := options["[site]"]; ok && siteOpt != nil {
		siteName := siteName
		if s, ok := siteOpt.(string); ok && s != "" {
			siteName = s
		}
		return u.Site(siteName, uri, options)
	}

	return uri
}

// Site builds site base url with path.
func (u *webUrl) Site(name string, path string, options ...Map) string {
	opts := Map{}
	if len(options) > 0 {
		opts = options[0]
	}

	site := module.sites[name]
	if site == nil {
		site = module.sites[infra.DEFAULT]
	}
	if site == nil {
		return path
	}

	host := u.resolveSiteHost(name, site)

	port := module.config.Port
	if port <= 0 {
		port = site.Config.Port
	}
	if !strings.Contains(host, ":") && port > 0 {
		if port != 80 && port != 443 {
			host = fmt.Sprintf("%s:%d", host, port)
		}
	}

	socket := false
	ssl := false
	if v, ok := opts["[socket]"].(bool); ok && v {
		socket = true
	}
	if v, ok := opts["[ssl]"].(bool); ok && v {
		ssl = true
	}

	scheme := "http://"
	if socket {
		scheme = "ws://"
	}
	if ssl {
		if socket {
			scheme = "wss://"
		} else {
			scheme = "https://"
		}
	}

	if path == "" {
		return scheme + host
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return scheme + host + path
}

func (u *webUrl) resolveSiteHost(name string, site *Site) string {
	// 1) explicit site domain first
	if site.Config.Domain != "" {
		return normalizeHost(site.Config.Domain)
	}
	if len(site.Config.Domains) > 0 && site.Config.Domains[0] != "" {
		return normalizeHost(site.Config.Domains[0])
	}

	// 2) global web domain fallback => <site>.<web.domain>
	if module.config.Domain != "" {
		base := normalizeHost(module.config.Domain)
		if base != "" {
			return normalizeHost(name + "." + base)
		}
	}

	// 3) no domain configured: use current host's main domain tail
	//    e.g. current www.dev.com + target file => file.dev.com
	if u.ctx != nil && u.ctx.Host != "" {
		curr := normalizeHost(u.ctx.Host)
		if tail := hostTail(curr); tail != "" {
			return normalizeHost(name + "." + tail)
		}
	}

	// 4) final fallback
	if site.Config.Host != "" {
		return normalizeHost(site.Config.Host)
	}
	return "localhost"
}

func hostTail(host string) string {
	host = normalizeHost(host)
	if host == "" {
		return ""
	}
	if ip := net.ParseIP(host); ip != nil {
		return ""
	}
	parts := strings.Split(host, ".")
	if len(parts) < 2 {
		return ""
	}
	return strings.Join(parts[1:], ".")
}

// RouteUrl shortcut
func RouteUrl(name string, values ...Map) string {
	return module.url().Route(name, values...)
}

// SiteUrl shortcut
func SiteUrl(name, path string, options ...Map) string {
	return module.url().Site(name, path, options...)
}
