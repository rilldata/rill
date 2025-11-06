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

  $: ({ reportSource } = $data);
  $: hasSelection = Boolean(
    reportSource?.metricsViewName || reportSource?.exploreName,
  );

  const canvasSourceOptionsStore = getPrimaryCanvasSourceOptions();
  $: hasCanvasSourceOptions = $canvasSourceOptionsStore.length > 0;
  const exploreSourceOptionsStore = getPrimaryExploreSourceOptions();
  $: hasExploreSourceOptions = $exploreSourceOptionsStore.length > 0;

  function onSelect(source: ReportSource) {
    $data.reportSource = source;
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
        <span>{reportSource.label}</span>
      {:else}
        <span class="text-muted-foreground">Select a source...</span>
      {/if}
    </button>
  </Dropdown.Trigger>
  <Dropdown.Content align="start" class="w-[250px]">
    {#if hasCanvasSourceOptions}
      <Dropdown.Label>Canvases</Dropdown.Label>

      {#each $canvasSourceOptionsStore as source, i (i)}
        <ReportSourceCanvasSubSelector
          selectedSource={reportSource}
          canvasNameForSubSelector={source.canvasName}
          {onSelect}
        />
      {/each}
    {/if}

    {#if hasExploreSourceOptions}
      {#if hasCanvasSourceOptions}
        <Dropdown.Separator />
      {/if}
      <Dropdown.Label>Metrics explores</Dropdown.Label>
      {#each $exploreSourceOptionsStore as source, i (i)}
        <ReportSourceItem selectedSource={reportSource} {source} {onSelect} />
      {/each}
    {/if}
  </Dropdown.Content>
</Dropdown.Root>
