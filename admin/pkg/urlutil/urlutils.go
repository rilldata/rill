package urlutil

import "net/url"

func UrlWithQuery(urlString string, query map[string]string) (string, error) {
	parsedURL, err := url.Parse(urlString)
	if err != nil {
		return "", err
	}

	qry := parsedURL.Query()
	for key, value := range query {
		qry.Set(key, value)
	}
	parsedURL.RawQuery = qry.Encode()
	return parsedURL.String(), nil
}

func MustJoinURL(base string, elem ...string) string {
	joinedURL, err := url.JoinPath(base, elem...)
	if err != nil {
		panic(err)
	}
	return joinedURL
}
