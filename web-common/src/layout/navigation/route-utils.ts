import type { Page } from "@sveltejs/kit";

export function isDeployPage(page: Page) {
  return page.route.id === "/(misc)/deploy";
}
