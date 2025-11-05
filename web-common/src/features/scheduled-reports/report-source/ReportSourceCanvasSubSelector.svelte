<script lang="ts">
  import * as Dropdown from "@rilldata/web-common/components/dropdown-menu";
  import {
    getSecondarySourceForCanvasOptions,
    type ReportSource,
  } from "@rilldata/web-common/features/scheduled-reports/report-source/utils.ts";
  import type { ReportValues } from "@rilldata/web-common/features/scheduled-reports/utils.ts";
  import { CheckIcon } from "lucide-svelte";
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
  <Dropdown.SubContent>
    {#each $sourceOptionsStore as source, i (i)}
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
  </Dropdown.SubContent>
</Dropdown.Sub>
