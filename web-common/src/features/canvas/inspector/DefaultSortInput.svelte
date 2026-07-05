<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import { Chip } from "@rilldata/web-common/components/chip";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import ArrowDown from "@rilldata/web-common/components/icons/ArrowDown.svelte";
  import ChevronRight from "@rilldata/web-common/components/icons/ChevronRight.svelte";
  import DelayedCircleOutlineSpinner from "@rilldata/web-common/components/spinner/DelayedCircleOutlineSpinner.svelte";
  import type { PivotCanvasComponent } from "../components/pivot";
  import {
    areSortChipsReady,
    getSortChips,
    getSortLabel,
    sortEquals,
  } from "../components/pivot/default-sort";

  export let component: PivotCanvasComponent;
  export let label: string;

  $: ({ pivotState, config, pivotDataStore, specStore } = component);

  // The user's current interactive sort on the table (raw TanStack sort id).
  $: activeSort = $pivotState.sorting[0];

  $: storedDefault =
    "default_sort" in $specStore ? $specStore.default_sort : undefined;

  $: columnDimensionAxes = $pivotDataStore?.columnDimensionAxes ?? {};

  // Show the active sort while the user is sorting; otherwise the stored default.
  $: shownSort = activeSort ?? storedDefault;

  $: chipsReady = shownSort
    ? areSortChipsReady(shownSort.id, $config, columnDimensionAxes)
    : true;

  $: chips = shownSort
    ? getSortChips(shownSort.id, $config, columnDimensionAxes)
    : [];

  $: canMakeDefault = !!activeSort && !sortEquals(storedDefault, activeSort);

  function makeDefault() {
    if (!activeSort) return;
    component.updateProperty("default_sort", {
      id: activeSort.id,
      desc: activeSort.desc,
      label: getSortLabel(activeSort.id, $config, columnDimensionAxes),
    });
  }

  function clearDefault() {
    component.updateProperty("default_sort", undefined);
  }
</script>

<div class="flex flex-col gap-y-2 py-1">
  <InputLabel small {label} id="default_sort" />

  {#if shownSort}
    {#if chipsReady}
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
          title={shownSort.desc ? "Descending" : "Ascending"}
        >
          <ArrowDown size="14px" />
        </span>
      </div>
    {:else}
      <div class="flex items-center gap-1 text-fg-secondary">
        <DelayedCircleOutlineSpinner isLoading size="14px" />
      </div>
    {/if}

    <div class="flex items-center justify-between">
      <Button type="text" disabled={!canMakeDefault} onclick={makeDefault}>
        Set default
      </Button>
      {#if storedDefault}
        <Button type="text" onclick={clearDefault}>Clear</Button>
      {/if}
    </div>
  {:else}
    <span class="text-xs text-fg-disabled">
      Sort a column on the table to set it as the default.
    </span>
  {/if}
</div>
