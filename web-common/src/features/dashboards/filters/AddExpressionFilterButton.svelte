<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import type { ExpressionFilterManager } from "@rilldata/web-common/features/dashboards/filters/ExpressionFilterManager.svelte.ts";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import Add from "@rilldata/web-common/components/icons/Add.svelte";
  import SearchableMenuContent from "@rilldata/web-common/components/searchable-filter-menu/SearchableMenuContent.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import {
    getDimensionDisplayName,
    getMeasureDisplayName,
  } from "@rilldata/web-common/features/dashboards/filters/getDisplayName.ts";
  import type { SearchableFilterSelectableGroup } from "@rilldata/web-common/components/searchable-filter-menu/SearchableFilterSelectableItem.ts";
  import { isSimpleMeasure } from "@rilldata/web-common/features/dashboards/state-managers/selectors/measures.ts";

  let {
    expressionFilterManager,
    excludedDimensions = {},
    excludedMeasures = {},
    addBorder = true,
    side = "bottom",
  }: {
    expressionFilterManager: ExpressionFilterManager;
    excludedDimensions?: Record<string, boolean>;
    excludedMeasures?: Record<string, boolean>;
    addBorder?: boolean;
    side?: "top" | "right" | "bottom" | "left";
  } = $props();

  let open = $state(false);

  let { metricsViewsProvider, filterManagersMap } = $derived(
    expressionFilterManager,
  );
  let { dimensions, measures } = $derived(metricsViewsProvider);

  let dimensionEntries = $derived(
    dimensions
      .filter(
        (d) =>
          !(d.name! in filterManagersMap) && !(d.name! in excludedDimensions),
      )
      .map((d) => ({
        label: getDimensionDisplayName(d),
        name: d.name!,
      })),
  );

  let measureEntries = $derived(
    measures
      .filter(
        (m) =>
          !(m.name! in filterManagersMap) &&
          !(m.name! in excludedMeasures) &&
          isSimpleMeasure(m),
      )
      .map((m) => ({
        label: getMeasureDisplayName(m),
        name: m.name!,
      })),
  );

  let selectableGroups = $derived([
    <SearchableFilterSelectableGroup>{
      name: m.filter_dimensions(),
      items: dimensionEntries,
    },
    <SearchableFilterSelectableGroup>{
      name: m.filter_measures(),
      items: measureEntries,
    },
  ]);
</script>

<DropdownMenu.Root bind:open>
  <DropdownMenu.Trigger>
    {#snippet child({ props })}
      <Tooltip distance={8} suppress={open}>
        <button
          {...props}
          class:addBorder
          class:active={open}
          aria-label={m.dashboard_add_filter_button()}
        >
          <Add size="17px" />
        </button>
        <TooltipContent slot="tooltip-content"
          >{m.dashboard_add_filter()}</TooltipContent
        >
      </Tooltip>
    {/snippet}
  </DropdownMenu.Trigger>

  <SearchableMenuContent
    allowMultiSelect={false}
    onSelect={(name) => expressionFilterManager.addNewFilter(name)}
    {selectableGroups}
    selectedItems={[]}
    {side}
  />
</DropdownMenu.Root>

<style lang="postcss">
  button {
    @apply w-[34px] h-[26px] rounded-2xl;
    @apply flex items-center justify-center;
    @apply bg-surface-subtle;
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
