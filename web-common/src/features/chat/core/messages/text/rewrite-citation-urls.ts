import {
  DASHBOARD_CITATION_URL_PATHNAME_REGEX,
  LEGACY_DASHBOARD_CITATION_URL_PATHNAME_REGEX,
} from "@rilldata/web-common/features/chat/core/messages/text/citation-url-mapper.ts";

// Matches absolute http(s) URLs and root-relative paths that contain `/-/`.
// Citation URLs always include `/-/` (either `/-/open-query` or `/-/ai/.../-/open`),
// so we use it as a cheap pre-filter before parsing.
const CITATION_URL_DETECT_REGEX = /https?:\/\/[^\s)<>"`]+|\/-\/[^\s)<>"`]+/g;

const FALLBACK_BASE = "http://localhost";

/**
 * Rewrites citation URLs in the given markdown content using the provided async mapper.
 * Non-citation URLs and non-URL text are left untouched. Per-URL mapper calls run in
 * parallel; replacements are applied in reverse index order so earlier offsets remain valid.
 */
export async function rewriteCitationUrls(
  content: string,
  mapper: (url: URL) => Promise<string>,
): Promise<string> {
  const matches = [...content.matchAll(CITATION_URL_DETECT_REGEX)];
  if (matches.length === 0) return content;

  const base =
    typeof window !== "undefined" ? window.location.origin : FALLBACK_BASE;

  const replacements = await Promise.all(
    matches.map(async (match) => {
      const raw = match[0];
      let url: URL;
      try {
        url = new URL(raw, base);
      } catch {
        return null;
      }
      const isCitation =
        DASHBOARD_CITATION_URL_PATHNAME_REGEX.test(url.pathname) ||
        LEGACY_DASHBOARD_CITATION_URL_PATHNAME_REGEX.test(url.pathname);
      if (!isCitation) return null;

      const mapped = await mapper(url);
      // Preserve relativity: if the source was a relative path and the mapper
      // returned an absolute URL on the synthetic base, strip the base back off.
      const wasRelative = raw.startsWith("/");
      const replacement =
        wasRelative && mapped.startsWith(base)
          ? mapped.slice(base.length)
          : mapped;
      return { index: match.index ?? 0, length: raw.length, replacement };
    }),
  );

  const valid = replacements
    .filter((r): r is { index: number; length: number; replacement: string } =>
      r !== null,
    )
    .sort((a, b) => b.index - a.index);

  let result = content;
  for (const { index, length, replacement } of valid) {
    result = result.slice(0, index) + replacement + result.slice(index + length);
  }
  return result;
}
