package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"time"

	"github.com/hilverd/sitemapper/crawler"
	"github.com/hilverd/sitemapper/linkextractor"
)

func main() {
	configuration, httpClient := parseCommandLineOptions(nil)
	sitemap := crawler.Crawl(configuration, httpClient)
	fmt.Fprintln(configuration.SitemapWriter, sitemap.PrettyPrint())
}

func parseCommandLineOptions(arguments []string) (crawler.Configuration, linkextractor.HTTPClient) {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, `Usage: sitemapper [OPTIONS] SEED_URL
Crawl web pages starting from SEED_URL and print a basic site map to standard output.

Options:
`)
		flag.PrintDefaults()
	}

	verbose := flag.Bool("v", false, "verbosely list pages as they are being processed")
	requestTimeoutSeconds := flag.Int("request-timeout", 30, "HTTP request timeout in seconds (zero means no timeout)")
	maxConcurrentRequests := flag.Int("max-concurrent-requests", runtime.GOMAXPROCS(0), "maximum number of concurrent requests")
	maxDepth := flag.Int("max-depth", 0, "maximum crawl depth, i.e. distance from seed URL (zero means no maximum)")

	if arguments == nil {
		flag.Parse()
	} else {
		_ = flag.CommandLine.Parse(arguments)
	}

	switch {
	case *requestTimeoutSeconds < 0:
		log.Fatal("request-timeout must be at least zero")
	case *maxConcurrentRequests <= 0:
		log.Fatal("max-concurrent-requests must be greater than zero")
	case *maxDepth < 0:
		log.Fatal("max-depth must be at least zero")
	}

	seedURLStrings := flag.Args()
	if len(seedURLStrings) != 1 {
		flag.Usage()
		os.Exit(1)
	}

	rawSeedURL := seedURLStrings[0]
	seedURL, err := normaliseURL(rawSeedURL)
	if err != nil || seedURL.Scheme != "http" && seedURL.Scheme != "https" {
		log.Fatalf("Invalid seed URL: %s", rawSeedURL)
	}

	progressWriter := ioutil.Discard
	if *verbose {
		progressWriter = os.Stderr
	}

	return crawler.Configuration{
			MaxConcurrentRequests: *maxConcurrentRequests,
			MaxDepth:              *maxDepth,
			SeedURL:               *seedURL,
			ProgressWriter:        progressWriter,
			SitemapWriter:         os.Stdout,
		}, linkextractor.HTTPClient{
			Do: (&http.Client{
				Timeout: time.Duration(*requestTimeoutSeconds) * time.Second,
			}).Do,
		}
}

func normaliseURL(rawurl string) (*url.URL, error) {
	result, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}

	if result.Scheme == "" {
		result.Scheme = "https"
	}

	result, err = url.Parse(result.String())
	if err != nil {
		return nil, err
	}

	if result.Path == "" {
		result.Path = "/"
	}

	return result, nil
}
