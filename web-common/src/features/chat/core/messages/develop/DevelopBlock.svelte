<script lang="ts">
  import { builderActions, getAttrs } from "bits-ui";
  import * as Collapsible from "../../../../../components/collapsible";
  import type { DevelopBlock } from "@rilldata/web-common/features/chat/core/messages/develop/develop-block.ts";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import { IconButton } from "@rilldata/web-common/components/button";
  import { UndoIcon, PenIcon } from "lucide-svelte";
  import { createRuntimeServiceRestoreGitCommit } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store.ts";
  import FileDiffBlock from "@rilldata/web-common/features/chat/core/messages/file-diff/FileDiffBlock.svelte";

  export let block: DevelopBlock;

  let isExpanded = true; // Make this expanded by default

  $: console.log(block);
  $: ({ instanceId } = $runtime);

  const restoreCommit = createRuntimeServiceRestoreGitCommit();
  async function undoChanges() {
    if (!block.checkpointCommitHash) return;
    await $restoreCommit.mutateAsync({
      instanceId,
      commitSha: block.checkpointCommitHash,
      data: {},
    });
  }
</script>

<Collapsible.Root bind:open={isExpanded} class="w-full max-w-full self-start">
  <Collapsible.Trigger asChild let:builder>
    <div class="flex flex-row items-center">
      <button
        class="develop-header"
        {...getAttrs([builder])}
        use:builderActions={{ builders: [builder] }}
      >
        <div class="develop-icon">
          {#if isExpanded}
            <CaretDownIcon size="14" color="currentColor" />
          {:else}
            <PenIcon size="14px" />
          {/if}
        </div>
        <div class="develop-title">
          Made {block.diffs.length} change(s)
        </div>
      </button>
      {#if block.checkpointCommitHash}
        <IconButton on:click={undoChanges}>
          <UndoIcon size="14px" />
        </IconButton>
      {/if}
    </div>
  </Collapsible.Trigger>

  <Collapsible.Content class="w-full max-w-full">
    <div class="develop-content">
      {#each block.diffs as diff}
        <FileDiffBlock block={diff} />
      {/each}
    </div>
  </Collapsible.Content>
</Collapsible.Root>

<style lang="postcss">
  .develop-header {
    @apply w-full flex items-center gap-1.5 py-1;
    @apply bg-transparent border-none cursor-pointer;
    @apply text-xs text-gray-500 transition-colors;
  }

  .develop-header:hover {
    @apply text-gray-600;
  }

  .develop-icon {
    @apply flex items-center;
  }

  .develop-title {
    @apply flex-1 text-left font-normal;
  }

  .develop-content {
    @apply flex flex-col gap-y-2 py-1 text-xs leading-relaxed break-words;
  }

  .develop-content :global(*) {
    @apply text-gray-500;
  }

  .develop-content :global(strong),
  .develop-content :global(b) {
    @apply text-gray-600 font-semibold;
  }

  .develop-content :global(a) {
    @apply text-gray-600 underline;
  }
</style>
