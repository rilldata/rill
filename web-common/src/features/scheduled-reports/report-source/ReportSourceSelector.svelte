<script lang="ts">
  import * as Dropdown from "@rilldata/web-common/components/dropdown-menu";
  import ReportSourceCanvasSubSelector from "@rilldata/web-common/features/scheduled-reports/report-source/ReportSourceCanvasSubSelector.svelte";
  import {
    getPrimaryCanvasSourceOptions,
    getPrimaryExploreSourceOptions,
    type ReportSource,
  } from "@rilldata/web-common/features/scheduled-reports/report-source/utils.ts";
  import type { ReportValues } from "@rilldata/web-common/features/scheduled-reports/utils.ts";
  import { CheckIcon } from "lucide-svelte";
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
  <Dropdown.Trigger>
    {#if hasSelection}
      {exploreName || metricsViewName}
    {:else}
      Select a source...
    {/if}
  </Dropdown.Trigger>
  <Dropdown.Content>
    {#each $exploreSourceOptionsStore as source, i (i)}
      {@const isSelected =
        source.metricsViewName === metricsViewName &&
        source.exploreName === exploreName &&
        source.canvasName === canvasName}
      <Dropdown.Item on:click={() => onSelect(source)}>
        <div class="w-5">
          {#if isSelected}
            <CheckIcon size="14px" />
          {/if}
        </div>
        <span>{source.label}</span>
      </Dropdown.Item>
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
