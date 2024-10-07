<script lang="ts">
  import { Chip } from "@rilldata/web-common/components/chip";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import TooltipTitle from "@rilldata/web-common/components/tooltip/TooltipTitle.svelte";
  import type { MeasureFilterEntry } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
  import MeasureFilterBody from "@rilldata/web-common/features/dashboards/filters/measure-filters/MeasureFilterBody.svelte";
  import MeasureFilterMenu from "@rilldata/web-common/features/dashboards/filters/measure-filters/MeasureFilterMenu.svelte";
  import { createEventDispatcher } from "svelte";
  import { fly } from "svelte/transition";

  export let dimensionName: string;
  export let name: string;
  export let label: string | undefined = undefined;
  export let filter: MeasureFilterEntry | undefined = undefined;

  const dispatch = createEventDispatcher();

  let active = !filter;

  function handleDismiss() {
    if (!filter) {
      dispatch("remove");
    } else {
      active = false;
    }
  }
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
        type="measure"
        {active}
        builders={[builder]}
        {label}
        on:remove={() => dispatch("remove")}
        removable
        removeTooltipText="Remove {label}"
      >
        <MeasureFilterBody {dimensionName} {filter} {label} slot="body" />
      </Chip>
      <div slot="tooltip-content" transition:fly={{ duration: 100, y: 4 }}>
        <TooltipContent maxWidth="400px">
          <TooltipTitle>
            <svelte:fragment slot="name">{name}</svelte:fragment>
            <svelte:fragment slot="description">{label || ""}</svelte:fragment>
          </TooltipTitle>

          <slot name="body-tooltip-content">Click to edit the values</slot>
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
        filter,
      })}
    open={active}
  />
</DropdownMenu.Root>
