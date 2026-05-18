import { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { derived } from "svelte/store";
import {
  createRuntimeServiceGitStatus,
  getRuntimeServiceGitStatusQueryKey,
  runtimeServiceGitStatus,
  type V1GitStatusResponse,
} from "@rilldata/web-common/runtime-client";
import type { QueryClient } from "@tanstack/query-core";

export function getDeploymentGithubStatus(
  runtimeClient: RuntimeClient,
  primaryBranch: string | undefined,
) {
  return derived(
    [
      createRuntimeServiceGitStatus(runtimeClient, {}),
      createRuntimeServiceGitStatus(runtimeClient, {
        remoteBranch: primaryBranch,
      }),
    ],
    ([currentBranchGitStatusResp, primaryBranchGitStatusResp]) => {
      const isPending =
        currentBranchGitStatusResp.isPending ||
        primaryBranchGitStatusResp.isPending;
      const error =
        currentBranchGitStatusResp.error || primaryBranchGitStatusResp.error;
      if (isPending || error) {
        return {
          isPending,
          error,
          data: {
            hasLocalChanges: false,
            hasChangesOnCurrent: false,
            hasRemoteChanges: false,
            hasLocalCommitsOnCurrent: false,
            alreadyOnPrimary: false,
            disabledPerGitStatus: true,
          },
        };
      }

      const currentBranch = currentBranchGitStatusResp.data?.branch ?? "";
      const hasChangesAgainstCurrent = Boolean(
        currentBranchGitStatusResp.data?.localCommits ||
          currentBranchGitStatusResp.data?.localChanges,
      );
      const hasChangesOnCurrent = Boolean(
        primaryBranchGitStatusResp.data?.localCommits ||
          primaryBranchGitStatusResp.data?.localChanges,
      );
      const hasLocalChanges = hasChangesAgainstCurrent || hasChangesOnCurrent;

      // GitPull cascades primary -> current's remote -> local, so collapsing
      // both signals into one is sufficient for the dialog flow.
      const hasRemoteChanges = Boolean(
        currentBranchGitStatusResp.data?.remoteCommits ||
          primaryBranchGitStatusResp.data?.remoteCommits,
      );
      const hasLocalCommitsOnCurrent = Boolean(
        currentBranchGitStatusResp.data?.localCommits,
      );

      const alreadyOnPrimary =
        !!primaryBranch && !!currentBranch && currentBranch === primaryBranch;

      const disabledPerGitStatus =
        !primaryBranch ||
        !currentBranch ||
        alreadyOnPrimary ||
        !hasLocalChanges;

      return {
        isPending: false,
        error: undefined,
        data: {
          hasLocalChanges,
          hasChangesOnCurrent,
          hasRemoteChanges,
          hasLocalCommitsOnCurrent,
          alreadyOnPrimary,
          disabledPerGitStatus,
        },
      };
    },
  );
}

export async function fetchDeploymentGithubStatusChanges(
  runtimeClient: RuntimeClient,
  queryClient: QueryClient,
  primaryBranch: string | undefined,
) {
  const currentBranchGitStatusResp = await queryClient.fetchQuery({
    queryKey: getRuntimeServiceGitStatusQueryKey(runtimeClient.instanceId, {}),
    queryFn: () => runtimeServiceGitStatus(runtimeClient, {}),
  });
  const hasChangesAgainstCurrent = Boolean(
    currentBranchGitStatusResp.localCommits ||
      currentBranchGitStatusResp.localChanges,
  );

  const primaryBranchGitStatusResp =
    queryClient.getQueryData<V1GitStatusResponse>(
      getRuntimeServiceGitStatusQueryKey(runtimeClient.instanceId, {
        remoteBranch: primaryBranch,
      }),
    );
  const hasChangesAgainstPrimary = Boolean(
    primaryBranchGitStatusResp?.localCommits ||
      primaryBranchGitStatusResp?.localChanges,
  );

  return hasChangesAgainstCurrent || hasChangesAgainstPrimary;
}
