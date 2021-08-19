package crawler

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/hilverd/sitemapper/crawlertest"
	"github.com/hilverd/sitemapper/linkextractor"
	"github.com/hilverd/sitemapper/sitemap"
)

type stubLinkExtractor struct {
	urlToLinks map[url.URL][]url.URL
}

func (stub stubLinkExtractor) ExtractLinks(URL url.URL) ([]url.URL, error) {
	if links, ok := stub.urlToLinks[URL]; ok {
		return links, nil
	} else {
		return []url.URL{}, fmt.Errorf("No links defined for %s", URL.String())
	}
}

func TestCrawl(t *testing.T) {
	type args struct {
		configuration Configuration
		linkextractor linkextractor.LinkExtractor
	}

	stub := stubLinkExtractor{
		urlToLinks: map[url.URL][]url.URL{
			crawlertest.MakeURL("https://example.com/"): {
				crawlertest.MakeURL("https://example.com/one"),
				crawlertest.MakeURL("https://example.com/two"),
				crawlertest.MakeURL("https://apple.com/mac/"),
			},
			crawlertest.MakeURL("https://example.com/one"): {},
			crawlertest.MakeURL("https://example.com/two"): {
				crawlertest.MakeURL("https://example.com/three"),
			},
			crawlertest.MakeURL("https://apple.com/mac/"): {
				crawlertest.MakeURL("https://apple.com/"),
			},
			crawlertest.MakeURL("https://example.com/three"): {},
		},
	}

	tests := []struct {
		name string
		args args
		want sitemap.Sitemap
	}{
		{
			name: "happy path",
			args: args{
				configuration: Configuration{
					SeedURL:        crawlertest.MakeURL("https://example.com/"),
					ProgressWriter: ioutil.Discard,
				},
				linkextractor: stub,
			},
			want: map[url.URL]sitemap.Page{
				crawlertest.MakeURL("https://example.com/"): {
					Depth: 0,
					URLs: []url.URL{
						crawlertest.MakeURL("https://example.com/one"),
						crawlertest.MakeURL("https://example.com/two"),
					},
				},
				crawlertest.MakeURL("https://example.com/one"): {
					Depth: 1,
					URLs:  []url.URL{},
				},
				crawlertest.MakeURL("https://example.com/two"): {
					Depth: 1,
					URLs: []url.URL{
						crawlertest.MakeURL("https://example.com/three"),
					},
				},
				crawlertest.MakeURL("https://example.com/three"): {
					Depth: 2,
					URLs:  []url.URL{},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Crawl(tt.args.configuration, tt.args.linkextractor); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Crawl() = %v, want %v", got, tt.want)
			}
		})
	}
}

type specialCaseLinkExtractor struct {
	eWasRetrieved bool
}

func (stub *specialCaseLinkExtractor) ExtractLinks(URL url.URL) ([]url.URL, error) {
	switch URL.String() {
	case "https://example.com/seed":
		return []url.URL{
			crawlertest.MakeURL("https://example.com/a"),
			crawlertest.MakeURL("https://example.com/b"),
		}, nil
	case "https://example.com/a":
		return []url.URL{
			crawlertest.MakeURL("https://example.com/c"),
		}, nil
	case "https://example.com/b":
		/*
			Delay returning a response for this page until the crawler has retrieved page E.
			This is to highlight a subtle concurrency issue where a page can be reached via
			multiple paths, and we need to make sure that its depth is recorded as the
			shortest of those paths.

			seed
			 | \
			 A  B
			 |   \
			 C -> D -> E -> F

			If we reach D via A -> C first, then it has depth 3.
			But it can also be reached via B so should have depth 2.
		*/
		deadline := time.Now().Add(2 * time.Second)
		for {
			if stub.eWasRetrieved || time.Now().After(deadline) {
				break
			}
		}

		return []url.URL{
			crawlertest.MakeURL("https://example.com/d"),
		}, nil
	case "https://example.com/c":
		return []url.URL{
			crawlertest.MakeURL("https://example.com/d"),
		}, nil
	case "https://example.com/d":
		return []url.URL{
			crawlertest.MakeURL("https://example.com/e"),
		}, nil
	case "https://example.com/e":
		stub.eWasRetrieved = true

		return []url.URL{
			crawlertest.MakeURL("https://example.com/f"),
		}, nil
	case "https://example.com/f":
		return []url.URL{}, nil
	default:
		return []url.URL{}, fmt.Errorf("No links defined for %s", URL.String())
	}
}

func TestCrawlWithMaxDepth(t *testing.T) {
	type args struct {
		configuration Configuration
		linkextractor linkextractor.LinkExtractor
	}

	tests := []struct {
		name string
		args args
		want sitemap.Sitemap
	}{
		{
			name: "links beyond the maximum depth are not crawled",
			args: args{
				configuration: Configuration{
					MaxDepth:       4,
					SeedURL:        crawlertest.MakeURL("https://example.com/seed"),
					ProgressWriter: ioutil.Discard,
				},
				linkextractor: &specialCaseLinkExtractor{},
			},
			want: map[url.URL]sitemap.Page{
				crawlertest.MakeURL("https://example.com/seed"): {
					Depth: 0,
					URLs: []url.URL{
						crawlertest.MakeURL("https://example.com/a"),
						crawlertest.MakeURL("https://example.com/b"),
					},
				},
				crawlertest.MakeURL("https://example.com/a"): {
					Depth: 1,
					URLs: []url.URL{
						crawlertest.MakeURL("https://example.com/c"),
					},
				},
				crawlertest.MakeURL("https://example.com/b"): {
					Depth: 1,
					URLs: []url.URL{
						crawlertest.MakeURL("https://example.com/d"),
					},
				},
				crawlertest.MakeURL("https://example.com/c"): {
					Depth: 2,
					URLs: []url.URL{
						crawlertest.MakeURL("https://example.com/d"),
					},
				},
				crawlertest.MakeURL("https://example.com/d"): {
					Depth: 2,
					URLs: []url.URL{
						crawlertest.MakeURL("https://example.com/e"),
					},
				},
				crawlertest.MakeURL("https://example.com/e"): {
					Depth: 3,
					URLs: []url.URL{
						crawlertest.MakeURL("https://example.com/f"),
					},
				},
				crawlertest.MakeURL("https://example.com/f"): {
					Depth: 4,
					URLs:  []url.URL{},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Crawl(tt.args.configuration, tt.args.linkextractor); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Crawl() = %v, want %v", got, tt.want)
			}
		})
	}
}
