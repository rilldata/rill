<script lang="ts">
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import type {
    PivotCanvasComponent,
    PivotConditionalFormatSpec,
  } from "@rilldata/web-common/features/canvas/components/pivot";
  import PivotFormattingPicker from "@rilldata/web-common/features/dashboards/pivot/PivotFormattingPicker.svelte";
  import type { PivotMeasureFormatting } from "@rilldata/web-common/features/dashboards/pivot/types";
  import { get } from "svelte/store";

  export let component: PivotCanvasComponent;
  export let label: string | undefined = undefined;

  $: spec = component.specStore;
  $: measureNames = ("measures" in $spec ? $spec.measures : []) ?? [];
  $: cfArray =
    ("conditional_format" in $spec ? $spec.conditional_format : []) ?? [];

  // Resolve measure display names from the metrics view, falling back to the name.
  $: metricsViewSpec = $spec.metrics_view
    ? get(
        component.parent.metricsView.getMetricsViewFromName($spec.metrics_view),
      ).metricsView
    : undefined;
  $: measures = measureNames.map((name) => ({
    name,
    label:
      metricsViewSpec?.measures?.find((m) => m.name === name)?.displayName ??
      name,
  }));

  $: measureFormatting = Object.fromEntries(
    cfArray.map((f) => [
      f.measure,
      { mode: f.mode, scheme: f.scheme, reverse: f.reverse },
    ]),
  ) as Record<string, PivotMeasureFormatting>;

  function onChange(measureName: string, fmt: PivotMeasureFormatting | null) {
    const next = { ...measureFormatting };
    if (fmt) {
      next[measureName] = fmt;
    } else {
      delete next[measureName];
    }
    const updated: PivotConditionalFormatSpec[] = Object.entries(next).map(
      ([measure, f]) => ({
        measure,
        mode: f.mode,
        scheme: f.scheme,
        ...(f.reverse ? { reverse: true } : {}),
      }),
    );
    component.updateProperty("conditional_format", updated);
  }
</script>

<div class="flex flex-col gap-y-2">
  <InputLabel
    small
    label={label ?? "Conditional formatting"}
    id="conditional_format"
  />
  <PivotFormattingPicker {measures} {measureFormatting} {onChange} />
</div>
