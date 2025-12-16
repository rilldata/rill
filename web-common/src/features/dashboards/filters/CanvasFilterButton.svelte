<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import Add from "@rilldata/web-common/components/icons/Add.svelte";
  import type { SearchableFilterSelectableGroup } from "@rilldata/web-common/components/searchable-filter-menu/SearchableFilterSelectableItem";
  import SearchableMenuContent from "@rilldata/web-common/components/searchable-filter-menu/SearchableMenuContent.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import {
    getDimensionDisplayName,
    getMeasureDisplayName,
  } from "./getDisplayName";
  import type { MetricsViewName } from "../../canvas/stores/filter-manager";
  import type {
    MetricsViewSpecDimension,
    MetricsViewSpecMeasure,
  } from "@rilldata/web-common/runtime-client";

  export let allDimensions: Map<
    string,
    Map<MetricsViewName, MetricsViewSpecDimension>
  >;
  export let filteredSimpleMeasures: Map<
    string,
    Map<MetricsViewName, MetricsViewSpecMeasure>
  >;
  export let addBorder = true;
  export let side: "top" | "right" | "bottom" | "left" = "bottom";
  export let dimensionHasFilter: (dimensionName: string) => boolean;
  export let measureHasFilter: (measureName: string) => boolean;
  export let setTemporaryFilterName: (name: string) => void;

  let open = false;

  $: dimensionEntries = Array.from(allDimensions.entries())
    .map(([id, mvNameToDimMap]) => {
      const representativeDimension = Array.from(mvNameToDimMap.values())[0];

      const label = getDimensionDisplayName(representativeDimension);
      return { label, name: id };
    })
    .filter((entry) => !dimensionHasFilter(entry.name));

  $: measureEntries = Array.from(filteredSimpleMeasures.entries())
    .map(([id, mvNameToMeasureMap]) => {
      const representativeMeasure = Array.from(mvNameToMeasureMap.values())[0];

      const label = getMeasureDisplayName(representativeMeasure);
      return { label, name: id };
    })
    .filter((entry) => !measureHasFilter(entry.name));

  $: selectableGroups = [
    <SearchableFilterSelectableGroup>{
      name: "DIMENSIONS",
      items: dimensionEntries,
    },
    <SearchableFilterSelectableGroup>{
      name: "MEASURES",
      items: measureEntries,
    },
  ];
</script>

<DropdownMenu.Root bind:open typeahead={false}>
  <DropdownMenu.Trigger asChild let:builder>
    <Tooltip distance={8} suppress={open}>
      <button
        class:addBorder
        class:active={open}
        use:builder.action
        {...builder}
        aria-label="Add filter button"
      >
        <Add size="17px" />
      </button>
      <TooltipContent slot="tooltip-content">Add filter</TooltipContent>
    </Tooltip>
  </DropdownMenu.Trigger>

  <SearchableMenuContent
    allowMultiSelect={false}
    onSelect={(name) => {
      setTemporaryFilterName(name);
    }}
    {selectableGroups}
    selectedItems={[]}
    {side}
  />
</DropdownMenu.Root>

<style lang="postcss">
  button {
    @apply w-[34px] h-[26px] rounded-2xl;
    @apply flex items-center justify-center;
    @apply bg-surface;
  }

  button.addBorder {
    @apply border border-dashed border-slate-300;
  }

  button:hover {
    @apply bg-slate-100;
  }

  button:active,
  .active {
    @apply bg-slate-200;
  }
</style>
