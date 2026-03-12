<script lang="ts">
  import { getRpcErrorMessage } from "@rilldata/web-admin/components/errors/error-utils";
  import { branchPathPrefix } from "@rilldata/web-admin/features/branches/branch-utils";
  import { Button } from "@rilldata/web-common/components/button";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import {
    createRuntimeServiceGitPushMutation,
    type RpcStatus,
  } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";

  export let organization: string;
  export let project: string;
  export let branch: string;

  let isCommitting = false;

  const client = useRuntimeClient();
  const gitPushMutation = createRuntimeServiceGitPushMutation(client);

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

  function handleClose() {
    // Full page navigation avoids a race where useRuntimeClient() is called
    // before the project layout's RuntimeProvider remounts.
    window.location.href = `/${organization}/${project}${branchPathPrefix(branch)}`;
  }
</script>

<Button type="secondary" disabled={isCommitting} onClick={handleClose}>
  Close editor
</Button>
<Button
  type="primary"
  disabled={isCommitting}
  loading={isCommitting}
  loadingCopy="Pushing..."
  onClick={handleCommit}
>
  Push to production
</Button>
