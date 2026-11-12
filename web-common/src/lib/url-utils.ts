/**
 * Copies all parameters from the source URLSearchParams object to the target URLSearchParams object,
 * modifying the target object directly. Any existing parameters in the target with the same keys
 * will be overwritten.
 *
 * Note: Unlike mergeAndRetainParams, this function modifies the target object directly
 * instead of creating a new URLSearchParams object.
 *
 * @param fromSearchParams - The source URLSearchParams object
 * @param toSearchParams - The target URLSearchParams object that will be modified
 */
export function copyParamsToTarget(
  fromSearchParams: URLSearchParams,
  toSearchParams: URLSearchParams,
) {
  fromSearchParams.forEach((value, key) => {
    toSearchParams.set(key, value);
  });
}

export function copyWithAdditionalArguments(
  url: URL,
  args: Record<string, string>,
  deleteArgs: Record<string, boolean> = {},
) {
  const newUrl = new URL(url);
  for (const [key, value] of Object.entries(args)) {
    newUrl.searchParams.set(key, value);
  }
  for (const key of Object.keys(deleteArgs)) {
    newUrl.searchParams.delete(key);
  }
  return newUrl;
}

const LOCAL_UI_HOSTS = new Set(["localhost", "127.0.0.1", "[::1]"]);

// `rill start` advertises http://localhost:<port> even when the UI is opened
// through another host. Keep the path, and use the page the browser loaded.
export function usePageOriginForLocalhost(
  advertised: string,
  pageHref: string,
): string {
  let advertisedURL: URL;
  try {
    advertisedURL = new URL(advertised);
  } catch {
    return advertised;
  }
  if (!LOCAL_UI_HOSTS.has(advertisedURL.hostname)) {
    return advertised;
  }
  const pageURL = new URL(pageHref);
  if (
    advertisedURL.protocol === pageURL.protocol &&
    advertisedURL.host === pageURL.host
  ) {
    return advertised;
  }
  advertisedURL.protocol = pageURL.protocol;
  advertisedURL.host = pageURL.host;
  return advertisedURL.toString();
}

export function copySubsetParams(src: URLSearchParams, keys: Set<string>) {
  const newParams = new URLSearchParams();
  for (const key of keys) {
    if (!src.has(key)) continue;
    newParams.set(key, src.get(key)!);
  }
  return newParams;
}
