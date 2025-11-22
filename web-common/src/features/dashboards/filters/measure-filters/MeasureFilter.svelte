<script lang="ts">
  import { Chip } from "@rilldata/web-common/components/chip";
  import * as Popover from "@rilldata/web-common/components/popover/";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import TooltipTitle from "@rilldata/web-common/components/tooltip/TooltipTitle.svelte";
  import type { MeasureFilterEntry } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
  import MeasureFilterBody from "@rilldata/web-common/features/dashboards/filters/measure-filters/MeasureFilterBody.svelte";
  import type { MetricsViewSpecDimension } from "@rilldata/web-common/runtime-client";
  import { fly } from "svelte/transition";
  import MeasureFilterForm from "./MeasureFilterForm.svelte";

  export let dimensionName: string;
  export let name: string;
  export let label: string | undefined = undefined;
  export let filter: MeasureFilterEntry | undefined = undefined;
  export let onRemove: () => void;
  export let onApply: (params: {
    dimension: string;
    oldDimension: string;
    filter: MeasureFilterEntry;
  }) => void;
  export let allDimensions: MetricsViewSpecDimension[];
  export let side: "top" | "right" | "bottom" | "left" = "bottom";

  let open = !filter;
</script>

<Popover.Root
  bind:open
  onOpenChange={(open) => {
    if (!open && !filter) {
      onRemove();
    }
  }}
  preventScroll
>
  <Popover.Trigger asChild let:builder>
    <Tooltip
      activeDelay={60}
      alignment="start"
      distance={8}
      location="bottom"
      suppress={open}
    >
      <Chip
        type="measure"
        active={open}
        builders={[builder]}
        {label}
        theme
        {onRemove}
        removable
        removeTooltipText="Remove {label}"
      >
        <MeasureFilterBody
          dimensionName={allDimensions.find((d) => {
            return d.name === dimensionName;
          })?.displayName ?? ""}
          {filter}
          {label}
          slot="body"
        />
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
  </Popover.Trigger>

  {#if open}
    <MeasureFilterForm
      bind:open
      {name}
      {filter}
      {dimensionName}
      {allDimensions}
      {onApply}
      {side}
    />
  {/if}
</Popover.Root>
