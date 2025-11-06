<script lang="ts">
  import * as Dropdown from "@rilldata/web-common/components/dropdown-menu";
  import CanvasIcon from "@rilldata/web-common/components/icons/CanvasIcon.svelte";
  import ReportSourceItem from "@rilldata/web-common/features/scheduled-reports/report-source/ReportSourceItem.svelte";
  import {
    getSecondarySourceForCanvasOptions,
    type ReportSource,
  } from "@rilldata/web-common/features/scheduled-reports/report-source/utils.ts";
  import { writable } from "svelte/store";

  export let selectedSource: ReportSource;
  export let canvasNameForSubSelector: string;
  export let onSelect: (source: ReportSource) => void;

  const canvasNameForSubSelectorStore = writable(canvasNameForSubSelector);
  $: canvasNameForSubSelectorStore.set(canvasNameForSubSelector);

  const sourceOptionsStore = getSecondarySourceForCanvasOptions(
    canvasNameForSubSelectorStore,
  );
</script>

<Dropdown.Sub>
  <Dropdown.SubTrigger class="gap-x-2">
    <CanvasIcon className="text-gray-400" size="14px" />
    {canvasNameForSubSelector}
  </Dropdown.SubTrigger>
  <Dropdown.SubContent class="w-[250px]">
    {#each $sourceOptionsStore as source, i (i)}
      <ReportSourceItem {selectedSource} {source} {onSelect} />
    {:else}
      <!-- We cannot query upfront if there are any valid metrics views for a canvas, so we show a placeholder message.
           Since we filter by valid canvas specs so there will be a valid metrics view most of the times -->,
      <div>No valid metrics views for this canvas</div>
    {/each}
  </Dropdown.SubContent>
</Dropdown.Sub>
