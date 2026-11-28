<script lang="ts">
  import * as Popover from "@rilldata/web-common/components/popover";
  import type { CanvasEntity } from "@rilldata/web-common/features/canvas/stores/canvas-entity";
  import type { V1Expression } from "@rilldata/web-common/runtime-client";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { derived } from "svelte/store";
  import { SlidersHorizontal } from "lucide-svelte";
  import {
    controlFieldLabels,
    controlLabel,
    controlSummary,
  } from "./control-labels";
  import { controlKindLabel } from "./editor/labels";
  import {
    ControlKind,
    filterValueViews,
    controlMetricsViews,
    offeredTimeGrains,
    type AppliedControl,
  } from "./controls";
  import ControlInput from "./ControlInput.svelte";

  // The viewer controls declared on one level (an item, a container, a tab group or the canvas).
  let {
    controls: allControls,
    canvasEntity,
    metricsViewName = undefined,
  }: {
    controls: AppliedControl[];
    canvasEntity: CanvasEntity;
    // Set for an item's own controls: its filter controls read values from this metrics view.
    metricsViewName?: string;
  } = $props();

  let enabled = $derived(canvasEntity.controls.enabled);
  let exportMode = $derived(canvasEntity.exportMode);
  let metricsViews = $derived(canvasEntity._metricsViews);
  let components = $derived(canvasEntity.componentsStore);
  let controlViews = $derived.by(() => {
    const instances = [...$components.values()];
    return derived(
      instances.map((c) => c.controls),
      (controls) =>
        controlMetricsViews(
          instances.map((c, i) => ({
            metricsViewName: c.metricsViewName,
            controls: controls[i],
          })),
        ),
    );
  });
  // Filter controls search their values within the canvas time range, like the filter bar.
  let { timeStart, timeEnd } = $derived(canvasEntity.timeFilterManager);

  // Controls the embedding page locked apply, but viewers don't see them.
  let controls = $derived(allControls.filter((c) => !c.locked));
  let changed = $derived(canvasEntity.controls.changed);
  let changedCount = $derived(
    controls.filter((c) => $changed.some((d) => d.urlKey === c.urlKey)).length,
  );

  // Display names of the measures and dimensions of the canvas's metrics views.
  let fieldLabels = $derived(controlFieldLabels($metricsViews));

  function fieldLabel(name: string): string {
    return fieldLabels.get(name) ?? name;
  }

  function metricsViewsFor(control: AppliedControl): string[] {
    if (control.control.kind !== ControlKind.Filter) return [];
    if (metricsViewName) return [metricsViewName];
    return filterValueViews(
      control.control,
      $controlViews.get(control.urlKey) ?? [],
      (view, dimension) =>
        !!$metricsViews[view]?.state?.validSpec?.dimensions?.some(
          (d) => d.name === dimension,
        ),
    );
  }

  // The filter bar's filters on the other dimensions, per metrics view a filter control reads
  // its values from. Required filters among them keep the values query from running without them.
  function filtersFor(
    control: AppliedControl,
    views: string[],
  ): Map<string, V1Expression> | undefined {
    if (control.control.kind !== ControlKind.Filter) return undefined;
    const filters = new Map<string, V1Expression>();
    for (const view of views) {
      const filter =
        canvasEntity.expressionFilterManager.getOtherDimensionsFilter(
          control.control.fields?.[view] ?? control.control.target ?? "",
          view,
        );
      if (filter) filters.set(view, filter);
    }
    return filters;
  }

  // A grain control offers only the grains the canvas's metrics views offer, which can change
  // after the control was written.
  let availableTimeGrains = $derived(
    canvasEntity.dashboardProvider.metricsViewsProvider.availableTimeGrains,
  );
  function withOfferedGrains(control: AppliedControl): AppliedControl {
    if (control.control.kind !== ControlKind.TimeGrain) return control;
    const options = control.control.options ?? [];
    const offered = offeredTimeGrains(options, availableTimeGrains);
    if (!offered.length || offered.length === options.length) return control;
    return { ...control, control: { ...control.control, options: offered } };
  }

  // In exports there is nothing to click, so a control reads as text: "Metric: Impressions".
  function staticText(control: AppliedControl): string {
    const label =
      controlLabel(control.control, fieldLabel) ||
      controlKindLabel(control.control.kind ?? "");
    const summary = controlSummary(
      control.control,
      control.values ?? control.shownValues ?? [],
      fieldLabel,
    );
    return `${label}: ${summary}`;
  }

  // When the controls don't fit on one line they collapse into an "Adjust (n)" menu.
  // The inline row is measured while it's shown; the width it needed decides when to bring it back.
  let width = $state(0);
  let row = $state<HTMLElement | undefined>(undefined);
  let neededWidth = $state(0);
  let collapsed = $state(false);
  let keys = $derived(controls.map((c) => c.urlKey).join("|"));

  $effect(() => {
    // A different set of controls needs a new measurement.
    void keys;
    collapsed = false;
    neededWidth = 0;
  });

  $effect(() => {
    void keys;
    if (controls.length < 2 || width === 0) return;
    if (!collapsed) {
      if (row && row.scrollWidth > row.clientWidth + 1) {
        neededWidth = row.scrollWidth;
        collapsed = true;
      }
    } else if (width >= neededWidth) {
      collapsed = false;
    }
  });
