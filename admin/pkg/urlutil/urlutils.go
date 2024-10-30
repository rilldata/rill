package urlutil

import "net/url"

func WithQuery(urlString string, query map[string]string) (string, error) {
	if len(query) == 0 {
		return urlString, nil
	}

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

func MustWithQuery(urlString string, query map[string]string) string {
	newURL, err := WithQuery(urlString, query)
	if err != nil {
		panic(err)
	}
	return newURL
}

func MustJoinURL(base string, elem ...string) string {
	joinedURL, err := url.JoinPath(base, elem...)
	if err != nil {
		panic(err)
	}
	return joinedURL
}
