<script lang="ts">
  import * as Dropdown from "@rilldata/web-common/components/dropdown-menu";
  import type { ReportSource } from "@rilldata/web-common/features/scheduled-reports/report-source/utils.ts";
  import { CheckIcon } from "lucide-svelte";

  export let metricsViewName;
  export let exploreName;
  export let canvasName;
  export let source: ReportSource;
  export let onSelect: (source: ReportSource) => void;

  const isSelected =
    source.metricsViewName === metricsViewName &&
    source.exploreName === exploreName &&
    source.canvasName === canvasName;
</script>

<Dropdown.Item
  on:click={() => onSelect(source)}
  class="justify-between gap-x-2 w-full"
>
  <span class="text-ellipsis overflow-hidden whitespace-nowrap">
    {source.label}
  </span>
  <div class="w-5">
    {#if isSelected}
      <CheckIcon size="14px" />
    {/if}
  </div>
</Dropdown.Item>
