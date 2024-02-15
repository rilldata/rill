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
  export let readOnly: boolean = false;

  const dispatch = createEventDispatcher();

  let active = !expr;

  function handleDismiss() {
    if (!expr) {
      dispatch("remove");
    } else {
      active = false;
    }
  }
</script>

{#if readOnly}
  <Chip {...colors} {active} extraRounded={false} {label} outline readOnly>
    <MeasureFilterBody {dimensionName} {expr} {label} readOnly slot="body" />
  </Chip>
{:else}
  <DropdownMenu.Root
    preventScroll
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
          extraRounded={false}
          {label}
          builders={[builder]}
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
          <MeasureFilterBody {dimensionName} {expr} {label} slot="body" />
        </Chip>
        <div slot="tooltip-content" transition:fly={{ duration: 100, y: 4 }}>
          <TooltipContent maxWidth="400px">
            <TooltipTitle>
              <svelte:fragment slot="name">{name}</svelte:fragment>
              <svelte:fragment slot="description">{label || ""}</svelte:fragment
              >
            </TooltipTitle>
            {#if $$slots["body-tooltip-content"]}
              <slot name="body-tooltip-content">Click to edit the values</slot>
            {/if}
          </TooltipContent>
        </div>
      </Tooltip>
    </DropdownMenu.Trigger>

    <MeasureFilterMenu {dimensionName} {expr} {name} on:apply open={active} />
  </DropdownMenu.Root>
{/if}
