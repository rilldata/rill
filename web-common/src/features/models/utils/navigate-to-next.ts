import { page } from "$app/stores";
import { get } from "svelte/store";

export function getNextRoute(assetRoutes: string[]) {
  const currentPage = get(page);
  const currentPageIndex = assetRoutes.indexOf(currentPage.url.pathname);

  if (assetRoutes.length <= 1 || currentPageIndex === -1) {
    return "/";
  }

  if (currentPageIndex === assetRoutes.length - 1) {
    return assetRoutes[0];
  }

  return assetRoutes[currentPageIndex + 1];
}
