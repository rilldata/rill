import { EmbedStore } from "@rilldata/web-common/features/embeds/embed-store.ts";
import { removeEmbedParams } from "@rilldata/web-admin/features/embeds/init-embed-public-api.ts";
import {
  decodeEmbedOpen,
  getEmbedOpenPath,
} from "@rilldata/web-admin/features/embeds/embed-open.ts";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
import { error, redirect } from "@sveltejs/kit";

export const load = async ({ url }) => {
  const embedStore = EmbedStore.getInstance();
  if (!embedStore) {
    EmbedStore.init(url);

    // Open links from report emails open the report instead of a dashboard.
    const open = url.searchParams.get("open");
    if (open) {
      const payload = decodeEmbedOpen(open);
      if (!payload) {
        throw error(400, m.embed_invalid_open_link());
      }
      throw redirect(307, getEmbedOpenPath(payload));
    }

    const resource = url.searchParams.get("resource");
    if (!resource) {
      throw redirect(307, "/-/embed");
    }

    const type =
      url.searchParams.get("type") === ResourceKind.Canvas
        ? "canvas"
        : "explore";
    // Retain non-embed search params
    const nonEmbedSearchParams = removeEmbedParams(url.searchParams);
    throw redirect(
      307,
      `/-/embed/${type}/${resource}?${nonEmbedSearchParams.toString()}`,
    );
  }

  const {
    instanceId,
    runtimeHost,
    accessToken,
    missingRequireParams,
    navigationEnabled,
    navigationBarEnabled,
    visibleExplores,
    adminToken,
    userEmail,
  } = embedStore;

  return {
    instanceId,
    runtimeHost,
    accessToken,
    missingRequireParams,
    navigationEnabled,
    navigationBarEnabled,
    visibleExplores,
    embedAdminActions: !!adminToken && !!userEmail,
  };
};
