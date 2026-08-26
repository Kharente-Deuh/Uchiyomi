// SPDX-License-Identifier: AGPL-3.0-or-later

package http

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kharente-deuh/uchiyomi-server/pkg/sources/kingofshojo/pkg/domain"
)

func TestKosOrder(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		sort  domain.SortType
		order domain.SortOrder
		want  string
	}{
		"popular":    {sort: domain.SortTypePopular, want: string(domain.SortTypePopular)},
		"latest":     {sort: domain.SortTypeLatest, want: "updated"},
		"newest":     {sort: domain.SortTypeNewest, want: "added"},
		"title asc":  {sort: domain.SortTypeTitle, order: domain.SortOrderAsc, want: "title"},
		"title desc": {sort: domain.SortTypeTitle, order: domain.SortOrderDesc, want: "titlereverse"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := kosOrder(tc.sort, tc.order); got != tc.want {
				t.Errorf("kosOrder() = %q, want %q", got, tc.want)
			}
		})
	}
}

func cardHTML(slug, title string) string {
	return `<div class="bsx"><a href="` + mangaPath + slug + `/"><div class="tt">` + title +
		`</div><div class="adds">Chapter 1</div></a></div>`
}

func pageHTML(cards string) string {
	return `<!DOCTYPE html><html><body><div class="listupd">` + cards +
		`</div><div class="hpage">Page 1 of 2</div></body></html>`
}

func TestSearchDoesNotFetchASecondHTMLPage(t *testing.T) {
	t.Parallel()

	var n int

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n++
		_, _ = w.Write([]byte(pageHTML(strings.Repeat(cardHTML("a", "A"), 40))))
	}))
	t.Cleanup(srv.Close)

	c, err := New(Config{BaseURL: srv.URL}, Deps{HTTP: srv.Client(), Logger: discardLogger(), Solver: nil})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	res, err := c.Search(context.Background(), domain.SearchCacheOpts{Page: 1})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	if n != 1 {
		t.Fatalf("requests = %d, want 1", n)
	}

	if len(res.Items) != 40 {
		t.Fatalf("len(items) = %d, want 40", len(res.Items))
	}
}

func discardLogger() *slog.Logger {
	return slog.New(slog.DiscardHandler)
}
