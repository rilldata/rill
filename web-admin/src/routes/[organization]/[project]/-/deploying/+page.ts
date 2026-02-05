import { DeployingDashboardUrlParam } from "@rilldata/web-common/features/project/deploy/utils.ts";

export const load = ({ url: { searchParams } }) => {
  const deployingDashboard = searchParams.get(DeployingDashboardUrlParam);

  return {
    deployingDashboard,
  };
};
