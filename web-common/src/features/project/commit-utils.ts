import { createRuntimeServiceListGitCommits } from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store.ts";
import { get } from "svelte/store";

/**
 * ListGitCommits API uses a commit hash to list starting from that hash. This is passed to pageToken param.
 * We use it to check if a commit has exists by fetching for it specifically.
 * @param commitHash
 */
export function getCommitExists(commitHash: string) {
  const instanceId = get(runtime).instanceId;

  return createRuntimeServiceListGitCommits(
    instanceId,
    {
      pageSize: 1,
      pageToken: commitHash,
    },
    {
      query: {
        select: (data) => data.commits?.[0].commitSha === commitHash,
        enabled: !!commitHash,
      },
    },
  );
}
