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

  export let allDimensions:
    | Map<string, (MetricsViewSpecDimension & { metricsViewName: string })[]>
    | MetricsViewSpecDimension[];
  export let filteredSimpleMeasures: MetricsViewSpecMeasure[];
  export let dimensionHasFilter: (dimensionName: string) => boolean;
  export let measureHasFilter: (measureName: string) => boolean;
  export let setTemporaryFilterName: (name: string) => void;
  export let addBorder = true;
  export let side: "top" | "right" | "bottom" | "left" = "bottom";

  let open = false;

  function makeDimensionGroup(
    dimensions:
      | Map<string, (MetricsViewSpecDimension & { metricsViewName: string })[]>
      | MetricsViewSpecDimension[],
  ): SearchableFilterSelectableGroup {
    if (Array.isArray(dimensions)) {
      return {
        name: "DIMENSIONS",
        items:
          dimensions
            ?.map((d) => ({
              id: (d.name || d.column) as string,
              labels: [getDimensionDisplayName(d)],
            }))
            .filter((d) => !dimensionHasFilter(d.id)) ?? [],
      };
    }

    return {
      name: "DIMENSIONS",
      items:
        Array.from(dimensions.entries())
          .map(([name, dims]) => ({
            id: name,
            labels: Array.from(
              new Set(dims.map((d) => getDimensionDisplayName(d)) || []),
            ),
            tooltip: dims.map((d) => d.metricsViewName).join(", "),
          }))
          .filter((d) => !dimensionHasFilter(d.id)) ?? [],
    };
  }

  $: selectableGroups = [
    makeDimensionGroup(allDimensions),
    <SearchableFilterSelectableGroup>{
      name: "MEASURES",
      items:
        filteredSimpleMeasures
          ?.map((m) => ({
            id: m.name as string,
            labels: Array.from(new Set([getMeasureDisplayName(m)])),
          }))
          .filter((m) => !measureHasFilter(m.id)) ?? [],
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
