<script lang="ts">
  import * as Dropdown from "@rilldata/web-common/components/dropdown-menu";
  import ReportSourceCanvasSubSelector from "@rilldata/web-common/features/scheduled-reports/report-source/ReportSourceCanvasSubSelector.svelte";
  import ReportSourceItem from "@rilldata/web-common/features/scheduled-reports/report-source/ReportSourceItem.svelte";
  import {
    getPrimaryCanvasSourceOptions,
    getPrimaryExploreSourceOptions,
    type ReportSource,
  } from "@rilldata/web-common/features/scheduled-reports/report-source/utils.ts";
  import type { ReportValues } from "@rilldata/web-common/features/scheduled-reports/utils.ts";
  import { builderActions, getAttrs } from "bits-ui";
  import type { Readable } from "svelte/store";

  export let data: Readable<ReportValues>;

  $: ({ metricsViewName, exploreName, canvasName } = $data);
  $: hasSelection = Boolean(metricsViewName || exploreName);

  const exploreSourceOptionsStore = getPrimaryExploreSourceOptions();
  const canvasSourceOptionsStore = getPrimaryCanvasSourceOptions();

  function onSelect(source: ReportSource) {
    $data.metricsViewName = source.metricsViewName;
    $data.exploreName = source.exploreName;
    $data.canvasName = source.canvasName;
  }
</script>

<Dropdown.Root>
  <Dropdown.Trigger asChild let:builder>
    <button
      class="flex h-8 w-full items-center rounded-[2px] border border-gray-300 bg-transparent px-2 py-2 text-xs ring-offset-background focus:outline-none focus:border-primary-400"
      {...getAttrs([builder])}
      use:builderActions={{ builders: [builder] }}
    >
      {#if hasSelection}
        <span>{exploreName || metricsViewName}</span>
      {:else}
        <span class="text-muted-foreground">Select a source...</span>
      {/if}
    </button>
  </Dropdown.Trigger>
  <Dropdown.Content align="start" class="w-[250px]">
    {#each $exploreSourceOptionsStore as source, i (i)}
      <ReportSourceItem
        {metricsViewName}
        {exploreName}
        {canvasName}
        {source}
        {onSelect}
      />
    {/each}

    {#each $canvasSourceOptionsStore as source, i (i)}
      <ReportSourceCanvasSubSelector
        {data}
        canvasNameForSubSelector={source.canvasName}
        {onSelect}
      />
    {/each}
  </Dropdown.Content>
</Dropdown.Root>
