import { getSingleUseUrlParam } from "@rilldata/web-admin/features/navigation/getSingleUseUrlParam";
import type { PageLoad } from "./$types";

export const load: PageLoad = async ({ url }) => {
  const showUpgradeDialog = !!getSingleUseUrlParam(
    url,
    "upgrade",
    "rill:app:showUpgrade",
  );
  return {
    showUpgradeDialog,
  };
};
