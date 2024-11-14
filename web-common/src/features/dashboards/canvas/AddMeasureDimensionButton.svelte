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
  } from "@rilldata/web-common/features/dashboards/filters/getDisplayName";
  import { createEventDispatcher } from "svelte";
  import { getStateManagers } from "../state-managers/state-managers";
  const {
    selectors: {
      dimensions: { allDimensions },
      measures: { filteredSimpleMeasures },
    },
  } = getStateManagers();

  const dispatch = createEventDispatcher();

  let open = false;

  $: selectableGroups = [
    <SearchableFilterSelectableGroup>{
      name: "MEASURES",
      items:
        $filteredSimpleMeasures()?.map((m) => ({
          name: m.name as string,
          label: getMeasureDisplayName(m),
        })) ?? [],
    },
    <SearchableFilterSelectableGroup>{
      name: "DIMENSIONS",
      items:
        $allDimensions?.map((d) => ({
          name: (d.name || d.column) as string,
          label: getDimensionDisplayName(d),
        })) ?? [],
    },
  ];
</script>

<DropdownMenu.Root bind:open typeahead={false}>
  <DropdownMenu.Trigger asChild let:builder>
    <Tooltip distance={8} suppress={open}>
      <button class:active={open} use:builder.action {...builder}>
        <Add size="17px" />
      </button>
      <TooltipContent slot="tooltip-content">Add field</TooltipContent>
    </Tooltip>
  </DropdownMenu.Trigger>

  <SearchableMenuContent
    allowMultiSelect={false}
    onSelect={(name) => {
      dispatch("addField", name);
    }}
    {selectableGroups}
    selectedItems={[]}
  />
</DropdownMenu.Root>

<style lang="postcss">
  button {
    @apply w-[24px] h-[24px] rounded-xl;
    @apply flex items-center justify-center;
    @apply border border-dashed border-slate-300;
    @apply bg-white;
  }

  button:hover {
    @apply bg-slate-100;
  }

  button:active,
  .active {
    @apply bg-slate-200;
  }
</style>
