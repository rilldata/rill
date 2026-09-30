<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import { Chip } from "@rilldata/web-common/components/chip";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import ArrowDown from "@rilldata/web-common/components/icons/ArrowDown.svelte";
  import ChevronRight from "@rilldata/web-common/components/icons/ChevronRight.svelte";
  import {
    decodePivotSort,
    pivotSortTargetsEqual,
  } from "@rilldata/web-common/features/dashboards/pivot/pivot-sort";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import type { PivotCanvasComponent } from "../components/pivot";
  import { getSortChips } from "../components/pivot/default-sort";

  export let component: PivotCanvasComponent;
  export let label: string;

  $: ({ pivotState, config, pivotDataStore, specStore } = component);

  // The user's current interactive sort on the table (raw TanStack sort id).
  $: activeSort = $pivotState.sorting[0];

  $: storedDefault =
    "default_sort" in $specStore ? $specStore.default_sort : undefined;

  $: columnDimensionAxes = $pivotDataStore?.columnDimensionAxes ?? {};

  $: activeTarget = activeSort
    ? decodePivotSort([activeSort], $config, columnDimensionAxes)
    : undefined;

  // Show the active sort while the user is sorting; otherwise the stored default.
  $: shownSort = activeTarget ?? storedDefault;

  $: chips = shownSort ? getSortChips(shownSort, $config) : [];

  $: canMakeDefault =
    !!activeTarget && !pivotSortTargetsEqual(storedDefault, activeTarget);

  function makeDefault() {
    if (!activeTarget) return;
    component.updateProperty("default_sort", activeTarget);
  }

  function clearDefault() {
    component.updateProperty("default_sort", undefined);
  }
</script>

<div class="flex flex-col gap-y-2 py-1">
  <InputLabel small {label} id="default_sort" />

  {#if shownSort}
    <div class="flex items-center gap-1 flex-wrap">
      {#each chips as chip, i (i)}
        {#if i > 0}
          <span class="text-fg-disabled shrink-0">
            <ChevronRight size="12px" />
          </span>
        {/if}
        <Chip readOnly compact type={chip.type} label={chip.label}>
          <span class="font-bold truncate" slot="body">{chip.label}</span>
        </Chip>
      {/each}
      <span
        class="ml-0.5 text-fg-secondary transition-transform"
        class:-rotate-180={!shownSort.desc}
        title={shownSort.desc
          ? m.canvas_default_sort_descending()
          : m.canvas_default_sort_ascending()}
      >
        <ArrowDown size="14px" />
      </span>
    </div>

    <div class="flex items-center justify-between">
      <Button type="text" disabled={!canMakeDefault} onclick={makeDefault}>
        {m.canvas_default_sort_set()}
      </Button>
      {#if storedDefault}
        <Button type="text" onclick={clearDefault}
          >{m.canvas_default_sort_clear()}</Button
        >
      {/if}
    </div>
  {:else}
    <span class="text-xs text-fg-disabled">
      {m.canvas_default_sort_empty()}
    </span>
  {/if}
</div>
