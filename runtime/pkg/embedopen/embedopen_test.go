package embedopen

import (
	"encoding/base64"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestEncodeDecode(t *testing.T) {
	ts := time.Date(2026, 10, 20, 0, 0, 0, 0, time.UTC)

	s, err := Encode(&Payload{Kind: KindReport, Name: "weekly_revenue", ExecutionTime: ts})
	require.NoError(t, err)
	require.NotContains(t, s, "=")
	data, err := base64.RawURLEncoding.DecodeString(s)
	require.NoError(t, err)
	require.JSONEq(t, `{"kind":"report","name":"weekly_revenue","execution_time":"2026-10-20T00:00:00Z"}`, string(data))

	p, err := Decode(s)
	require.NoError(t, err)
	require.Equal(t, KindReport, p.Kind)
	require.Equal(t, "weekly_revenue", p.Name)
	require.True(t, ts.Equal(p.ExecutionTime))

	// The execution time is optional
	s, err = Encode(&Payload{Kind: KindReport, Name: "r1"})
	require.NoError(t, err)
	data, err = base64.RawURLEncoding.DecodeString(s)
	require.NoError(t, err)
	require.JSONEq(t, `{"kind":"report","name":"r1"}`, string(data))
	p, err = Decode(s)
	require.NoError(t, err)
	require.Equal(t, &Payload{Kind: KindReport, Name: "r1"}, p)

	// Non-UTC times are encoded in UTC
	s, err = Encode(&Payload{Kind: KindReport, Name: "r1", ExecutionTime: ts.In(time.FixedZone("x", 3600))})
	require.NoError(t, err)
	data, err = base64.RawURLEncoding.DecodeString(s)
	require.NoError(t, err)
	require.Contains(t, string(data), `"2026-10-20T00:00:00Z"`)
}

func TestValidation(t *testing.T) {
	_, err := Encode(&Payload{Kind: "dashboard", Name: "x"})
	require.ErrorContains(t, err, "invalid kind")
	_, err = Encode(&Payload{Kind: KindReport})
	require.ErrorContains(t, err, "missing name")

	encode := func(s string) string { return base64.RawURLEncoding.EncodeToString([]byte(s)) }
	_, err = Decode("not base64!")
	require.ErrorContains(t, err, "invalid encoding")
	_, err = Decode(encode("not json"))
	require.ErrorContains(t, err, "invalid payload")
	_, err = Decode(encode(`{"kind":"canvas","name":"x"}`))
	require.ErrorContains(t, err, "invalid kind")
	_, err = Decode(encode(`{"kind":"report"}`))
	require.ErrorContains(t, err, "missing name")
	_, err = Decode(encode(`{"kind":"report","name":"x","execution_time":"yesterday"}`))
	require.ErrorContains(t, err, "invalid payload")
	// Padded input is rejected
	_, err = Decode(base64.URLEncoding.EncodeToString([]byte(`{"kind":"report","name":"x1"}`)))
	require.Error(t, err)
}

func TestLink(t *testing.T) {
	p := &Payload{Kind: KindReport, Name: "r1"}
	v, err := Encode(p)
	require.NoError(t, err)

	link, err := Link("https://app.acme.com/analytics/open?tenant=foo", p)
	require.NoError(t, err)
	u, err := url.Parse(link)
	require.NoError(t, err)
	require.Equal(t, "app.acme.com", u.Host)
	require.Equal(t, "/analytics/open", u.Path)
	require.Equal(t, "foo", u.Query().Get("tenant"))
	require.Equal(t, v, u.Query().Get(QueryParam))

	_, err = Link("https://app.acme.com", &Payload{Kind: "x", Name: "r1"})
	require.Error(t, err)
}

func TestValidateOpenURL(t *testing.T) {
	require.NoError(t, ValidateOpenURL("https://app.acme.com/open"))
	require.NoError(t, ValidateOpenURL("http://localhost:3000/open?x=1"))
	require.Error(t, ValidateOpenURL("/open"))
	require.Error(t, ValidateOpenURL("app.acme.com/open"))
	require.Error(t, ValidateOpenURL("javascript:alert(1)"))
	require.Error(t, ValidateOpenURL("ftp://acme.com/open"))
	require.Error(t, ValidateOpenURL("https://"))
}
