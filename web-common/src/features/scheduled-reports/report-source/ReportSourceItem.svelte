<script lang="ts">
  import * as Dropdown from "@rilldata/web-common/components/dropdown-menu";
  import { resourceIconMapping } from "@rilldata/web-common/features/entity-management/resource-icon-mapping.ts";
  import type { ReportSource } from "@rilldata/web-common/features/scheduled-reports/report-source/utils.ts";
  import { CheckIcon } from "lucide-svelte";

  export let selectedSource: ReportSource;
  export let source: ReportSource;
  export let onSelect: (source: ReportSource) => void;

  const isSelected =
    source.metricsViewName === selectedSource.metricsViewName &&
    source.exploreName === selectedSource.exploreName &&
    source.canvasName === selectedSource.canvasName;

  $: iconComponent = resourceIconMapping[source.kind];
</script>

<Dropdown.Item
  on:click={() => onSelect(source)}
  class="justify-between gap-x-2 w-full"
>
  <span
    class="flex flex-row items-center gap-x-2 text-ellipsis overflow-hidden whitespace-nowrap"
  >
    <svelte:component
      this={iconComponent}
      className="text-gray-400"
      size="14px"
    />
    {source.label}
  </span>
  <div class="w-5">
    {#if isSelected}
      <CheckIcon size="14px" />
    {/if}
  </div>
</Dropdown.Item>
