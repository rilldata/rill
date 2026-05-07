import {
  DeployingDashboardUrlParam,
  PreCommitShaUrlParam,
} from "@rilldata/web-common/features/project/deploy/utils.ts";

export const load = ({ url: { searchParams } }) => {
  const deployingDashboard = searchParams.get(DeployingDashboardUrlParam);
  const preCommitSha = searchParams.get(PreCommitShaUrlParam);

  return {
    deployingDashboard,
    preCommitSha,
  };
};
