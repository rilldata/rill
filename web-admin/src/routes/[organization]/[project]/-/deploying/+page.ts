import {
  TargetDashboardUrlParam,
  PreCommitShaUrlParam,
} from "@rilldata/web-common/features/project/deploy/utils.ts";

export const load = ({ url: { searchParams } }) => {
  const targetDashboard = searchParams.get(TargetDashboardUrlParam);
  const preCommitSha = searchParams.get(PreCommitShaUrlParam);

  return {
    targetDashboard,
    preCommitSha,
  };
};
