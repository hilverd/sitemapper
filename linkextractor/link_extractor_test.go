package linkextractor

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/hilverd/sitemapper/crawlertest"
)

func TestHTTPClient_ExtractLinks(t *testing.T) {
	type fields struct {
		Do func(req *http.Request) (*http.Response, error)
	}
	type args struct {
		URL url.URL
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []url.URL
		wantErr bool
	}{
		{
			name: "links are identified, processed and de-duplicated",
			fields: fields{
				Do: stubHttpClientDo(http.StatusOK, "text/html", `
				<html>
				  <body>
					<a href="/foo?sort-by=priority">Foo</a>
					<a href="bar#top">Bar</a>
					<a href="https://example.org">example.org</a>
					<a href="bar#bottom">Bar again</a>
					<a href="/about/">Link to self</a>
				  </body>
				</html>
			`),
			},
			args: args{
				URL: crawlertest.MakeURL("https://example.com/about/"),
			},
			want: []url.URL{
				crawlertest.MakeURL("https://example.com/foo?sort-by=priority"),
				crawlertest.MakeURL("https://example.com/about/bar"),
				crawlertest.MakeURL("https://example.org"),
			},
			wantErr: false,
		},
		{
			name: "no response means no links are returned",
			fields: fields{
				Do: func(req *http.Request) (*http.Response, error) {
					return nil, fmt.Errorf("Request failed")
				},
			},
			args: args{
				URL: crawlertest.MakeURL("https://example.com/about/"),
			},
			want:    []url.URL{},
			wantErr: true,
		},
		{
			name: "non-200 response means no links are returned",
			fields: fields{
				Do: stubHttpClientDo(http.StatusNotFound, "text/html", `
				<html><body><a href="main">Main</a></body></html>
			`),
			},
			args: args{
				URL: crawlertest.MakeURL("https://example.com/about/"),
			},
			want:    []url.URL{},
			wantErr: true,
		},
		{
			name: "non-HTML response means no links are returned",
			fields: fields{
				Do: stubHttpClientDo(http.StatusOK, "image/svg+xml", `
				<svg viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg"><a href="main">Main</a></svg>
			`),
			},
			args: args{
				URL: crawlertest.MakeURL("https://example.com/about/"),
			},
			want:    []url.URL{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := HTTPClient{
				Do: tt.fields.Do,
			}
			got, err := client.ExtractLinks(tt.args.URL)
			if (err != nil) != tt.wantErr {
				t.Errorf("HTTPClient.ExtractLinks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HTTPClient.ExtractLinks() = %v, want %v", got, tt.want)
			}
		})
	}
}

func stubHttpClientDo(statusCode int, contentType string, responseBody string) func(req *http.Request) (*http.Response, error) {
	return func(req *http.Request) (*http.Response, error) {
		switch {
		case len(req.Header["Accept"]) == 0 || req.Header["Accept"][0] != "text/html,application/xhtml+xml,application/xml":
			return makeHttpResponse(statusCode, contentType, `
				<html><body><a href="/expected-accept-header-not-present">Error</a></body></html>
				`)
		case len(req.Header["Cache-Control"]) == 0 || req.Header["Cache-Control"][0] != "no-cache":
			return makeHttpResponse(statusCode, contentType, `
				<html><body><a href="/expected-cache-control-header-not-present">Error</a></body></html>
				`)
		case len(req.Header["User-Agent"]) == 0 || req.Header["User-Agent"][0] != "Mozilla/5.0 (compatible; sitemapper/0.1)":
			return makeHttpResponse(statusCode, contentType, `
					<html><body><a href="/expected-user-agent-header-not-present">Error</a></body></html>
					`)
		default:
			return makeHttpResponse(statusCode, contentType, responseBody)
		}
	}
}

func makeHttpResponse(statusCode int, contentType string, responseBody string) (*http.Response, error) {
	return &http.Response{
		Status:     http.StatusText(statusCode),
		StatusCode: statusCode,
		Header:     map[string][]string{"Content-Type": {contentType}},
		Body:       ioutil.NopCloser(strings.NewReader(responseBody)),
	}, nil
}
