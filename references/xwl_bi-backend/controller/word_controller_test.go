package controller

import (
	"errors"
	"github.com/1340691923/xwl_bi/engine/logs"
	"github.com/PuerkitoBio/goquery"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
	"strings"
	"testing"
)

func TestGetWordParseReturnsErrorWhenLookupFails(t *testing.T) {
	oldLoadWordDocument := loadWordDocument
	defer func() {
		loadWordDocument = oldLoadWordDocument
	}()

	loadWordDocument = func(word string) (*goquery.Document, error) {
		return nil, errors.New("lookup failed")
	}
	logs.Logger = zap.NewNop()

	var req fasthttp.Request
	req.Header.SetMethod("GET")
	req.URI().QueryArgs().Add("word", "hello")

	var ctx fasthttp.RequestCtx
	ctx.Init(&req, nil, nil)

	GetWordParse(&ctx)

	body := string(ctx.Response.Body())
	if !strings.Contains(body, `"code":500`) {
		t.Fatalf("expected error response, got %s", body)
	}
	if !strings.Contains(body, `"msg":"服务异常"`) {
		t.Fatalf("expected service error message, got %s", body)
	}
}
