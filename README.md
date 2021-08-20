# Sitemapper

This is a very basic site mapper -- implementing it helped me get familiar with the Go programming language. It only supports crawling pages from a single domain.

## Usage

After cloning this repository and compiling the application using

```
go build github.com/hilverd/sitemapper
```

you can run

```
./sitemapper -help
```

to get a help message. Here is an example of how to (politely) crawl part of [apple.com](https://apple.com):

```
./sitemapper -v -max-depth 2 -max-concurrent-requests 8 apple.com | tee apple-sitemap.txt
```

## Development

You can use

```
./test-all.sh
```

to run the tests for each package as well as the (separate) end-to-end test.

## Architecture

There is a bit more documentation available via the site that is used in the end-to-end test:

```
cd end-to-end-test && go run server.go
```
