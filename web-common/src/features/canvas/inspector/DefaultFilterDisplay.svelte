<script lang="ts">
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { getCanvasStore } from "../state-managers/state-managers";
  import DimensionFilterReadOnlyChip from "../../dashboards/filters/dimension-filters/DimensionFilterReadOnlyChip.svelte";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Trash from "@rilldata/web-common/components/icons/Trash.svelte";

  export let canvasName: string;

  $: ({ instanceId } = $runtime);

  $: ({
    canvasEntity: {
      clearDefaultFilters,
      filterManager: { _defaultUIFilters },
      timeControls: { interval: _interval },
    },
  } = getCanvasStore(canvasName, instanceId));

  $: ({ dimensions } = $_defaultUIFilters);

  $: interval = $_interval;
</script>

<div class="flex-col flex h-full">
  <div class="page-param">
    <p class="text-muted-foreground mb-4">
      The filters listed below are saved as your default view and will
      automatically apply each time you open this dashboard in Rill Cloud.
    </p>

    <div class="flex flex-col gap-y-2 w-full flex-none">
      {#each dimensions as [name, entry] (name)}
        {@const metricsViewNames = Array.from(entry.dimensions.keys())}
        {@const dimension = entry.dimensions.get(metricsViewNames[0])}

        {#if dimension && dimension.name}
          <DimensionFilterReadOnlyChip
            name={dimension.name}
            {metricsViewNames}
            label={dimension.displayName ||
              dimension.name ||
              dimension.column ||
              "Unnamed Dimension"}
            mode={entry.mode}
            values={entry.selectedValues}
            inputText={entry.inputText}
            isInclude={entry.isInclude}
            timeStart={interval?.start?.toISO()}
            timeEnd={interval?.end?.toISO()}
          />
        {/if}
      {/each}
    </div>
  </div>

  <div class="mt-auto border-t w-full px-5 py-3">
    <Button type="secondary" wide onClick={clearDefaultFilters}>
      <Trash />
      Clear default filters
    </Button>
  </div>
</div>

<style lang="postcss">
  .page-param {
    @apply py-3 px-5;
    @apply border-t;
  }
</style>
