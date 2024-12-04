import { getSingleUseUrlParam } from "@rilldata/web-admin/features/navigation/getSingleUseUrlParam";
import type { PageLoad } from "./$types";

export const load: PageLoad = async ({ params: { organization }, url }) => {
  const showUpgradeDialog = !!getSingleUseUrlParam(
    url,
    "upgrade",
    "rill:app:showUpgrade",
  );
  return {
    organization,
    showUpgradeDialog,
  };
};
