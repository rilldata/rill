import { goto } from "$app/navigation";
import { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { adminServiceCreateDeployment } from "@rilldata/web-admin/client";
import {
  runtimeServiceGitMergeToBranch,
  runtimeServiceGitPush,
} from "@rilldata/web-common/runtime-client";

export async function completeInitialAddData(
  runtimeClient: RuntimeClient,
  org: string,
  project: string,
  skipCommit?: boolean,
) {
  if (skipCommit) {
    // Push the initial commit to the current branch.
    await runtimeServiceGitPush(runtimeClient, {
      commitMessage: "Initial dashboard commit",
    });
  }
  // Then push the changes to the main branch.
  await runtimeServiceGitMergeToBranch(runtimeClient, {
    branch: "main", // TODO: get primary branch
  });

  // Create a new deployment for the project.
  await adminServiceCreateDeployment(org, project, {
    environment: "prod",
  });

  return goto(`/${org}/${project}`);
}
