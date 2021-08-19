package sitemap

import (
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/hilverd/sitemapper/crawlertest"
)

func TestSitemap_PrettyPrint(t *testing.T) {
	tests := []struct {
		name    string
		sitemap Sitemap
		want    string
	}{
		{
			name:    "empty sitemap",
			sitemap: map[url.URL]Page{},
			want:    "[Empty sitemap]",
		},
		{
			name: "simple sitemap",
			sitemap: map[url.URL]Page{
				crawlertest.MakeURL("https://example.com/first"): {
					0,
					[]url.URL{
						crawlertest.MakeURL("https://example.com/second"),
						crawlertest.MakeURL("https://example.com/third"),
					},
				},
				crawlertest.MakeURL("https://example.com/second"): {
					1,
					[]url.URL{},
				},
				crawlertest.MakeURL("https://example.com/third"): {
					1,
					[]url.URL{
						crawlertest.MakeURL("https://example.com/first"),
					},
				},
			},
			want: `
https://example.com/first
  -> https://example.com/second
  -> https://example.com/third

https://example.com/second

https://example.com/third
  -> https://example.com/first
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want := strings.TrimSpace(tt.want)
			if got := tt.sitemap.PrettyPrint(); got != want {
				t.Errorf("Sitemap.PrettyPrint() = %v, want %v", got, want)
			}
		})
	}
}

func TestSitemap_FilterOutLinksThatHaveNoPage(t *testing.T) {
	tests := []struct {
		name    string
		sitemap Sitemap
		want    Sitemap
	}{
		{
			name: "links for which no page was created get removed",
			sitemap: map[url.URL]Page{
				crawlertest.MakeURL("https://example.com/retrieved-1"): {
					0,
					[]url.URL{
						crawlertest.MakeURL("https://example.com/retrieved-2"),
						crawlertest.MakeURL("https://example.com/not-retrieved-1"),
					},
				},
				crawlertest.MakeURL("https://example.com/retrieved-2"): {
					1,
					[]url.URL{
						crawlertest.MakeURL("https://example.com/retrieved-3"),
						crawlertest.MakeURL("https://example.com/not-retrieved-2"),
					},
				},
				crawlertest.MakeURL("https://example.com/retrieved-3"): {
					1,
					[]url.URL{},
				},
			},
			want: map[url.URL]Page{
				crawlertest.MakeURL("https://example.com/retrieved-1"): {
					0,
					[]url.URL{
						crawlertest.MakeURL("https://example.com/retrieved-2"),
					},
				},
				crawlertest.MakeURL("https://example.com/retrieved-2"): {
					1,
					[]url.URL{
						crawlertest.MakeURL("https://example.com/retrieved-3"),
					},
				},
				crawlertest.MakeURL("https://example.com/retrieved-3"): {
					1,
					[]url.URL{},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.sitemap.FilterOutLinksThatHaveNoPage(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Sitemap.FilterOutLinksThatHaveNoPage() = %v, want %v", got, tt.want)
			}
		})
	}
}
