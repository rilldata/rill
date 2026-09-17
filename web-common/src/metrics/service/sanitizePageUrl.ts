// Deliberately free of characters that a URL would percent-encode, so the marker reads the same
// whether it lands in the path or the query string.
const Redacted = "redacted";

// Query params that carry a credential rather than dashboard state.
// `token` is used by public report and alert links, `access_token` by embeds.
const CredentialParams = ["token", "access_token"];

// Public URLs put the magic token in the path, e.g. /{org}/{project}/-/share/{token}/explore/{name}.
const ShareTokenPattern = /\/-\/share\/[^/]+/;

/**
 * Strips credentials out of a URL before it is recorded as telemetry.
 *
 * Public URL and embed tokens grant access to a dashboard and stay valid after use, so they must not
 * be sent to the telemetry pipeline: the events are queryable, and are exposed to customers through
 * the monitoring dashboards built on them.
 *
 * Returns an empty string for input that isn't a parseable URL, so that a malformed location can
 * never pass a raw token through untouched.
 */
export function sanitizePageUrl(href: string): string {
  let url: URL;
  try {
    url = new URL(href);
  } catch {
    return "";
  }

  url.pathname = url.pathname.replace(
    ShareTokenPattern,
    `/-/share/${Redacted}`,
  );

  for (const param of CredentialParams) {
    if (url.searchParams.has(param)) {
      url.searchParams.set(param, Redacted);
    }
  }

  return url.toString();
}
