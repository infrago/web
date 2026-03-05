package web

import (
	"testing"

	. "github.com/infrago/base"
)

func TestContextViewUsesHtmlByDefault(t *testing.T) {
	ctx := &Context{}

	ctx.View("home/index")

	if ctx.Type != "html" {
		t.Fatalf("expected default type html, got %q", ctx.Type)
	}

	body, ok := ctx.Body.(httpViewBody)
	if !ok {
		t.Fatalf("expected httpViewBody, got %T", ctx.Body)
	}
	if body.view != "home/index" {
		t.Fatalf("expected view home/index, got %q", body.view)
	}
	if body.model != nil {
		t.Fatalf("expected nil model by default, got %#v", body.model)
	}
}

func TestContextViewAcceptsCodeMimeAndModel(t *testing.T) {
	ctx := &Context{}
	model := Map{"name": "demo"}

	ctx.View("home/detail", 201, "text/html", model)

	if ctx.Code != 201 {
		t.Fatalf("expected status code 201, got %d", ctx.Code)
	}
	if ctx.Type != "text/html" {
		t.Fatalf("expected type text/html, got %q", ctx.Type)
	}

	body, ok := ctx.Body.(httpViewBody)
	if !ok {
		t.Fatalf("expected httpViewBody, got %T", ctx.Body)
	}
	if body.view != "home/detail" {
		t.Fatalf("expected view home/detail, got %q", body.view)
	}
	if body.model["name"] != "demo" {
		t.Fatalf("expected model.name demo, got %#v", body.model["name"])
	}
}
