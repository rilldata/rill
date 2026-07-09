<script lang="ts">
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import {
    getSchemeStops,
    PIVOT_FORMATTING_SCHEMES,
  } from "./pivot-conditional-formatting";
  import type { PivotMeasureFormatting } from "./types";

  // Measures available to format, in display order.
  export let measures: { name: string; label: string }[];
  export let measureFormatting:
    | Record<string, PivotMeasureFormatting>
    | undefined;
  // Emits the new config for a measure, or null to clear formatting.
  export let onChange: (
    measureName: string,
    fmt: PivotMeasureFormatting | null,
  ) => void;

  const DEFAULT_SCHEME = "theme-sequential";

  const modeOptions = [
    { value: "none", label: "None" },
    { value: "heatmap", label: "Heatmap" },
    { value: "data_bar", label: "Data bar" },
  ];

  const schemeOptions = PIVOT_FORMATTING_SCHEMES.map((s) => ({
    value: s.key,
    label: s.label,
  }));

  function swatchGradient(scheme: string, reverse: boolean): string {
    return `linear-gradient(to right, ${getSchemeStops(scheme, reverse).join(", ")})`;
  }

  function handleModeChange(measureName: string, mode: string) {
    if (mode === "none") {
      onChange(measureName, null);
      return;
    }
    const current = measureFormatting?.[measureName];
    onChange(measureName, {
      mode: mode as PivotMeasureFormatting["mode"],
      scheme: current?.scheme ?? DEFAULT_SCHEME,
      reverse: current?.reverse,
    });
  }

  function handleSchemeChange(measureName: string, scheme: string) {
    const current = measureFormatting?.[measureName];
    if (!current) return;
    onChange(measureName, { ...current, scheme });
  }

  function handleReverseChange(measureName: string, reverse: boolean) {
    const current = measureFormatting?.[measureName];
    if (!current) return;
    onChange(measureName, { ...current, reverse });
  }
</script>

<div class="flex flex-col gap-y-3">
  {#if measures.length === 0}
    <span class="text-fg-secondary text-xs">No measures to format.</span>
  {/if}
  {#each measures as measure (measure.name)}
    {@const fmt = measureFormatting?.[measure.name]}
    {@const mode = fmt?.mode ?? "none"}
    {@const scheme = fmt?.scheme ?? DEFAULT_SCHEME}
    <div class="flex flex-col gap-y-1.5">
      <div class="flex items-center justify-between gap-x-2">
        <span class="text-xs font-medium truncate" title={measure.label}>
          {measure.label}
        </span>
        <Select
          id="pivot-format-mode-{measure.name}"
          value={mode}
          options={modeOptions}
          onChange={(v) => handleModeChange(measure.name, v)}
          size="sm"
          width={104}
        />
      </div>
      {#if mode !== "none"}
        <div class="flex items-center gap-x-2 pl-1">
          <div
            class="swatch"
            style:background={swatchGradient(scheme, fmt?.reverse ?? false)}
          ></div>
          <Select
            id="pivot-format-scheme-{measure.name}"
            value={scheme}
            options={schemeOptions}
            onChange={(v) => handleSchemeChange(measure.name, v)}
            size="sm"
            full
          />
        </div>
        {#if mode === "heatmap"}
          <div class="flex items-center justify-between gap-x-2 pl-1">
            <span class="text-fg-secondary text-xs">Reverse colors</span>
            <Switch
              small
              checked={fmt?.reverse ?? false}
              onCheckedChange={(checked) =>
                handleReverseChange(measure.name, Boolean(checked))}
            />
          </div>
        {/if}
      {/if}
    </div>
  {/each}
</div>

<style lang="postcss">
  .swatch {
    @apply h-5 w-8 flex-none rounded-sm border border-gray-200;
  }
</style>