</script>

{#snippet inputs()}
  {#each controls as control (control.urlKey)}
    {@const views = metricsViewsFor(control)}
    <!-- Pointing at a control of a level shows which widgets it changes (see CanvasComponent). -->
    <div
      class={controls.length > 1 ? "flex-none" : "min-w-0"}
      role="presentation"
      onmouseenter={() => {
        if (control.level !== "item") {
          canvasEntity.controls.hovered.set(control.urlKey);
        }
      }}
      onmouseleave={() => canvasEntity.controls.hovered.set(null)}
    >
      <ControlInput
        control={withOfferedGrains(control)}
        {fieldLabel}
        metricsViewNames={views}
        {timeStart}
        {timeEnd}
        filters={filtersFor(control, views)}
        onChange={(values) => canvasEntity.controls.set(control, values)}
      />
    </div>
  {/each}
{/snippet}

{#if $enabled && controls.length}
  {#if $exportMode}
    <div class="static-controls">
      {#each controls as control (control.urlKey)}
        <span>{staticText(control)}</span>
      {/each}
    </div>
  {:else}
    <div
      class="w-full min-w-0"
      role="presentation"
      bind:clientWidth={width}
      onmousedown={(e) => e.stopPropagation()}
    >
      {#if collapsed}
        <Popover.Root>
          <Popover.Trigger>
            {#snippet child({ props })}
              <button
                {...props}
                type="button"
                class="adjust"
                class:changed={changedCount > 0}
              >
                <SlidersHorizontal size="12px" />
                {m.canvas_controls_adjust({ count: controls.length })}
              </button>
            {/snippet}
          </Popover.Trigger>
          <Popover.Content align="start" class="w-fit max-w-sm p-3">
            <div class="flex flex-col items-start gap-y-2">
              {@render inputs()}
            </div>
          </Popover.Content>
        </Popover.Root>
      {:else}
        <div class="control-bar" bind:this={row}>
          {@render inputs()}
        </div>
      {/if}
    </div>
  {/if}
{/if}

<style lang="postcss">
  .control-bar {
    @apply flex items-center gap-x-3 min-w-0 overflow-hidden;
  }

  .static-controls {
    @apply flex flex-wrap gap-x-4 gap-y-0.5 text-xs text-fg-secondary;
  }

  .adjust {
    @apply flex items-center gap-x-1 h-6 px-2 text-xs whitespace-nowrap;
    @apply border rounded-sm bg-surface-background text-fg-primary;
  }

  .adjust:hover {
    @apply bg-surface-hover;
  }

  .adjust.changed {
    @apply border-primary-400 text-primary-700;
  }
</style>
