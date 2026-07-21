import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  getRuntimeServiceGitDiffQueryKey,
  getRuntimeServiceGitStatusQueryKey,
  runtimeServiceGitRevert,
} from "@rilldata/web-common/runtime-client";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";

/**
 * revertFiles discards local changes for the given subpath-relative paths (empty paths reverts all
 * changed files), then invalidates the GitDiff and GitStatus queries so the changed-files list and
 * the publish button's enabled state refresh. It returns the number of files that were reverted.
 */
export async function revertFiles(
  client: RuntimeClient,
  remoteBranch: string | undefined,
  paths: string[],
): Promise<number> {
  const resp = await runtimeServiceGitRevert(client, { remoteBranch, paths });

  // Both queries are cached under several parameter variants (current branch and primary branch for
  // GitStatus; list and full-diff for GitDiff). Invalidate by the shared key prefix so every variant
  // refetches. The first three key segments are ["RuntimeService", "<method>", instanceId].
  const diffPrefix = getRuntimeServiceGitDiffQueryKey(client.instanceId).slice(
    0,
    3,
  );
  const statusPrefix = getRuntimeServiceGitStatusQueryKey(
    client.instanceId,
  ).slice(0, 3);
  await Promise.all([
    queryClient.invalidateQueries({ queryKey: diffPrefix }),
    queryClient.invalidateQueries({ queryKey: statusPrefix }),
  ]);

  return resp.revertedPaths?.length ?? 0;
}
