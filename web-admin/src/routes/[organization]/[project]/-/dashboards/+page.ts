import { DeployingDashboardUrlParam } from "@rilldata/web-common/features/project/deploy/utils.ts";

export const load = ({ url: { searchParams } }) => {
  const deploying = searchParams.get("deploying");
  const deployingDashboard = searchParams.get(DeployingDashboardUrlParam);

  return {
    deploying,
    deployingDashboard,
  };
};
