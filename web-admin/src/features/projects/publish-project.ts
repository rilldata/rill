import { goto } from "$app/navigation";
import { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { runtimeServiceGitPush } from "@rilldata/web-common/runtime-client";

export async function publishProjectAndRedirect(
  runtimeClient: RuntimeClient,
  organization: string,
  project: string,
) {
  // Push the initial commit to the current branch.
  await runtimeServiceGitPush(runtimeClient, {
    commitMessage: "Initial dashboard commit",
  });

  // TODO: push env once that API is ready.

  // TODO: push changes from dev branch to main once env is pushed.

  // TODO: create primary deployment after merging to main branch.

  // TODO: land user to edit screen when that is available
  const destinationPath = `/${organization}/${project}`;
  // Without this delay, the navigation is getting cancelled.
  setTimeout(() => void goto(destinationPath), 50);
}
