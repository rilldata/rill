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

  export let allDimensions: MetricsViewSpecDimension[];
  export let filteredSimpleMeasures: MetricsViewSpecMeasure[];
  export let dimensionHasFilter: (dimensionName: string) => boolean;
  export let measureHasFilter: (measureName: string) => boolean;
  export let setTemporaryFilterName: (name: string) => void;
  export let addBorder = true;
  export let side: "top" | "right" | "bottom" | "left" = "bottom";

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
    @apply bg-surface-container;
  }

  button.addBorder {
    @apply border border-dashed border-gray-300;
  }

  button:hover {
    @apply bg-gray-100;
  }

  button:active,
  .active {
    @apply bg-gray-200;
  }
</style>
