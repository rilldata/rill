package oauthclient

import (
	"encoding/json"
	"fmt"
	"html"
	"net/url"
	"strings"
)

// CallbackPage renders the page shown at the end of a browser-based authorization flow.
// It notifies the window that opened the popup (the chat) and closes itself; if it was not opened as a popup,
// it offers a link back to redirectURL. errMsg is empty on success.
func CallbackPage(connector, errMsg, redirectURL string) string {
	payload, _ := json.Marshal(map[string]string{ // nolint:errchkjson // Marshaling a map of strings cannot fail
		"type":      "rill-mcp-oauth",
		"connector": connector,
		"status":    map[bool]string{true: "error", false: "connected"}[errMsg != ""],
		"error":     errMsg,
	})
	var body string
	if errMsg != "" {
		body = fmt.Sprintf("<h1>Could not connect</h1><p>%s</p>", html.EscapeString(errMsg))
	} else {
		body = "<h1>Connected</h1><p>You can close this window and return to Rill.</p>"
	}
	link := ""
	if isLinkableRedirect(redirectURL) {
		link = `<p><a href="` + html.EscapeString(redirectURL) + `">Back to Rill</a></p>`
	}
	return fmt.Sprintf(`<!doctype html>
<html><head><meta charset="utf-8"><title>Rill</title>
<style>body{font-family:system-ui,sans-serif;margin:48px auto;max-width:480px;color:#1f2937}h1{font-size:20px}a{color:#2563eb}</style>
</head><body>%s%s
<script>
(function(){
  var msg=%s;
  try{ if(window.opener){ window.opener.postMessage(msg,"*"); } }catch(e){}
  if(window.opener){ setTimeout(function(){ window.close(); }, 500); }
})();
</script></body></html>`, body, link, string(payload))
}

// isLinkableRedirect reports whether a return URL may be rendered as a link on the callback page:
// a relative path or an absolute http(s) URL. The callers validate the host; this guards against javascript: and data: URLs.
func isLinkableRedirect(redirectURL string) bool {
	if redirectURL == "" {
		return false
	}
	u, err := url.Parse(redirectURL)
	if err != nil {
		return false
	}
	if u.Scheme == "" && u.Host == "" {
		return strings.HasPrefix(u.Path, "/") && !strings.HasPrefix(u.Path, "//")
	}
	return (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}
