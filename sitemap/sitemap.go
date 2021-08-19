package sitemap

import (
	"fmt"
	"net/url"
	"sort"
	"strings"
)

type Page struct {
	Depth int
	URLs  []url.URL
}

type Sitemap map[url.URL]Page

func (page Page) String() string {
	lines := make([]string, 0)

	for _, URL := range page.URLs {
		lines = append(lines, fmt.Sprintf("  -> %s", URL.String()))
	}

	sort.Slice(lines, func(i, j int) bool { return lines[i] < lines[j] })
	return strings.Join(lines, "\n")
}

func (sitemap Sitemap) PrettyPrint() string {
	type element struct {
		URL  url.URL
		page Page
	}

	if len(sitemap) == 0 {
		return "[Empty sitemap]"
	}

	sortedByDepth := make([]element, 0)

	for URL, page := range sitemap {
		sortedByDepth = append(sortedByDepth, element{URL, page})
	}

	sort.Slice(sortedByDepth, func(i, j int) bool {
		switch {
		case sortedByDepth[i].page.Depth < sortedByDepth[j].page.Depth:
			return true
		case sortedByDepth[j].page.Depth < sortedByDepth[i].page.Depth:
			return false
		default:
			return sortedByDepth[i].URL.String() < sortedByDepth[j].URL.String()
		}
	})

	lines := make([]string, 0)

	for _, element := range sortedByDepth {
		if len(element.page.URLs) > 0 {
			lines = append(lines, fmt.Sprintf("%s\n%s", element.URL.String(), element.page.String()))
		} else {
			lines = append(lines, element.URL.String())
		}
	}

	return strings.Join(lines, "\n\n")
}

func (sitemap Sitemap) FilterOutLinksThatHaveNoPage() Sitemap {
	result := map[url.URL]Page{}

	for pageURL, page := range sitemap {
		filteredURLs := []url.URL{}
		for _, url := range page.URLs {
			if _, ok := sitemap[url]; ok {
				filteredURLs = append(filteredURLs, url)
			}
		}

		result[pageURL] = Page{page.Depth, filteredURLs}
	}

	return result
}
