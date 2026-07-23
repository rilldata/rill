<script lang="ts">
  import { fly } from "svelte/transition";
  import { Chip } from "@rilldata/web-common/components/chip";
  import * as Popover from "@rilldata/web-common/components/popover/";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import TooltipTitle from "@rilldata/web-common/components/tooltip/TooltipTitle.svelte";
  import type { MeasureFilterEntry } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
  import MeasureFilterBody from "@rilldata/web-common/features/dashboards/filters/measure-filters/MeasureFilterBody.svelte";
  import type { MetricsViewSpecDimension } from "@rilldata/web-common/runtime-client";
  import MeasureFilterForm from "@rilldata/web-common/features/dashboards/filters/measure-filters/MeasureFilterForm.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import type { MeasureFilterManager } from "@rilldata/web-common/features/dashboards/filters/manager/measure-filter-manager.svelte.ts";

  let {
    measureManager,
    openOnMount = true,
    allDimensions,
    side = "bottom",
  }: {
    measureManager: MeasureFilterManager;
    openOnMount?: boolean;
    allDimensions: MetricsViewSpecDimension[];
    side?: "top" | "right" | "bottom" | "left";
  } = $props();

  // svelte-ignore state_referenced_locally
  let open = $state(openOnMount && !measureManager.expr);
  // svelte-ignore state_referenced_locally
  let curPinned = $state(measureManager.pinned);
  // svelte-ignore state_referenced_locally
  let curRequired = $state(measureManager.required);

  let missingRequired = $derived(
    measureManager.required && !measureManager.expr,
  );

  let filter = $derived(
    measureManager.expr
      ? <MeasureFilterEntry>{
          measure: measureManager.name,
          operation: measureManager.operation,
          type: measureManager.type,
          value1: measureManager.value1,
          value2: measureManager.value2,
        }
      : undefined,
  );

  function onApply(dimension: string, filter: MeasureFilterEntry) {
    if (measureManager.pinned !== curPinned) {
      measureManager.togglePinned();
    }
    if (measureManager.required !== curRequired) {
      measureManager.toggleRequired();
    }
    measureManager.apply(dimension, filter);
  }
</script>

<Popover.Root
  bind:open
  onOpenChange={() => {
    if (!open) return;
    if (measureManager.pinned !== curPinned) {
      measureManager.togglePinned();
    }
    if (measureManager.required !== curRequired) {
      measureManager.toggleRequired();
    }
  }}
>
  <Popover.Trigger>
    {#snippet child({ props })}
      <Tooltip
        activeDelay={60}
        alignment="start"
        distance={8}
        location="bottom"
        suppress={open}
      >
        <Chip
          {...props}
          type="measure"
          active={open}
          label={measureManager.label}
          gray={!measureManager.expr}
          error={!!missingRequired}
          theme
          onRemove={() => measureManager.clear()}
          removable={!curPinned && !measureManager.required}
          removeTooltipText={m.dashboard_remove_label({
            label: measureManager.label,
          })}
        >
          <MeasureFilterBody
            dimensionName={allDimensions.find((d) => {
              return d.name === measureManager.dimension;
            })?.displayName ?? ""}
            {filter}
            label={measureManager.label}
            slot="body"
          />
        </Chip>
        <div slot="tooltip-content" transition:fly={{ duration: 100, y: 4 }}>
          <TooltipContent maxWidth="400px">
            <TooltipTitle>
              <svelte:fragment slot="name"
                >{measureManager.name}</svelte:fragment
              >
              <svelte:fragment slot="description"
                >{measureManager.required
                  ? m.dashboard_required_measure()
                  : measureManager.label || ""}</svelte:fragment
              >
            </TooltipTitle>

            {#if missingRequired}
              {m.dashboard_filter_required_set_value()}
            {:else}
              <slot name="body-tooltip-content">
                {m.dashboard_click_to_edit_values()}
              </slot>
            {/if}
          </TooltipContent>
        </div>
      </Tooltip>
    {/snippet}
  </Popover.Trigger>

  {#if open}
    <MeasureFilterForm
      bind:open
      name={measureManager.name}
      {filter}
      label={measureManager.label}
      dimensionName={measureManager.dimension}
      {allDimensions}
      onApply={({ dimension, filter }) => onApply(dimension, filter)}
      bind:pinned={curPinned}
      bind:required={curRequired}
      showPinControl={measureManager.editing}
      showRequiredControl={measureManager.editing}
      {side}
    />
  {/if}
</Popover.Root>
