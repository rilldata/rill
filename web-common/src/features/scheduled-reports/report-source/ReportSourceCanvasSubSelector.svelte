<script lang="ts">
  import * as Dropdown from "@rilldata/web-common/components/dropdown-menu";
  import ReportSourceItem from "@rilldata/web-common/features/scheduled-reports/report-source/ReportSourceItem.svelte";
  import {
    getSecondarySourceForCanvasOptions,
    type ReportSource,
  } from "@rilldata/web-common/features/scheduled-reports/report-source/utils.ts";
  import type { ReportValues } from "@rilldata/web-common/features/scheduled-reports/utils.ts";
  import { type Readable, writable } from "svelte/store";

  export let data: Readable<ReportValues>;
  export let canvasNameForSubSelector: string;
  export let onSelect: (source: ReportSource) => void;

  const canvasNameForSubSelectorStore = writable(canvasNameForSubSelector);
  $: canvasNameForSubSelectorStore.set(canvasNameForSubSelector);

  $: ({ metricsViewName, exploreName, canvasName } = $data);

  const sourceOptionsStore = getSecondarySourceForCanvasOptions(
    canvasNameForSubSelectorStore,
  );
</script>

<Dropdown.Sub>
  <Dropdown.SubTrigger>
    {canvasNameForSubSelector}
  </Dropdown.SubTrigger>
  <Dropdown.SubContent class="w-[250px]">
    {#each $sourceOptionsStore as source, i (i)}
      <ReportSourceItem
        {metricsViewName}
        {exploreName}
        {canvasName}
        {source}
        {onSelect}
      />
    {/each}
  </Dropdown.SubContent>
</Dropdown.Sub>
