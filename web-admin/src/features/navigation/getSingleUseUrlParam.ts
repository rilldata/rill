import { redirect } from "@sveltejs/kit";

export function getSingleUseUrlParam(
  url: URL,
  param: string,
  storageKey: string,
) {
  // Save the state in localStorage and redirect to the url without it.
  // This prevents a refresh or saving the url from re-triggering in the page
  if (url.searchParams.has(param)) {
    try {
      localStorage.setItem(storageKey, url.searchParams.get(param));
    } catch {
      // no-op
    }
    const redirectUrl = new URL(url);
    redirectUrl.searchParams.delete(param);
    if (url.searchParams.has("redirect")) {
      redirectUrl.searchParams.set(
        "redirect",
        url.searchParams.get("redirect"),
      );
    }
    throw redirect(307, redirectUrl.pathname + redirectUrl.search);
  }

  const value = localStorage.getItem(storageKey);
  localStorage.removeItem(storageKey);
  return value;
}
