<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import {
    Popover,
    PopoverContent,
    PopoverTrigger,
  } from "@rilldata/web-common/components/popover";
  import {
    comparisonWidthPatch,
    PIVOT_COLUMN_ALIGNS,
    PIVOT_COMPARISONS,
    type ColumnSettingsCapabilities,
    type PivotComparison,
    type PivotFieldConfig,
    type PivotFieldConfigPatch,
  } from "@rilldata/web-common/features/canvas/components/pivot/field-config";
  import MeasureFormattingControls from "@rilldata/web-common/features/dashboards/pivot/MeasureFormattingControls.svelte";
  import PivotChip from "@rilldata/web-common/features/dashboards/pivot/PivotChip.svelte";
  import { roleWidthBounds } from "@rilldata/web-common/features/dashboards/pivot/pivot-column-width-utils";
  import {
    PivotChipType,
    type PivotChipData,
    type PivotColumnAlign,
    type PivotMeasureFormatting,
  } from "@rilldata/web-common/features/dashboards/pivot/types";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { FormatPreset } from "@rilldata/web-common/lib/number-formatting/humanizer-types";
  import { Palette, SlidersHorizontal } from "lucide-svelte";

  type Props = {
    item: PivotChipData;
    // The field's entry in the spec, if it is an object.
    config: PivotFieldConfig | undefined;
    // The header text used when no label override is set.
    defaultLabel: string;
    capabilities: ColumnSettingsCapabilities;
    removable?: boolean;
    grab?: boolean;
    fullWidth?: boolean;
    // Conditional formatting, offered when the chip is a measure.
    fmt?: PivotMeasureFormatting | undefined;
    lowerIsBetter?: boolean;
    onFormatChange?: ((fmt: PivotMeasureFormatting | null) => void) | undefined;
    onConfigChange: (patch: PivotFieldConfigPatch) => void;
    onRemove?: () => void;
    onmousedown?: (e: MouseEvent) => void;
  };

  let {
    item,
    config,
    defaultLabel,
    capabilities,
    removable = false,
    grab = false,
    fullWidth = false,
    fmt = undefined,
    lowerIsBetter = false,
    onFormatChange = undefined,
    onConfigChange,
    onRemove = () => {},
    onmousedown = undefined,
  }: Props = $props();

  let open = $state(false);

  const isMeasure = $derived(item.type === PivotChipType.Measure);
  const bounds = $derived(roleWidthBounds(isMeasure ? "measure" : "dimension"));
  const hasOverrides = $derived(
    !!config &&
      Object.keys(config).some((key) => key !== "name" && key in config),
  );

  // Text inputs edit a local draft and commit on blur or enter.
  let labelDraft = $state("");
  let widthDraft = $state("");
  let d3Draft = $state("");
  // The widths of the measure's delta and percent-change columns.
  let comparisonDrafts = $state<Record<PivotComparison, string>>({
    delta: "",
    percent_change: "",
  });
  $effect(() => {
    labelDraft = config?.label ?? "";
    widthDraft = config?.width !== undefined ? String(config.width) : "";
    d3Draft = config?.format_d3 ?? "";
    comparisonDrafts = {
      delta: comparisonWidthText("delta"),
      percent_change: comparisonWidthText("percent_change"),
    };
  });

  function comparisonWidthText(comparison: PivotComparison): string {
    const width = config?.[comparison]?.width;
    return width === undefined ? "" : String(width);
  }

  function commitLabel() {
    const value = labelDraft.trim();
    if (value === (config?.label ?? "")) return;
    onConfigChange({ label: value === "" ? null : value });
  }

  function commitWidth() {
    const text = widthDraft.trim();
    if (text === "") {
      if (config?.width !== undefined) onConfigChange({ width: null });
      return;
    }
    const parsed = Math.round(Number(text));
    if (!Number.isFinite(parsed)) {
      widthDraft = config?.width !== undefined ? String(config.width) : "";
      return;
    }
    const clamped = Math.min(bounds.max, Math.max(bounds.min, parsed));
    widthDraft = String(clamped);
    if (clamped !== config?.width) onConfigChange({ width: clamped });
  }

  const measureBounds = roleWidthBounds("measure");
  const comparisonWidthLabels: Record<PivotComparison, () => string> = {
    delta: m.canvas_column_delta_width_label,
    percent_change: m.canvas_column_percent_change_width_label,
  };

  function commitComparisonWidth(comparison: PivotComparison) {
    const current = config?.[comparison]?.width;
    const text = comparisonDrafts[comparison].trim();
    if (text === "") {
      if (current !== undefined) {
        onConfigChange(comparisonWidthPatch(comparison, null));
      }
      return;
    }
    const parsed = Math.round(Number(text));
    if (!Number.isFinite(parsed)) {
      comparisonDrafts[comparison] = comparisonWidthText(comparison);
      return;
    }
    const clamped = Math.min(
      measureBounds.max,
      Math.max(measureBounds.min, parsed),
    );
    comparisonDrafts[comparison] = String(clamped);
    if (clamped !== current) {
      onConfigChange(comparisonWidthPatch(comparison, clamped));
    }
  }

  function commitD3() {
    const value = d3Draft.trim();
    if (value === (config?.format_d3 ?? "")) return;
    onConfigChange(
      value === ""
        ? { format_d3: null }
        : { format_d3: value, format_preset: null },
    );
  }

  const wrapValue = $derived(
    config?.wrap === undefined ? "" : config.wrap ? "on" : "off",
  );
  const wrapOptions = [
    { value: "", label: m.canvas_column_option_default() },
    { value: "on", label: m.canvas_column_wrap_on() },
    { value: "off", label: m.canvas_column_wrap_off() },
  ];

  const alignLabels: Record<PivotColumnAlign, () => string> = {
    left: m.canvas_column_align_left,
    center: m.canvas_column_align_center,
    right: m.canvas_column_align_right,
  };
  const alignOptions = [
    { value: "", label: m.canvas_column_option_default() },
    ...PIVOT_COLUMN_ALIGNS.map((align) => ({
      value: align,
      label: alignLabels[align](),
    })),
  ];

  const presetLabels: Record<FormatPreset, () => string> = {
    [FormatPreset.HUMANIZE]: m.canvas_format_preset_humanize,
    [FormatPreset.NONE]: m.canvas_format_preset_none,
    [FormatPreset.CURRENCY_USD]: m.canvas_format_preset_currency_usd,
    [FormatPreset.CURRENCY_EUR]: m.canvas_format_preset_currency_eur,
    [FormatPreset.PERCENTAGE]: m.canvas_format_preset_percentage,
    [FormatPreset.INTERVAL]: m.canvas_format_preset_interval_ms,
  };
  const presetOptions = [
    { value: "", label: m.canvas_column_format_from_metrics_view() },
    ...Object.values(FormatPreset).map((preset) => ({
      value: preset,
      label: presetLabels[preset](),
    })),
  ];
