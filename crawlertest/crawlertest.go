package crawlertest

import (
	"net/url"
)

func MakeURL(rawurl string) url.URL {
	result, err := url.Parse(rawurl)
	if err != nil {
		panic(err)
	}
	return *result
}
