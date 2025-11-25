<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import Add from "@rilldata/web-common/components/icons/Add.svelte";
  import type { SearchableFilterSelectableGroup } from "@rilldata/web-common/components/searchable-filter-menu/SearchableFilterSelectableItem";
  import SearchableMenuContent from "@rilldata/web-common/components/searchable-filter-menu/SearchableMenuContent.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { getDimensionDisplayName } from "@rilldata/web-common/features/dashboards/filters/getDisplayName";
  import type {
    MetricsViewSpecDimension,
    MetricsViewSpecMeasure,
  } from "@rilldata/web-common/runtime-client";
  import { getMeasureDisplayName } from "./getDisplayName";
  import type { DimensionLookup } from "../../canvas/stores/filters";

  export let allDimensions: DimensionLookup;
  export let filteredSimpleMeasures: MetricsViewSpecMeasure[];
  export let addBorder = true;
  export let side: "top" | "right" | "bottom" | "left" = "bottom";
  export let dimensionHasFilter: (dimensionName: string) => boolean;
  export let measureHasFilter: (measureName: string) => boolean;
  export let setTemporaryFilterName: (name: string) => void;

  let open = false;

  $: consolidated = Array.from(allDimensions.entries()).map(
    ([id, mvDimMap]) => {
      return { ...Array.from(mvDimMap.values())[0], id: id };
    },
  );

  $: selectableGroups = [
    <SearchableFilterSelectableGroup>{
      name: "DIMENSIONS",
      items:
        consolidated
          ?.map((d) => ({
            name: d.id,
            // label: getDimensionDisplayName(d),
            label: d.id,
          }))
          .filter((d) => !dimensionHasFilter(d.name)) ?? [],
    },
    <SearchableFilterSelectableGroup>{
      name: "MEASURES",
      items:
        filteredSimpleMeasures
          ?.map((m) => ({
            name: m.name as string,
            label: getMeasureDisplayName(m),
          }))
          .filter((m) => !measureHasFilter(m.name)) ?? [],
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
