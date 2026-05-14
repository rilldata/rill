import { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { derived } from "svelte/store";
import { createRuntimeServiceGitStatus } from "@rilldata/web-common/runtime-client";

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
      const hasChangesAgainstPrimary = Boolean(
        primaryBranchGitStatusResp.data?.localCommits ||
          primaryBranchGitStatusResp.data?.localChanges,
      );
      const hasLocalChanges =
        hasChangesAgainstCurrent || hasChangesAgainstPrimary;

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
          alreadyOnPrimary,
          disabledPerGitStatus,
        },
      };
    },
  );
}
