<script lang="ts">
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import SettingsSlider from "@rilldata/web-common/components/icons/SettingsSlider.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import * as Popover from "@rilldata/web-common/components/popover";
  import type { ChartFieldInput } from "@rilldata/web-common/features/canvas/inspector/types";
  import type {
    ChartLegend,
    FieldConfig,
  } from "@rilldata/web-common/features/components/charts/types";
  import SortConfig from "./SortConfig.svelte";

  export let fieldConfig: FieldConfig;
  export let onChange: (property: keyof FieldConfig, value: any) => void;
  export let chartFieldInput: ChartFieldInput | undefined = undefined;
  export let label: string;

  $: isDimension = fieldConfig?.type === "nominal";
  $: isMeasure = fieldConfig?.type === "quantitative";

  let limit =
    fieldConfig?.limit || chartFieldInput?.limitSelector?.defaultLimit || 5000;
  let min = fieldConfig?.min;
  let max = fieldConfig?.max;
  let isDropdownOpen = false;

  // An unset angle lets the chart pick the orientation from the available width.
  const labelOrientationAngles: Record<string, number | undefined> = {
    auto: undefined,
    horizontal: 0,
    angled: -45,
    vertical: -90,
  };

  $: labelOrientation =
    Object.keys(labelOrientationAngles).find(
      (orientation) =>
        labelOrientationAngles[orientation] === fieldConfig?.labelAngle,
    ) ?? "custom";
  // A YAML angle outside the presets is shown as a read-only entry so it is not silently lost.
  $: labelOrientationOptions = [
    ...(labelOrientation === "custom"
      ? [
          {
            label: m.canvas_label_orientation_custom({
              angle: String(fieldConfig?.labelAngle),
            }),
            value: "custom",
            disabled: true,
          },
        ]
      : []),
    { label: m.canvas_label_orientation_auto(), value: "auto" },
    { label: m.canvas_label_orientation_horizontal(), value: "horizontal" },
    { label: m.canvas_label_orientation_angled(), value: "angled" },
    { label: m.canvas_label_orientation_vertical(), value: "vertical" },
  ];

  $: legendOptions = [
    { label: m.canvas_legend_top(), value: "top" },
    { label: m.canvas_legend_right(), value: "right" },
    { label: m.canvas_legend_bottom(), value: "bottom" },
    { label: m.canvas_legend_left(), value: "left" },
    { label: m.canvas_legend_none(), value: "none" },
  ] as { label: string; value: ChartLegend }[];

  $: showAxisTitle = chartFieldInput?.axisTitleSelector ?? false;
  $: showOrigin = chartFieldInput?.originSelector ?? false;
  $: sortConfig = chartFieldInput?.sortSelector ?? { enable: false };
  $: showLimit = chartFieldInput?.limitSelector ?? false;
  $: showTotal = chartFieldInput?.totalSelector ?? false;
  $: showNull = chartFieldInput?.nullSelector ?? false;
  $: showLabelAngle = chartFieldInput?.labelAngleSelector ?? false;
  $: showLegend = chartFieldInput?.defaultLegendOrientation ?? false;
  $: showAxisRange = chartFieldInput?.axisRangeSelector ?? false;
</script>

<Popover.Root bind:open={isDropdownOpen}>
  <Popover.Trigger class="flex-none">
    <IconButton rounded active={isDropdownOpen}>
      <SettingsSlider size="14px" />
    </IconButton>
  </Popover.Trigger>
  <Popover.Content align="start" class="w-[280px] p-0 overflow-visible">
    <div class="px-3.5 py-2 border-b border-gray-200">
      <span class="text-xs font-medium"
        >{m.canvas_configuration({ label })}</span
      >
    </div>
    <div class="px-3.5 pb-1.5">
      {#if showLegend}
        <div class="py-1 flex items-center justify-between">
          <span class="text-xs">{m.canvas_legend_orientation()}</span>
          <Select
            size="sm"
            id="legend-orientation-select"
            width={180}
            options={legendOptions}
            value={fieldConfig?.legendOrientation ||
              chartFieldInput?.defaultLegendOrientation}
            onChange={(value) => onChange("legendOrientation", value)}
          />
        </div>
      {/if}
      {#if showAxisTitle}
        <div class="py-1.5 flex items-center justify-between">
          <span class="text-xs">{m.canvas_show_axis_title()}</span>
          <Switch
            small
            checked={fieldConfig?.showAxisTitle}
            onclick={() => {
              onChange("showAxisTitle", !fieldConfig?.showAxisTitle);
            }}
          />
        </div>
      {/if}
      {#if isDimension}
        {#if showNull}
          <div class="py-1.5 flex items-center justify-between">
            <span class="text-xs">{m.canvas_show_null_values()}</span>
            <Switch
              small
              checked={fieldConfig?.showNull}
              onclick={() => {
                onChange("showNull", !fieldConfig?.showNull);
              }}
            />
          </div>
        {/if}
        <SortConfig {fieldConfig} {onChange} {sortConfig} />
        {#if showLimit}
          <div class="py-1 flex items-center justify-between">
            <span class="text-xs">{m.canvas_limit()}</span>
            <Input
              size="sm"
              width="72px"
              id="limit-select"
              inputType="number"
              bind:value={limit}
              onBlur={() => {
                onChange("limit", limit);
              }}
              onEnter={() => {
                onChange("limit", limit);
              }}
            />
          </div>
        {/if}
      {/if}
      {#if isMeasure}
        {#if showOrigin}
          <div class="py-1.5 flex items-center justify-between">
            <span class="text-xs">{m.canvas_zero_based_origin()}</span>
            <Switch
              small
              checked={fieldConfig?.zeroBasedOrigin}
              onclick={() => {
                onChange("zeroBasedOrigin", !fieldConfig?.zeroBasedOrigin);
              }}
            />
          </div>
        {/if}
        {#if showTotal}
          <div class="py-1.5 flex items-center justify-between">
            <span class="text-xs">{m.canvas_show_totals_value()}</span>
            <Switch
              small
              checked={fieldConfig?.showTotal}
              onclick={() => {
                onChange("showTotal", !fieldConfig?.showTotal);
              }}
            />
          </div>
        {/if}
        {#if showAxisRange}
          <div class="py-1.5 flex items-center justify-between">
            <span class="text-xs">{m.canvas_min()}</span>
            <Input
              size="sm"
              width="120px"
              id="axis-min-value-select"
              inputType="number"
              placeholder={m.canvas_enter_a_number()}
              bind:value={min}
              onBlur={() => {
                onChange("min", min);
              }}
              onEnter={() => {
                onChange("min", min);
              }}
            />
          </div>
          <div class="py-1.5 flex items-center justify-between">
            <span class="text-xs">{m.canvas_max()}</span>
            <Input
              size="sm"
              width="120px"
              id="axis-min-value-select"
              inputType="number"
              placeholder={m.canvas_enter_a_number()}
              bind:value={max}
              onBlur={() => {
                onChange("max", max);
              }}
              onEnter={() => {
                onChange("max", max);
              }}
            />
          </div>
        {/if}
      {/if}
      {#if showLabelAngle && fieldConfig?.type !== "temporal"}
        <div class="py-1 flex items-center justify-between">
          <span class="text-xs">{m.canvas_label_orientation()}</span>
          <Select
            size="sm"
            id="label-orientation-select"
            width={180}
            options={labelOrientationOptions}
            value={labelOrientation}
            onChange={(value: string) =>
              onChange("labelAngle", labelOrientationAngles[value])}
          />
        </div>
      {/if}
    </div>
  </Popover.Content>
</Popover.Root>
