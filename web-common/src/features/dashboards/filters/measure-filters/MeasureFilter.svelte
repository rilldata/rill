<script lang="ts">
  import { Chip } from "@rilldata/web-common/components/chip";
  import {
    ChipColors,
    measureChipColors,
  } from "@rilldata/web-common/components/chip/chip-types";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import TooltipTitle from "@rilldata/web-common/components/tooltip/TooltipTitle.svelte";
  import {
    mapExprToMeasureFilter,
    mapMeasureFilterToExpr,
    MeasureFilterEntry,
  } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
  import MeasureFilterBody from "@rilldata/web-common/features/dashboards/filters/measure-filters/MeasureFilterBody.svelte";
  import MeasureFilterMenu from "@rilldata/web-common/features/dashboards/filters/measure-filters/MeasureFilterMenu.svelte";
  import type { V1Expression } from "@rilldata/web-common/runtime-client";
  import { createEventDispatcher } from "svelte";
  import { fly } from "svelte/transition";

  export let dimensionName: string;
  export let name: string;
  export let label: string | undefined = undefined;
  export let colors: ChipColors = measureChipColors;
  export let expr: V1Expression | undefined = undefined;

  const dispatch = createEventDispatcher();

  let active = !expr;

  function handleDismiss() {
    if (!expr) {
      dispatch("remove");
    } else {
      active = false;
    }
  }

  // TODO: in the next round of refactor update upstream to use MeasureFilterEntry
  let filter: MeasureFilterEntry | undefined;
  $: filter = mapExprToMeasureFilter(expr);
</script>

<DropdownMenu.Root
  bind:open={active}
  onOpenChange={(open) => {
    if (!open) {
      // Clicking outside a menu triggers a transition
      // Wait for that transition to finish before dismissing the pill
      setTimeout(() => {
        handleDismiss();
      }, 60);
    }
  }}
  preventScroll
>
  <DropdownMenu.Trigger asChild let:builder>
    <Tooltip
      activeDelay={60}
      alignment="start"
      distance={8}
      location="bottom"
      suppress={active}
    >
      <Chip
        {...colors}
        {active}
        builders={[builder]}
        extraRounded={false}
        {label}
        on:remove={() => dispatch("remove")}
        outline
        removable
      >
        <!-- remove button tooltip -->
        <svelte:fragment slot="remove-tooltip">
          <slot name="remove-tooltip-content">
            Remove {label}
          </slot>
        </svelte:fragment>
        <!-- body -->
        <MeasureFilterBody {dimensionName} {filter} {label} slot="body" />
      </Chip>
      <div slot="tooltip-content" transition:fly={{ duration: 100, y: 4 }}>
        <TooltipContent maxWidth="400px">
          <TooltipTitle>
            <svelte:fragment slot="name">{name}</svelte:fragment>
            <svelte:fragment slot="description">{label || ""}</svelte:fragment>
          </TooltipTitle>
          {#if $$slots["body-tooltip-content"]}
            <slot name="body-tooltip-content">Click to edit the values</slot>
          {/if}
        </TooltipContent>
      </div>
    </Tooltip>
  </DropdownMenu.Trigger>

  <MeasureFilterMenu
    {dimensionName}
    {filter}
    {name}
    on:apply={({ detail: { dimension, oldDimension, filter } }) =>
      dispatch("apply", {
        dimension,
        oldDimension,
        expr: mapMeasureFilterToExpr(filter),
      })}
    open={active}
  />
</DropdownMenu.Root>
