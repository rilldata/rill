import type { Reroute } from "@sveltejs/kit";
import { deLocalizeUrl } from "@rilldata/web-common/lib/i18n/gen/runtime";

export const reroute: Reroute = (request) => {
  return deLocalizeUrl(request.url).pathname;
};
