<script lang="ts">
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import FormattingRulesEditor from "./FormattingRulesEditor.svelte";
  import {
    getSchemeStops,
    PIVOT_FORMATTING_SCHEMES,
  } from "./pivot-conditional-formatting";
  import type { PivotFormatRule, PivotMeasureFormatting } from "./types";

  type Props = {
    // Unique per rendered instance, used to build stable input ids.
    id: string;
    fmt: PivotMeasureFormatting | undefined;
    lowerIsBetter?: boolean;
    // Emits the new config, or null to clear formatting.
    onChange: (fmt: PivotMeasureFormatting | null) => void;
  };

  let { id, fmt, lowerIsBetter = false, onChange }: Props = $props();

  const DEFAULT_SCHEME = "theme-sequential";

  const modeOptions = [
    { value: "none", label: "None" },
    { value: "heatmap", label: "Heatmap" },
    { value: "data_bar", label: "Data bar" },
    { value: "rules", label: "Rules" },
  ];

  const schemeOptions = PIVOT_FORMATTING_SCHEMES.map((s) => ({
    value: s.key,
    label: s.label,
  }));

  let mode = $derived(fmt?.mode ?? "none");
  let scheme = $derived(
    fmt && fmt.mode !== "rules" ? fmt.scheme : DEFAULT_SCHEME,
  );
  let rules = $derived(fmt?.mode === "rules" ? fmt.rules : []);

  function swatchGradient(schemeKey: string, reversed: boolean): string {
    return `linear-gradient(to right, ${getSchemeStops(schemeKey, reversed).join(", ")})`;
  }

  function handleModeChange(newMode: string) {
    if (newMode === "none") {
      onChange(null);
    } else if (newMode === "rules") {
      // Seed with a starter rule so the editor never opens empty.
      onChange({
        mode: "rules",
        rules: rules.length
          ? rules
          : [{ operator: "gt", value: 0, color: "positive" }],
      });
    } else {
      onChange({ mode: newMode as "heatmap" | "data_bar", scheme });
    }
  }

  function handleRulesChange(newRules: PivotFormatRule[]) {
    if (newRules.length === 0) {
      onChange(null);
      return;
    }
    onChange({ mode: "rules", rules: newRules });
  }
</script>

<div class="flex flex-col gap-y-1.5">
  <div class="flex items-center justify-between gap-x-2">
    <span class="text-xs font-semibold">Format style</span>
    <Select
      id="pivot-format-mode-{id}"
      value={mode}
      options={modeOptions}
      onChange={handleModeChange}
      size="sm"
      width={104}
    />
  </div>

  {#if mode === "heatmap" || mode === "data_bar"}
    <div class="flex items-center gap-x-2">
      <div
        class="swatch"
        style:background={swatchGradient(
          scheme,
          mode === "heatmap" && lowerIsBetter,
        )}
      ></div>
      <Select
        id="pivot-format-scheme-{id}"
        value={scheme}
        options={schemeOptions}
        onChange={(v: string) => onChange({ mode, scheme: v })}
        size="sm"
        full
      />
    </div>
  {:else if mode === "rules"}
    <FormattingRulesEditor {rules} onChange={handleRulesChange} />
  {/if}
</div>

<style lang="postcss">
  .swatch {
    @apply h-5 w-8 flex-none rounded-sm border border-gray-200;
  }
</style>
