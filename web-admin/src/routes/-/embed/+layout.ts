import { EmbedStore } from "@rilldata/web-admin/features/embeds/embed-store.ts";
import { redirect } from "@sveltejs/kit";

export const load = ({ url }) => {
  const embedStore = EmbedStore.getInstance();
  if (!embedStore) {
    const redirectUrl = EmbedStore.init(url);
    throw redirect(307, redirectUrl);
  }

  const { instanceId, runtimeHost, accessToken, navigationEnabled } =
    embedStore;

  return {
    instanceId,
    runtimeHost,
    accessToken,
    navigationEnabled,
  };
};
