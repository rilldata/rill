<script lang="ts">
  import { getRpcErrorMessage } from "@rilldata/web-admin/components/errors/error-utils";
  import { branchPathPrefix } from "@rilldata/web-admin/features/branches/branch-utils";
  import { Button } from "@rilldata/web-common/components/button";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import {
    createRuntimeServiceGitPushMutation,
    type RpcStatus,
  } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { RocketIcon } from "lucide-svelte";

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
        message: "Changes merged to production",
      });
    } catch (err) {
      const message = getRpcErrorMessage(err as RpcStatus);
      eventBus.emit("notification", {
        type: "error",
        message: message ?? "Failed to merge changes",
      });
    } finally {
      isCommitting = false;
    }
  }

  $: closeHref = `/${organization}/${project}${branchPathPrefix(branch)}`;

  function handleClose(e: MouseEvent) {
    // Full page navigation avoids a race where useRuntimeClient() is called
    // before the project layout's RuntimeProvider remounts.
    e.preventDefault();
    window.location.href = closeHref;
  }
</script>

<Tooltip distance={8}>
  <Button
    type="secondary"
    href={closeHref}
    disabled={isCommitting}
    onClick={handleClose}
  >
    Done
  </Button>
  <TooltipContent slot="tooltip-content" maxWidth="200px">
    <span class="text-xs">Return to project home</span>
  </TooltipContent>
</Tooltip>
<Button
  type="primary"
  disabled={isCommitting}
  loading={isCommitting}
  loadingCopy="Merging..."
  onClick={handleCommit}
>
  <RocketIcon size="14" />
  Merge to production
</Button>
