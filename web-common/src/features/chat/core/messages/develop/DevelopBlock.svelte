<script lang="ts">
  import { builderActions, getAttrs } from "bits-ui";
  import * as Collapsible from "../../../../../components/collapsible";
  import {
    type DevelopBlock,
    getGenerateCTAs,
  } from "@rilldata/web-common/features/chat/core/messages/develop/develop-block.ts";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import { PenIcon } from "lucide-svelte";
  import { createRuntimeServiceRestoreGitCommit } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store.ts";
  import FileDiffBlock from "@rilldata/web-common/features/chat/core/messages/file-diff/FileDiffBlock.svelte";
  import { sidebarActions } from "@rilldata/web-common/features/chat/layouts/sidebar/sidebar-store.ts";
  import { getCommitExists } from "@rilldata/web-common/features/project/commit-utils.ts";

  export let block: DevelopBlock;

  let isExpanded = true; // Make this expanded by default

  $: ({ instanceId } = $runtime);
  $: ctas = getGenerateCTAs(instanceId, block);
  $: console.log($ctas);
  $: hasSomeCTA = $ctas.models.length > 0 || $ctas.metricsViews.length > 0;

  // Check if the checkpoint commit exists. The tree might have changed since this chat.
  $: checkpointCommitExistsQuery = getCommitExists(block.checkpointCommitHash);
  $: checkpointCommitExists = !!$checkpointCommitExistsQuery.data;

  const restoreCommit = createRuntimeServiceRestoreGitCommit();
  async function undoChanges() {
    if (!block.checkpointCommitHash) return;
    await $restoreCommit.mutateAsync({
      instanceId,
      commitSha: block.checkpointCommitHash,
      data: {},
    });
  }

  function promptGenerateMetricsView(model: string) {
    sidebarActions.startChat(`Generate a metrics view for ${model}`);
  }

  function promptGenerateExplore(metricsView: string) {
    sidebarActions.startChat(`Generate an explore for ${metricsView}`);
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
            <CaretDownIcon size="14" />
          {:else}
            <PenIcon size="14px" />
          {/if}
        </div>
        <div class="develop-title">
          Made {block.diffs.length} change(s)
        </div>
      </button>
      {#if checkpointCommitExists}
        <Button onClick={undoChanges} noStroke>Undo</Button>
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

{#if hasSomeCTA}
  {@const hasOnlyOneModel = $ctas.models.length === 1}
  {@const hasOnlyOneMetricsView = $ctas.metricsViews.length === 1}
  <div class="develop-cta">
    {#each $ctas.models as model (model)}
      <Button onClick={() => promptGenerateMetricsView(model)}>
        {#if hasOnlyOneModel}
          Generate a metrics view
        {:else}
          Generate metrics view for {model}
        {/if}
      </Button>
    {/each}

    {#each $ctas.metricsViews as metricsView (metricsView)}
      <Button onClick={() => promptGenerateExplore(metricsView)}>
        {#if hasOnlyOneMetricsView}
          Generate an explore dashboard
        {:else}
          Generate explore dashboard for {metricsView}
        {/if}
      </Button>
    {/each}
  </div>
{/if}

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

  .develop-cta {
    @apply flex flex-row items-center gap-y-1;
  }
</style>
