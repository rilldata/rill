package pathutil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test(t *testing.T) {
	tt := []struct {
		a    string
		b    string
		want string
	}{
		{
			a:    "a/b/c/d",
			b:    "a/b/c/e",
			want: "a/b/c",
		},
		{
			a:    "a/b/c/d",
			b:    "a/b/c/d",
			want: "a/b/c/d",
		},
		{
			a:    "a/b/c/d",
			b:    "a/b/c",
			want: "a/b/c",
		},
		{
			a:    "a/b/c/d",
			b:    "a",
			want: "a",
		},
		{
			a:    "a/b",
			b:    "c/d",
			want: "",
		},
		{
			a:    "a/b/",
			b:    "a/b/",
			want: "a/b/",
		},
		{
			a:    "a/b/",
			b:    "a/b",
			want: "a/b",
		},
		{
			a:    "a/b",
			b:    "/a/b",
			want: "",
		},
		{
			a:    "",
			b:    "",
			want: "",
		},
		{
			a:    "/",
			b:    "/",
			want: "/",
		},
		{
			a:    "///",
			b:    "//",
			want: "//",
		},
		{
			a:    "a//b",
			b:    "a/b",
			want: "a",
		},
		{
			a:    "aa",
			b:    "ab",
			want: "",
		},
	}

	for _, tc := range tt {
		// Test both directions
		got := CommonPrefix(tc.a, tc.b)
		require.Equal(t, tc.want, got)
		got = CommonPrefix(tc.b, tc.a)
		require.Equal(t, tc.want, got)
	}

}
