import { redirect } from "@sveltejs/kit";

export function getSingleUseUrlParam(
  url: URL,
  param: string,
  storageKey: string,
) {
  // Save the state in localStorage and redirect to the url without it.
  // This prevents a refresh or saving the url from re-triggering in the page
  const paramValue = url.searchParams.get(param);
  if (paramValue) {
    try {
      localStorage.setItem(storageKey, paramValue);
    } catch {
      // no-op
    }
    const redirectUrl = new URL(url);
    redirectUrl.searchParams.delete(param);
    throw redirect(307, redirectUrl.pathname + redirectUrl.search);
  }

  const localStorageValue = localStorage.getItem(storageKey);
  localStorage.removeItem(storageKey);
  return localStorageValue;
}