</script>

<Popover bind:open>
  <PopoverTrigger>
    {#snippet child({ props })}
      <div {...props}>
        <PivotChip
          {item}
          {removable}
          {grab}
          {fullWidth}
          {onmousedown}
          {onRemove}
        >
          <div class="format-dropdown flex items-center gap-x-1" slot="body">
            {#if hasOverrides}
              <SlidersHorizontal size="12px" />
            {/if}
            {#if fmt}
              <Palette size="12px" />
            {/if}
            <span
              class={["flex-none transition-transform", open && "-rotate-180"]}
            >
              <CaretDownIcon size="12px" />
            </span>
          </div>
        </PivotChip>
      </div>
    {/snippet}
  </PopoverTrigger>
  <PopoverContent align="start" side="bottom" class="w-[340px] p-0">
    <div class="flex flex-col gap-y-0.5 px-3.5 pb-2.5 pt-3">
      <span class="text-sm font-semibold truncate" title={item.title}>
        {item.title}
      </span>
      <span class="text-xs text-fg-secondary">
        {m.canvas_column_settings_title()}
      </span>
    </div>
    <hr class="border-gray-200" />
    <div class="flex flex-col gap-y-3 px-3.5 pb-3 pt-2.5">
      <Input
        id="{item.id}-label"
        label={m.canvas_column_label_label()}
        capitalizeLabel={false}
        textClass="text-sm"
        size="sm"
        labelGap={2}
        placeholder={defaultLabel}
        bind:value={labelDraft}
        onBlur={commitLabel}
        onEnter={commitLabel}
      />

      {#if capabilities.width}
        <div class="flex items-end gap-x-2">
          <div class="grow">
            <Input
              id="{item.id}-width"
              inputType="number"
              label={m.canvas_column_width_label()}
              capitalizeLabel={false}
              textClass="text-sm"
              size="sm"
              labelGap={2}
              placeholder={m.canvas_column_width_auto()}
              hint={m.canvas_column_width_hint({
                min: String(bounds.min),
                max: String(bounds.max),
              })}
              bind:value={widthDraft}
              onBlur={commitWidth}
              onEnter={commitWidth}
            />
          </div>
          <Button
            type="secondary"
            small
            disabled={config?.width === undefined}
            onClick={() => {
              widthDraft = "";
              onConfigChange({ width: null });
            }}
          >
            {m.canvas_column_width_auto()}
          </Button>
        </div>
      {/if}

      {#if capabilities.comparison}
        <div class="flex flex-col gap-y-1">
          <div class="flex gap-x-2">
            {#each PIVOT_COMPARISONS as comparison (comparison)}
              <div class="grow min-w-0">
                <Input
                  id="{item.id}-{comparison}-width"
                  inputType="number"
                  label={comparisonWidthLabels[comparison]()}
                  capitalizeLabel={false}
                  textClass="text-sm"
                  size="sm"
                  labelGap={2}
                  placeholder={m.canvas_column_width_auto()}
                  bind:value={comparisonDrafts[comparison]}
                  onBlur={() => commitComparisonWidth(comparison)}
                  onEnter={() => commitComparisonWidth(comparison)}
                />
              </div>
            {/each}
          </div>
          <span class="text-xs text-fg-secondary">
            {m.canvas_column_comparison_width_hint()}
          </span>
        </div>
      {/if}

      {#if capabilities.wrap}
        <Select
          id="{item.id}-wrap"
          label={m.canvas_column_wrap_label()}
          options={wrapOptions}
          value={wrapValue}
          placeholder={m.canvas_column_option_default()}
          full
          size="sm"
          sameWidth
          fontSize={12}
          onChange={(value) =>
            onConfigChange({ wrap: value === "" ? null : value === "on" })}
        />
      {/if}

      {#if capabilities.align}
        <Select
          id="{item.id}-align"
          label={m.canvas_column_align_label()}
          options={alignOptions}
          value={config?.align ?? ""}
          placeholder={m.canvas_column_option_default()}
          full
          size="sm"
          sameWidth
          fontSize={12}
          onChange={(value) =>
            onConfigChange({
              align: value === "" ? null : (value as PivotColumnAlign),
            })}
        />
      {/if}

      {#if capabilities.format}
        <Select
          id="{item.id}-format-preset"
          label={m.canvas_column_format_label()}
          options={presetOptions}
          value={config?.format_preset ?? ""}
          placeholder={m.canvas_column_format_from_metrics_view()}
          full
          size="sm"
          sameWidth
          fontSize={12}
          onChange={(value) =>
            onConfigChange(
              value === ""
                ? { format_preset: null }
                : { format_preset: value, format_d3: null },
            )}
        />
        <Input
          id="{item.id}-format-d3"
          label={m.canvas_column_format_d3_label()}
          capitalizeLabel={false}
          textClass="text-sm"
          size="sm"
          labelGap={2}
          placeholder={m.canvas_column_format_d3_placeholder()}
          bind:value={d3Draft}
          onBlur={commitD3}
          onEnter={commitD3}
        />
      {/if}
    </div>

    {#if onFormatChange}
      <hr class="border-gray-200" />
      <div class="px-3.5 pb-3 pt-2.5">
        <MeasureFormattingControls
          id={item.id}
          {fmt}
          {lowerIsBetter}
          onChange={onFormatChange}
        />
      </div>
    {/if}
  </PopoverContent>
</Popover>
