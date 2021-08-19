package crawler

import (
	"fmt"
	"io"
	"net/url"

	"github.com/hilverd/sitemapper/linkextractor"
	"github.com/hilverd/sitemapper/sitemap"
)

type Configuration struct {
	MaxConcurrentRequests int
	MaxDepth              int
	SeedURL               url.URL
	ProgressWriter        io.Writer
	SitemapWriter         io.Writer
}

type urlAtDepth struct {
	URL   url.URL
	depth int
}
type crawlState struct {
	linksToBeCrawled  []urlAtDepth
	linksBeingCrawled map[urlAtDepth]bool
	sitemap           sitemap.Sitemap
}

type extractionResult struct {
	pageURL urlAtDepth
	urls    *[]url.URL
}

func Crawl(configuration Configuration, linkextractor linkextractor.LinkExtractor) sitemap.Sitemap {
	state := initialCrawlState(configuration.SeedURL)
	extractionResults := make(chan *extractionResult, configuration.MaxConcurrentRequests)

	state = extractLinksFromNextLink(configuration, linkextractor, state, extractionResults)

	for extractionResult := range extractionResults {
		delete(state.linksBeingCrawled, extractionResult.pageURL)

		if extractionResult.urls != nil {
			if page, alreadyCrawled := state.sitemap[extractionResult.pageURL.URL]; !alreadyCrawled || extractionResult.pageURL.depth < page.Depth {
				state.sitemap[extractionResult.pageURL.URL] = sitemap.Page{Depth: extractionResult.pageURL.depth, URLs: *extractionResult.urls}

				potentiallySuitableLinks := make([]urlAtDepth, 0)

				for _, linkURL := range *extractionResult.urls {
					potentiallySuitableLink := urlAtDepth{linkURL, extractionResult.pageURL.depth + 1}
					potentiallySuitableLinks = append(potentiallySuitableLinks, potentiallySuitableLink)
				}

				suitableLinks := filterOutUnsuitableLinks(configuration, state, potentiallySuitableLinks)
				state.linksToBeCrawled = append(state.linksToBeCrawled, suitableLinks...)
			}
		}

		for shouldExtractLinksFromAnotherLink(configuration, state) {
			state = extractLinksFromNextLink(configuration, linkextractor, state, extractionResults)
		}

		if len(state.linksBeingCrawled) == 0 {
			break
		}
	}

	return state.sitemap.FilterOutLinksThatHaveNoPage()
}

func initialCrawlState(seedURL url.URL) crawlState {
	return crawlState{
		linksToBeCrawled:  []urlAtDepth{{seedURL, 0}},
		linksBeingCrawled: map[urlAtDepth]bool{},
		sitemap:           map[url.URL]sitemap.Page{},
	}
}

func extractLinksFromNextLink(
	configuration Configuration,
	linkextractor linkextractor.LinkExtractor,
	state crawlState,
	extractionResults chan *extractionResult,
) crawlState {
	newState := state
	link := newState.linksToBeCrawled[0]
	newState.linksToBeCrawled = newState.linksToBeCrawled[1:]
	newState.linksBeingCrawled[link] = true
	URL := link.URL

	go func() {
		if page, alreadyCrawled := state.sitemap[URL]; alreadyCrawled {
			fmt.Fprintf(configuration.ProgressWriter, "Using cached links from %s\n", URL.String())
			extractionResults <- &extractionResult{pageURL: link, urls: &page.URLs}
		} else {
			fmt.Fprintf(configuration.ProgressWriter, "Extracting links from %s\n", URL.String())
			links, err := linkextractor.ExtractLinks(URL)

			if err == nil {
				extractionResults <- &extractionResult{pageURL: link, urls: &links}
			} else {
				fmt.Fprintf(configuration.ProgressWriter, "Warning: failed to extract links from %s: %s\n", URL.String(), err)
				extractionResults <- &extractionResult{pageURL: link, urls: nil}
			}
		}
	}()

	return newState
}

func filterOutUnsuitableLinks(configuration Configuration, state crawlState, links []urlAtDepth) []urlAtDepth {
	result := make([]urlAtDepth, 0)
	for _, link := range links {
		if linkIsSuitable(configuration, state, link) {
			result = append(result, link)
		}
	}

	return result
}

func linkIsSuitable(configuration Configuration, state crawlState, link urlAtDepth) bool {
	if page, alreadyCrawled := state.sitemap[link.URL]; alreadyCrawled && page.Depth <= link.depth {
		return false
	}

	switch {
	case 0 < configuration.MaxDepth && configuration.MaxDepth < link.depth:
		return false
	case state.linksBeingCrawled[link]:
		return false
	case link.URL.Host != configuration.SeedURL.Host:
		return false
	case link.URL.Scheme != "http" && link.URL.Scheme != "https":
		return false
	default:
		return true
	}
}

func shouldExtractLinksFromAnotherLink(configuration Configuration, state crawlState) bool {
	switch {
	case len(state.linksToBeCrawled) == 0:
		return false
	case 0 < configuration.MaxConcurrentRequests && configuration.MaxConcurrentRequests <= len(state.linksBeingCrawled):
		return false
	default:
		return true
	}
}
