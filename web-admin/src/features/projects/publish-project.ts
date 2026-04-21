import { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { runtimeServiceGitPush } from "@rilldata/web-common/runtime-client";
import { DeployingDashboardUrlParam } from "@rilldata/web-common/features/project/deploy/utils.ts";

export const CreateProjectDevBranchName = "dev";

export async function publishProjectAndRedirect(
  runtimeClient: RuntimeClient,
  organization: string,
  project: string,
  generatedDashboard?: string,
) {
  // Push the initial commit to the current branch.
  await runtimeServiceGitPush(runtimeClient, {
    commitMessage: "Initial dashboard commit",
  });

  // TODO: push env once that API is ready.

  // TODO: push changes from dev branch to main once env is pushed.

  // TODO: create primary deployment after merging to main branch.

  // TODO: land user to edit screen when that is available
  return `/${organization}/${project}/@${CreateProjectDevBranchName}/-/deploying?${DeployingDashboardUrlParam}=${generatedDashboard ?? ""}`;
}
