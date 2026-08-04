<script lang="ts">
  import { fly } from "svelte/transition";
  import { Chip } from "web-common/src/components/chip";
  import * as Popover from "web-common/src/components/popover";
  import Tooltip from "web-common/src/components/tooltip/Tooltip.svelte";
  import TooltipContent from "web-common/src/components/tooltip/TooltipContent.svelte";
  import TooltipTitle from "web-common/src/components/tooltip/TooltipTitle.svelte";
  import type { MeasureFilterEntry } from "web-common/src/features/dashboards/filters/measure-filters/measure-filter-entry";
  import MeasureFilterBody from "web-common/src/features/dashboards/filters/measure-filters/MeasureFilterBody.svelte";
  import type { MetricsViewSpecDimension } from "web-common/src/runtime-client";
  import MeasureFilterForm from "web-common/src/features/dashboards/filters/measure-filters/MeasureFilterForm.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import type { MeasureFilterManager } from "./MeasureFilterManager.svelte.ts";
  import type { YAMLConfigProvider } from "web-common/src/features/dashboards/providers/YAMLConfigProvider.svelte.ts";

  let {
    measureManager,
    yamlConfigProvider,
    openOnMount = true,
    allDimensions,
    removable = true,
    side = "bottom",
  }: {
    measureManager: MeasureFilterManager;
    yamlConfigProvider: YAMLConfigProvider;
    openOnMount?: boolean;
    allDimensions: MetricsViewSpecDimension[];
    removable?: boolean;
    side?: "top" | "right" | "bottom" | "left";
  } = $props();

  // svelte-ignore state_referenced_locally
  let open = $state(openOnMount && !measureManager.expr);

  let pinned = $derived(
    Boolean(yamlConfigProvider.pinnedFilters[measureManager.name]),
  );
  let required = $derived(
    Boolean(yamlConfigProvider.requiredFilters[measureManager.name]),
  );
  let missingRequired = $derived(required && !measureManager.expr);

  // Pinned and required are edited locally and only persisted once the popover closes.
  // svelte-ignore state_referenced_locally
  let curPinned = $state(pinned);
  // svelte-ignore state_referenced_locally
  let curRequired = $state(required);

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
    persistPinnedAndRequired();
    measureManager.setMeasureFilter(dimension, filter);
  }

  function persistPinnedAndRequired() {
    if (pinned !== curPinned) {
      yamlConfigProvider.togglePinnedFilter(measureManager.name);
    }
    if (required !== curRequired) {
      yamlConfigProvider.toggleRequiredFilter(measureManager.name);
    }
  }
</script>

<Popover.Root
  bind:open
  onOpenChange={(open) => {
    if (open) {
      curPinned = pinned;
      curRequired = required;
    } else {
      persistPinnedAndRequired();
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
          removable={removable && !pinned && !required}
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
                >{required
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
      showPinControl={yamlConfigProvider.editable}
      showRequiredControl={yamlConfigProvider.editable}
      {side}
    />
  {/if}
</Popover.Root>
