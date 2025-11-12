<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import {
    getMeasureDisplayName,
    getDimensionDisplayName,
  } from "@rilldata/web-common/features/dashboards/filters/getDisplayName.ts";
  import { getValidDashboardsQueryOptions } from "@rilldata/web-common/features/dashboards/selectors.ts";
  import { getExploreValidSpecQueryOptions } from "@rilldata/web-common/features/explores/selectors.ts";
  import { createQuery } from "@tanstack/svelte-query";
  import { writable } from "svelte/store";

  export let exploreName: string;
  export let onAdd: (value: string, label: string) => void;

  const exploreNameStore = writable("");
  $: exploreNameStore.set(exploreName);

  const exploresSpecQuery = createQuery(getValidDashboardsQueryOptions());
  $: exploreOptions =
    $exploresSpecQuery.data?.map((r) => {
      const exploreName = r.meta?.name?.name ?? "";
      const exploreSpec = r.explore?.state?.validSpec ?? {};
      return {
        value: exploreName,
        label: exploreSpec.displayName || exploreName,
      };
    }) ?? [];

  const validSpecQuery = createQuery(
    getExploreValidSpecQueryOptions(exploreNameStore),
  );
  $: metricsViewSpec = $validSpecQuery.data?.metricsViewSpec ?? {};
  $: exploreSpec = $validSpecQuery.data?.exploreSpec ?? {};
  $: measures =
    metricsViewSpec.measures
      ?.filter((m) => exploreSpec.measures?.includes(m.name!))
      .map((m) => ({
        value: m.name ?? "",
        label: getMeasureDisplayName(m),
      })) ?? [];
  $: dimensions =
    metricsViewSpec.dimensions
      ?.filter((d) => exploreSpec.dimensions?.includes(d.name!))
      .map((d) => ({
        value: d.name ?? "",
        label: getDimensionDisplayName(d),
      })) ?? [];

  enum SelectionMode {
    Main,
    Measures,
    Dimensions,
    Explores,
  }
  $: otherOptions = [
    {
      value: SelectionMode.Measures,
      label: "Measures",
      subOptions: measures,
    },
    {
      value: SelectionMode.Dimensions,
      label: "Dimensions",
      subOptions: dimensions,
    },
    {
      value: SelectionMode.Explores,
      label: "Explores",
      subOptions: exploreOptions,
    },
  ];
  let selectionMode: SelectionMode = SelectionMode.Main;
  $: selectedOtherOption = otherOptions.find((o) => o.value === selectionMode);

  function selectMode(e, mode: SelectionMode) {
    e.preventDefault();
    e.stopPropagation();
    selectionMode = mode;
  }
</script>

<DropdownMenu.Root
  onOpenChange={(o) => {
    if (!o) selectionMode = SelectionMode.Main;
  }}
>
  <DropdownMenu.Trigger>@</DropdownMenu.Trigger>
  <DropdownMenu.Content>
    {#if selectionMode === SelectionMode.Main}
      {#each otherOptions as otherOption (otherOption.value)}
        <DropdownMenu.Item on:click={(e) => selectMode(e, otherOption.value)}>
          <span>{otherOption.label}</span>
          <span>{">"}</span>
        </DropdownMenu.Item>
      {/each}
    {:else if selectedOtherOption?.subOptions}
      <DropdownMenu.Label class="flex flex-row items-center">
        <button on:click={() => (selectionMode = SelectionMode.Main)}>
          {"<"}
        </button>
        <span>{selectedOtherOption.label}</span>
      </DropdownMenu.Label>
      {#each selectedOtherOption.subOptions as subOption (subOption.value)}
        <DropdownMenu.Item
          on:click={() =>
            onAdd(
              `${selectedOtherOption.label}:"${subOption.value}"`,
              subOption.label,
            )}
        >
          {subOption.label}
        </DropdownMenu.Item>
      {/each}
    {/if}
  </DropdownMenu.Content>
</DropdownMenu.Root>
