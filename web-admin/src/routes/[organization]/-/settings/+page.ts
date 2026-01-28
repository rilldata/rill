import { getSingleUseUrlParam } from "@rilldata/web-admin/features/navigation/getSingleUseUrlParam";
import type { PageLoad } from "./$types";

export const load: PageLoad = async ({ url, parent }) => {
  const showUpgradeDialog = !!getSingleUseUrlParam(
    url as URL,
    "upgrade",
    "rill:app:showUpgrade",
  );
  const {
    organizationLogoUrl,
    organizationLogoDarkUrl,
    organizationFaviconUrl,
  } = (await parent()) as {
    organizationLogoUrl?: string;
    organizationLogoDarkUrl?: string;
    organizationFaviconUrl?: string;
  };
  return {
    showUpgradeDialog,
    organizationLogoUrl,
    organizationLogoDarkUrl,
    organizationFaviconUrl,
  };
};
