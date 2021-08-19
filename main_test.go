package main

import (
	"net/url"
	"os"
	"reflect"
	"testing"

	"github.com/hilverd/sitemapper/crawler"
	"github.com/hilverd/sitemapper/crawlertest"
)

func Test_parseCommandLineOptions(t *testing.T) {
	type args struct {
		arguments []string
	}
	tests := []struct {
		name string
		args args
		want crawler.Configuration
	}{
		{
			name: "known command line options",
			args: args{
				arguments: []string{
					"-v",
					"-request-timeout", "10",
					"-max-concurrent-requests", "2",
					"-max-depth", "3",
					"apple.com",
				},
			},
			want: crawler.Configuration{
				MaxConcurrentRequests: 2,
				MaxDepth:              3,
				SeedURL:               crawlertest.MakeURL("https://apple.com/"),
				ProgressWriter:        os.Stderr,
				SitemapWriter:         os.Stdout,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := parseCommandLineOptions(tt.args.arguments)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseCommandLineOptions() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_normaliseURL(t *testing.T) {
	type args struct {
		rawurl string
	}
	tests := []struct {
		name    string
		args    args
		want    url.URL
		wantErr bool
	}{
		{
			name: "https scheme gets added if no scheme present",
			args: args{
				rawurl: "apple.com/ipad",
			},
			want:    crawlertest.MakeURL("https://apple.com/ipad"),
			wantErr: false,
		},
		{
			name: "trailing slash gets added if path is empty",
			args: args{
				rawurl: "http://apple.com",
			},
			want:    crawlertest.MakeURL("http://apple.com/"),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normaliseURL(tt.args.rawurl)
			if (err != nil) != tt.wantErr {
				t.Errorf("normaliseURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(*got, tt.want) {
				t.Errorf("normaliseURL() = %v, want %v", *got, tt.want)
			}
		})
	}
}
