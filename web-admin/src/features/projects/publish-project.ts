import { goto } from "$app/navigation";
import { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { runtimeServiceGitPush } from "@rilldata/web-common/runtime-client";

export const CreateProjectBranchName = "dev";

/**
 * Checkpoints the current project state and redirects to the project dashboard.
 * Will be the cloud editing screen once it's available.
 *
 * Note that publishing the project to prod deployment will be through explicit user action.
 */
export async function checkpointProjectAndRedirect(
  runtimeClient: RuntimeClient,
  organization: string,
  project: string,
) {
  // Push the initial commit to the current branch.
  await runtimeServiceGitPush(runtimeClient, {
    commitMessage: "Initial project setup",
  });

  // TODO: land user to edit screen when that is available
  const destinationPath = `/${organization}/${project}`;
  // Without this delay, the navigation is getting cancelled.
  setTimeout(() => void goto(destinationPath), 50);
}

// TODO: add the publish function that,
// 1. Pushes the env through a new API yet to be added.
// 2. Merges the changes to the primary branch.
// 2. Creates a prod deployment if not present.
// 3. Deletes the dev deployment and the remote branch.
