package linkextractor

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type HTTPClient struct {
	Do func(req *http.Request) (*http.Response, error)
}

type LinkExtractor interface {
	ExtractLinks(URL url.URL) ([]url.URL, error)
}

func (client HTTPClient) ExtractLinks(URL url.URL) ([]url.URL, error) {
	request, err := http.NewRequest("GET", URL.String(), nil)
	if err != nil {
		return []url.URL{}, err
	}

	request.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml")
	request.Header.Set("Cache-Control", "no-cache")
	request.Header.Set("User-Agent", "Mozilla/5.0 (compatible; sitemapper/0.1)")

	response, err := client.Do(request)
	if err != nil {
		return []url.URL{}, fmt.Errorf("GET request failed: %s", err)
	}
	defer response.Body.Close()

	switch {
	case response.StatusCode != http.StatusOK:
		return []url.URL{}, fmt.Errorf("Got a %s response", response.Status)
	case !strings.HasPrefix(response.Header.Get("Content-Type"), "text/html"):
		return []url.URL{}, fmt.Errorf("Content type is not HTML")
	}

	return extractLinksFromBody(URL, response.Body)
}

func extractLinksFromBody(URL url.URL, readCloser io.ReadCloser) ([]url.URL, error) {
	document, err := goquery.NewDocumentFromReader(readCloser)
	if err != nil {
		return []url.URL{}, fmt.Errorf("Failed to parse response body: %s", err)
	}

	result := make([]url.URL, 0)

	document.Find("a").Each(func(index int, selection *goquery.Selection) {
		href, hrefExists := selection.Attr("href")
		if hrefExists {
			parsedUrl, err := URL.Parse(href)
			if err == nil {
				parsedUrl.Fragment = ""
				result = append(result, *parsedUrl)
			}
		}
	})

	return removeDuplicates(URL, result), nil
}

func removeDuplicates(pageURL url.URL, linkURLs []url.URL) []url.URL {
	seen := map[url.URL]bool{}
	result := make([]url.URL, 0)

	for _, linkURL := range linkURLs {
		if linkURL != pageURL && !seen[linkURL] {
			seen[linkURL] = true
			result = append(result, linkURL)
		}
	}

	return result
}
