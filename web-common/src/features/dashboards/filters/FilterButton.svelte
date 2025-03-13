<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import Add from "@rilldata/web-common/components/icons/Add.svelte";
  import type { SearchableFilterSelectableGroup } from "@rilldata/web-common/components/searchable-filter-menu/SearchableFilterSelectableItem";
  import SearchableMenuContent from "@rilldata/web-common/components/searchable-filter-menu/SearchableMenuContent.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { getDimensionDisplayName } from "@rilldata/web-common/features/dashboards/filters/getDisplayName";
  import type {
    MetricsViewSpecDimensionV2,
    MetricsViewSpecMeasureV2,
  } from "@rilldata/web-common/runtime-client";
  import { getMeasureDisplayName } from "./getDisplayName";

  export let allDimensions: MetricsViewSpecDimensionV2[];
  export let filteredSimpleMeasures: MetricsViewSpecMeasureV2[];
  export let dimensionHasFilter: (dimensionName: string) => boolean;
  export let measureHasFilter: (measureName: string) => boolean;
  export let setTemporaryFilterName: (name: string) => void;
  export let addBorder = true;

  let open = false;

  $: selectableGroups = [
    <SearchableFilterSelectableGroup>{
      name: "DIMENSIONS",
      items:
        allDimensions
          ?.map((d) => ({
            name: (d.name || d.column) as string,
            label: getDimensionDisplayName(d),
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
  <DropdownMenu.Trigger asChild let:builder id="filter-add-btn">
    <Tooltip distance={8} suppress={open}>
      <button
        class:addBorder
        class:active={open}
        use:builder.action
        {...builder}
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
  />
</DropdownMenu.Root>

<style lang="postcss">
  button {
    @apply w-[34px] h-[26px] rounded-2xl;
    @apply flex items-center justify-center;

    @apply bg-white;
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
