import { EmbedStorageNamespacePrefix } from "@rilldata/web-admin/features/embeds/constants.ts";
import { clearExploreSessionStore } from "@rilldata/web-common/features/dashboards/state-managers/loaders/explore-web-view-store.ts";

export const load = async ({ params, parent }) => {
  const exploreName = params.name;
  const { visibleExplores } = await parent();

  // Check visibleExplores for more details
  if (!visibleExplores.has(exploreName)) {
    clearExploreSessionStore(exploreName, EmbedStorageNamespacePrefix);
    visibleExplores.add(exploreName);
  }

  return { exploreName };
};
