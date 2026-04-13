<script lang="ts">
  import { getRpcErrorMessage } from "@rilldata/web-admin/components/errors/error-utils";
  import { Button } from "@rilldata/web-common/components/button";
  import * as Popover from "@rilldata/web-common/components/popover";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import {
    createRuntimeServiceGitPushMutation,
    createRuntimeServiceGitStatus,
    type RpcStatus,
  } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";

  let commitMessage = "";
  let isCommitting = false;
  let open = false;

  const client = useRuntimeClient();
  const gitPushMutation = createRuntimeServiceGitPushMutation(client);

  // Invalidated by the file watcher on write/delete events; no polling needed
  const gitStatusQuery = createRuntimeServiceGitStatus(client, {});
  $: hasChanges = $gitStatusQuery.data?.localChanges ?? false;

  async function handleCommit() {
    if (!commitMessage.trim()) return;
    isCommitting = true;
    try {
      await $gitPushMutation.mutateAsync({
        commitMessage: commitMessage.trim(),
        force: false,
      });
      eventBus.emit("notification", {
        type: "success",
        message: "Changes committed",
      });
      commitMessage = "";
      open = false;
    } catch (err) {
      const message = getRpcErrorMessage(err as RpcStatus);
      eventBus.emit("notification", {
        type: "error",
        message: message ?? "Failed to commit changes",
      });
    } finally {
      isCommitting = false;
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      void handleCommit();
    }
  }
</script>

<Popover.Root bind:open>
  <Popover.Trigger>
    {#snippet child({ props })}
      <Button
        {...props}
        type="secondary"
        disabled={!hasChanges || isCommitting}
      >
        Commit
      </Button>
    {/snippet}
  </Popover.Trigger>
  <Popover.Content align="end">
    <div class="flex flex-col gap-y-2">
      <textarea
        class="commit-input"
        bind:value={commitMessage}
        onkeydown={handleKeydown}
        placeholder="Describe your changes..."
        rows="3"
      ></textarea>
      <Button
        type="primary"
        small
        disabled={!commitMessage.trim() || isCommitting}
        loading={isCommitting}
        loadingCopy="Committing..."
        onClick={handleCommit}
      >
        Commit
      </Button>
    </div>
  </Popover.Content>
</Popover.Root>

<style lang="postcss">
  .commit-input {
    @apply w-full text-xs px-2 py-1.5 rounded border border-gray-300 resize-none;
    @apply focus:outline-none focus:ring-1 focus:ring-primary-500 focus:border-primary-500;
  }
</style>
