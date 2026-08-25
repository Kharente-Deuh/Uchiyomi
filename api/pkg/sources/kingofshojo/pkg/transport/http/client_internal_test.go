// SPDX-License-Identifier: AGPL-3.0-or-later

package http

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kharente-deuh/uchiyomi-server/pkg/sources/kingofshojo/pkg/domain"
)

func TestKosPage(t *testing.T) {
	t.Parallel()

	page, start := kosPage(0, 20)
	if page != 1 || start != 0 {
		t.Fatalf("offset 0: page=%d start=%d", page, start)
	}

	page, start = kosPage(40, 20)
	if page != 2 || start != 0 {
		t.Fatalf("offset 40: page=%d start=%d", page, start)
	}

	page, start = kosPage(38, 5)
	if page != 1 || start != 38 {
		t.Fatalf("offset 38: page=%d start=%d", page, start)
	}
}

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
	return `<div class="bsx"><a href="/manga/` + slug + `/"><div class="tt">` + title +
		`</div><div class="adds">Chapter 1</div></a></div>`
}

func pageHTML(cards string) string {
	return `<!DOCTYPE html><html><body><div class="listupd">` + cards +
		`</div><div class="hpage">Page 1 of 2</div></body></html>`
}

func TestSearchSpansTwoKoSPages(t *testing.T) {
	t.Parallel()

	page1Cards := strings.Repeat(cardHTML("a", "A"), 40)
	page2Cards := cardHTML("b", "B") + cardHTML("c", "C") + cardHTML("last", "Last")

	var gotPaths []string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPaths = append(gotPaths, r.URL.Path)

		switch r.URL.Path {
		case "/manga/":
			_, _ = w.Write([]byte(pageHTML(page1Cards)))
		case "/manga/page/2/":
			_, _ = w.Write([]byte(pageHTML(page2Cards)))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	c, err := New(
		Config{BaseURL: srv.URL},
		Deps{HTTP: srv.Client(), Logger: discardLogger(), Solver: nil},
	)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	res, err := c.Search(context.Background(), domain.SearchCacheOpts{Offset: 38, Limit: 5})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	if len(res.Items) != 5 {
		t.Fatalf("len(items) = %d, want 5", len(res.Items))
	}

	if res.Items[0].Slug != "a" || res.Items[4].Slug != "last" {
		t.Fatalf("items = %+v", res.Items)
	}

	if len(gotPaths) != 2 {
		t.Fatalf("requests = %v, want 2 paths", gotPaths)
	}
}

func TestSearchFetchesUntilLimit(t *testing.T) {
	t.Parallel()

	pageBody := func(page int) string {
		var cards strings.Builder
		for i := range kosPerPage {
			n := (page-1)*kosPerPage + i
			slug := fmt.Sprintf("s-%d", n)
			cards.WriteString(cardHTML(slug, slug))
		}

		return `<!DOCTYPE html><html><body><div class="listupd">` + cards.String() +
			`</div><div class="hpage">Page ` + fmt.Sprintf("%d", page) + ` of 10</div></body></html>`
	}

	var gotPaths []string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPaths = append(gotPaths, r.URL.Path)

		switch r.URL.Path {
		case "/manga/":
			_, _ = w.Write([]byte(pageBody(1)))
		case "/manga/page/2/":
			_, _ = w.Write([]byte(pageBody(2)))
		case "/manga/page/3/":
			_, _ = w.Write([]byte(pageBody(3)))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	c, err := New(
		Config{BaseURL: srv.URL},
		Deps{HTTP: srv.Client(), Logger: discardLogger(), Solver: nil},
	)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	res, err := c.Search(context.Background(), domain.SearchCacheOpts{Offset: 0, Limit: 100})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	if len(res.Items) <= 80 {
		t.Fatalf("len(items) = %d, want more than 80", len(res.Items))
	}

	if len(res.Items) != 100 {
		t.Fatalf("len(items) = %d, want 100", len(res.Items))
	}

	if res.Items[0].Slug != "s-0" || res.Items[99].Slug != "s-99" {
		t.Fatalf("items[0]=%q items[99]=%q", res.Items[0].Slug, res.Items[99].Slug)
	}

	if len(gotPaths) != 3 {
		t.Fatalf("requests = %v, want 3 paths", gotPaths)
	}
}

func discardLogger() *slog.Logger {
	return slog.New(slog.DiscardHandler)
}
