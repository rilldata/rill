<!--
  Renders a checkpoint block with collapsible tool call header.
  Shows the commit SHA and provides a revert button.
-->
<script lang="ts">
  import { createMutation } from "@tanstack/svelte-query";
  import type { V1Tool } from "../../../../../runtime-client";
  import { runtimeServiceRevertToCommit } from "../../../../../runtime-client";
  import { runtime } from "../../../../../runtime-client/runtime-store";
  import Button from "../../../../../components/button/Button.svelte";
  import ToolCall from "../tools/ToolCall.svelte";
  import type { CheckpointBlock } from "./checkpoint-block";

  export let block: CheckpointBlock;
  export let tools: V1Tool[] | undefined = undefined;

  let isReverting = false;
  let revertError: string | undefined = undefined;
  let revertSuccess = false;

  // Shorten SHA for display (first 7 characters)
  $: shortSha = block.commitSha.substring(0, 7);

  // Check if this is an initial commit (we can't easily revert those)
  $: isInitialCommit = block.commitMessage.includes("Initial commit");

  // Create mutation for reverting
  const revertMutation = createMutation({
    mutationFn: async () => {
      const response = await runtimeServiceRevertToCommit(
        $runtime.instanceId,
        block.commitSha,
      );
      return response.data;
    },
    onSuccess: () => {
      revertSuccess = true;
      revertError = undefined;
      // Clear success message after 5 seconds
      setTimeout(() => {
        revertSuccess = false;
      }, 5000);
    },
    onError: (error: Error) => {
      revertError = error.message || "Failed to revert commit";
      revertSuccess = false;
    },
  });

  function handleRevert() {
    if (isReverting || revertSuccess) return;
    revertError = undefined;
    $revertMutation.mutate();
  }

  $: isReverting = $revertMutation.isPending;
</script>

<div class="checkpoint-block">
  <ToolCall
    message={block.message}
    resultMessage={block.resultMessage}
    {tools}
    variant="block"
  />

  <div class="checkpoint-container">
    <div class="checkpoint-header">
      <div class="checkpoint-info">
        <div class="checkpoint-sha" title={block.commitSha}>
          <span class="sha-label">Commit:</span>
          <code class="sha-value">{shortSha}</code>
        </div>
        <div class="checkpoint-message">{block.commitMessage}</div>
        {#if !block.hadChanges}
          <div class="checkpoint-status">No changes committed</div>
        {/if}
      </div>
      <div class="checkpoint-actions">
        {#if !isInitialCommit}
          <Button
            type="secondary"
            onClick={handleRevert}
            disabled={isReverting || revertSuccess}
          >
            {isReverting
              ? "Reverting..."
              : revertSuccess
                ? "Reverted"
                : "Revert to this checkpoint"}
          </Button>
        {:else}
          <span class="initial-commit-note">
            Initial checkpoint (cannot revert)
          </span>
        {/if}
      </div>
    </div>

    {#if revertSuccess}
      <div class="revert-success">
        Successfully reverted to commit {shortSha}
      </div>
    {/if}

    {#if revertError}
      <div class="revert-error">
        {revertError}
      </div>
    {/if}
  </div>
</div>

<style lang="postcss">
  .checkpoint-block {
    @apply w-full max-w-full self-start;
  }

  .checkpoint-container {
    @apply border border-gray-200 rounded-md overflow-hidden;
  }

  .checkpoint-header {
    @apply flex items-center justify-between gap-4 px-4 py-3;
    @apply bg-gray-50;
  }

  .checkpoint-info {
    @apply flex flex-col gap-1;
  }

  .checkpoint-sha {
    @apply flex items-center gap-2;
  }

  .sha-label {
    @apply text-xs text-gray-500 font-medium;
  }

  .sha-value {
    @apply text-xs font-mono bg-gray-100 px-1.5 py-0.5 rounded;
    @apply text-gray-700;
  }

  .checkpoint-message {
    @apply text-sm text-gray-600;
  }

  .checkpoint-status {
    @apply text-xs text-gray-500 italic;
  }

  .checkpoint-actions {
    @apply flex-shrink-0;
  }

  .initial-commit-note {
    @apply text-xs text-gray-500 italic;
  }

  .revert-success {
    @apply px-4 py-2 text-sm;
    @apply bg-green-50 text-green-700 border-t border-green-200;
  }

  .revert-error {
    @apply px-4 py-2 text-sm;
    @apply bg-red-50 text-red-700 border-t border-red-200;
  }
</style>
