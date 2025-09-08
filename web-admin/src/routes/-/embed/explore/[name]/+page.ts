import { EmbedStore } from "@rilldata/web-admin/features/embeds/embed-store.ts";
import { EmbedStorageNamespacePrefix } from "@rilldata/web-admin/features/embeds/constants.ts";
import { clearExploreSessionStore } from "@rilldata/web-common/features/dashboards/state-managers/loaders/explore-web-view-store.ts";

export const load = ({ params }) => {
  const exploreName = params.name;

  const embedStore = EmbedStore.getInstance();
  if (!embedStore) {
    return { exploreName };
  }

  // Clean session storage for dashboards that are navigated to for the 1st time.
  // This way once the page is loaded, the dashboard state is persisted.
  // But the moment the user moves away to another page within the parent page, then it will be cleared next time the user comes back to the dashboard.
  if (!embedStore.exploreSeen.has(exploreName)) {
    clearExploreSessionStore(exploreName, EmbedStorageNamespacePrefix);
    embedStore.exploreSeen.add(exploreName);
  }

  return { exploreName };
};
