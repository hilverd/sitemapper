package crawlertest

import (
	"net/url"
	"reflect"
	"testing"
)

func TestMakeURL(t *testing.T) {
	type args struct {
		rawurl string
	}
	tests := []struct {
		name string
		args args
		want url.URL
	}{
		{
			name: "valid URL",
			args: args{
				rawurl: "https://www.example.com/foo/bar",
			},
			want: url.URL{
				Scheme: "https",
				Host:   "www.example.com",
				Path:   "/foo/bar",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MakeURL(tt.args.rawurl); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MakeURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
