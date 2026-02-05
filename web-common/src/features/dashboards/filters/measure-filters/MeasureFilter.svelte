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
  import type { FilterManager } from "@rilldata/web-common/features/canvas/stores/filter-manager";
  import type { MeasureFilterItem } from "../../state-managers/selectors/measure-filters";

  export let filterData: MeasureFilterItem;
  export let openOnMount = false;
  export let allDimensions: MetricsViewSpecDimension[];
  export let side: "top" | "right" | "bottom" | "left" = "bottom";
  export let toggleFilterPin:
    | FilterManager["actions"]["toggleFilterPin"]
    | undefined = undefined;
  export let onRemove: () => void;
  export let onApply: (params: {
    dimension: string;
    oldDimension: string;
    filter: MeasureFilterEntry;
  }) => void;

  let open = openOnMount && !filterData.filter;
  let curPinned = filterData.pinned;

  $: ({ filter, pinned, label, measures, dimensionName, name } = filterData);

  $: metricsViewNames = measures ? Array.from(measures.keys()) : [];
</script>

<Popover.Root
  bind:open
  preventScroll
  onOpenChange={() => {
    if (open && pinned !== curPinned) {
      toggleFilterPin?.(name, metricsViewNames);
    }
  }}
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
        gray={!filter}
        theme
        {onRemove}
        removable={!curPinned}
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
      {label}
      {dimensionName}
      {allDimensions}
      onApply={(params) => {
        if (pinned !== curPinned) {
          toggleFilterPin?.(name, metricsViewNames);
        }
        onApply(params);
      }}
      bind:pinned={curPinned}
      showPinControl={!!toggleFilterPin}
      {side}
    />
  {/if}
</Popover.Root>
