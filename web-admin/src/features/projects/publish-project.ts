import { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { runtimeServiceGitPush } from "@rilldata/web-common/runtime-client";

export const CreateProjectBranchName = "dev";

/**
 * Checkpoints the current project state. Redirect should already be handled through indiviual components.
 *
 * Note that publishing the project to prod deployment will be through explicit user action.
 */
export async function checkpointProject(runtimeClient: RuntimeClient) {
  // Push the initial commit to the current branch.
  await runtimeServiceGitPush(runtimeClient, {
    commitMessage: "Initial project setup",
  });
}

// TODO: add the publish function that,
// 1. Pushes the env through a new API yet to be added.
// 2. Merges the changes to the primary branch.
// 2. Creates a prod deployment if not present.
// 3. Deletes the dev deployment and the remote branch.
