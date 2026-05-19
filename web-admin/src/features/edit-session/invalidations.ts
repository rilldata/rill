import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
import { getRuntimeServiceGitStatusQueryKey } from "@rilldata/web-common/runtime-client";
import { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";

export function invalidateGitStatusQueries(
  runtimeClient: RuntimeClient,
  primaryBranch: string | undefined,
) {
  // gitStatus tracks localChanges and currentBranch; mutations below change
  // both, so refresh after every flow. We cache both an empty-remoteBranch
  // and a primary-branch keyed query (see `getDeploymentGithubStatus`), so
  // invalidate both.
  void queryClient.invalidateQueries({
    queryKey: getRuntimeServiceGitStatusQueryKey(runtimeClient.instanceId, {}),
  });
  if (primaryBranch) {
    void queryClient.invalidateQueries({
      queryKey: getRuntimeServiceGitStatusQueryKey(runtimeClient.instanceId, {
        remoteBranch: primaryBranch,
      }),
    });
  }
}
