import { EmbedStore } from "@rilldata/web-common/features/embeds/embed-store.ts";
import { removeEmbedParams } from "@rilldata/web-admin/features/embeds/init-embed-public-api.ts";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
import { redirect } from "@sveltejs/kit";

export const load = ({ url }) => {
  const embedStore = EmbedStore.getInstance();
  if (!embedStore) {
    EmbedStore.init(url);

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
    visibleExplores,
  } = embedStore;

  return {
    instanceId,
    runtimeHost,
    accessToken,
    missingRequireParams,
    navigationEnabled,
    visibleExplores,
  };
};
