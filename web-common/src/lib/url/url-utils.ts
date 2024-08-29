import { page } from "$app/stores";
import { get } from "svelte/store";

export function getUrlForPath(path: string, retainParams = ["features"]): URL {
  const url = get(page).url;
  if (!path.startsWith("/")) path = "/" + path;
  const newUrl = new URL(`${url.protocol}//${url.host}${path}`);

  for (const param of retainParams) {
    const value = url.searchParams.get(param);
    if (!value) continue;
    newUrl.searchParams.set(param, value);
  }
  return newUrl;
}

export function getFullUrlForPath(
  path: string,
  retainParams = ["features"],
): string {
  const newUrl = getUrlForPath(path, retainParams);

  if (newUrl.search !== "") {
    return `${newUrl.pathname}${newUrl.search}`;
  }
  return newUrl.pathname;
}
