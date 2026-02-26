<script lang="ts">
  import { goto } from "$app/navigation";
  import {
    createAdminServiceDeleteDeployment,
    getAdminServiceListDeploymentsQueryKey,
  } from "@rilldata/web-admin/client";
  import { getRpcErrorMessage } from "@rilldata/web-admin/components/errors/error-utils";
  import { Button } from "@rilldata/web-common/components/button";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import {
    createRuntimeServiceGitPush,
    type RpcStatus,
  } from "@rilldata/web-common/runtime-client";

  export let organization: string;
  export let project: string;
  export let deploymentId: string;
  export let instanceId: string;

  let isCommitting = false;
  let isDiscarding = false;
  let commitError: string | null = null;

  const gitPushMutation = createRuntimeServiceGitPush();
  const deleteMutation = createAdminServiceDeleteDeployment();

  async function handleCommit() {
    isCommitting = true;
    commitError = null;
    try {
      await $gitPushMutation.mutateAsync({
        instanceId,
        data: {
          commitMessage: "Changes from Rill Cloud edit session",
          force: false,
        },
      });
      eventBus.emit("notification", {
        type: "success",
        message: "Changes pushed to production",
      });
    } catch (err) {
      const message = getRpcErrorMessage(err as RpcStatus);
      if (message?.includes("diverged") || message?.includes("conflict")) {
        commitError =
          "Push failed: the remote branch has diverged. Pull the latest changes locally and resolve conflicts.";
      } else {
        commitError = message ?? "Failed to push changes";
      }
      eventBus.emit("notification", {
        type: "error",
        message: commitError,
      });
    } finally {
      isCommitting = false;
    }
  }

  async function handleDiscard() {
    isDiscarding = true;
    try {
      await $deleteMutation.mutateAsync({ deploymentId });
      void queryClient.invalidateQueries({
        queryKey: getAdminServiceListDeploymentsQueryKey(
          organization,
          project,
          {
            environment: "dev",
          },
        ),
      });
      await goto(`/${organization}/${project}`);
    } catch (err) {
      eventBus.emit("notification", {
        type: "error",
        message: `Failed to end edit session: ${getRpcErrorMessage(err as RpcStatus)}`,
      });
    } finally {
      isDiscarding = false;
    }
  }
</script>

<div class="toolbar">
  <div class="toolbar-left">
    <span class="label">Editing</span>
    <span class="project-name">{project}</span>
  </div>
  <div class="toolbar-right">
    {#if commitError}
      <span class="error-text">{commitError}</span>
    {/if}
    <Button
      type="secondary"
      disabled={isCommitting || isDiscarding}
      loading={isDiscarding}
      loadingCopy="Ending..."
      onClick={handleDiscard}
    >
      End session
    </Button>
    <Button
      type="primary"
      disabled={isCommitting || isDiscarding}
      loading={isCommitting}
      loadingCopy="Pushing..."
      onClick={handleCommit}
    >
      Push to production
    </Button>
  </div>
</div>

<style lang="postcss">
  .toolbar {
    @apply flex items-center justify-between;
    @apply px-4 py-2 border-b;
    @apply bg-surface-base;
  }

  .toolbar-left {
    @apply flex items-center gap-2;
  }

  .label {
    @apply text-xs font-medium text-gray-500 uppercase tracking-wide;
  }

  .project-name {
    @apply text-sm font-semibold text-gray-800;
  }

  .toolbar-right {
    @apply flex items-center gap-3;
  }

  .error-text {
    @apply text-xs text-red-600 max-w-md truncate;
  }
</style>
