<script lang="ts">
  import { goto } from "$app/navigation";
  import {
    createAdminServiceDeleteDeployment,
    getAdminServiceListDeploymentsQueryKey,
  } from "@rilldata/web-admin/client";
  import { getRpcErrorMessage } from "@rilldata/web-admin/components/errors/error-utils";
  import { requestSkipBranchInjection } from "@rilldata/web-admin/features/branches/branch-utils";
  import { Button } from "@rilldata/web-common/components/button";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import {
    createRuntimeServiceGitPushMutation,
    type RpcStatus,
  } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { editSessionState } from "./use-edit-session";

  let isCommitting = false;
  let isDiscarding = false;

  const client = useRuntimeClient();
  const gitPushMutation = createRuntimeServiceGitPushMutation(client);
  const deleteMutation = createAdminServiceDeleteDeployment();

  async function handleCommit() {
    isCommitting = true;
    try {
      await $gitPushMutation.mutateAsync({
        commitMessage: "Changes from Rill Cloud edit session",
        force: false,
      });
      eventBus.emit("notification", {
        type: "success",
        message: "Changes pushed to production",
      });
    } catch (err) {
      const message = getRpcErrorMessage(err as RpcStatus);
      eventBus.emit("notification", {
        type: "error",
        message: message ?? "Failed to push changes",
      });
    } finally {
      isCommitting = false;
    }
  }

  async function handleDiscard() {
    if (!$editSessionState) return;
    const { deploymentId, organization, project } = $editSessionState;

    isDiscarding = true;
    try {
      await $deleteMutation.mutateAsync({ deploymentId });
      void queryClient.invalidateQueries({
        queryKey: getAdminServiceListDeploymentsQueryKey(
          organization,
          project,
          { environment: "dev" },
        ),
      });
      requestSkipBranchInjection();
      await goto(`/${organization}/${project}`);
    } catch (err) {
      eventBus.emit("notification", {
        type: "error",
        message: `Failed to end session: ${getRpcErrorMessage(err as RpcStatus)}`,
      });
    } finally {
      isDiscarding = false;
    }
  }
</script>

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
