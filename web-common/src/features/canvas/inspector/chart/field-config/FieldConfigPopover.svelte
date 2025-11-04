<script lang="ts">
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import SettingsSlider from "@rilldata/web-common/components/icons/SettingsSlider.svelte";
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
  let labelAngle =
    fieldConfig?.labelAngle ?? (fieldConfig?.type === "temporal" ? 0 : -90);
  let isDropdownOpen = false;

  const legendOptions: { label: string; value: ChartLegend }[] = [
    { label: "Top", value: "top" },
    { label: "Right", value: "right" },
    { label: "Bottom", value: "bottom" },
    { label: "Left", value: "left" },
    { label: "None", value: "none" },
  ];

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
      <span class="text-xs font-medium">{label} Configuration</span>
    </div>
    <div class="px-3.5 pb-1.5">
      {#if showLegend}
        <div class="py-1 flex items-center justify-between">
          <span class="text-xs">Legend orientation</span>
          <Select
            size="sm"
            id="legend-orientation-select"
            width={180}
            options={legendOptions}
            value={fieldConfig?.legendOrientation ||
              chartFieldInput?.defaultLegendOrientation}
            on:change={(e) => onChange("legendOrientation", e.detail)}
          />
        </div>
      {/if}
      {#if showAxisTitle}
        <div class="py-1.5 flex items-center justify-between">
          <span class="text-xs">Show axis title</span>
          <Switch
            small
            checked={fieldConfig?.showAxisTitle}
            on:click={() => {
              onChange("showAxisTitle", !fieldConfig?.showAxisTitle);
            }}
          />
        </div>
      {/if}
      {#if isDimension}
        {#if showNull}
          <div class="py-1.5 flex items-center justify-between">
            <span class="text-xs">Show null values</span>
            <Switch
              small
              checked={fieldConfig?.showNull}
              on:click={() => {
                onChange("showNull", !fieldConfig?.showNull);
              }}
            />
          </div>
        {/if}
        <SortConfig {fieldConfig} {onChange} {sortConfig} />
        {#if showLimit}
          <div class="py-1 flex items-center justify-between">
            <span class="text-xs">Limit</span>
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
            <span class="text-xs">Zero based origin</span>
            <Switch
              small
              checked={fieldConfig?.zeroBasedOrigin}
              on:click={() => {
                onChange("zeroBasedOrigin", !fieldConfig?.zeroBasedOrigin);
              }}
            />
          </div>
        {/if}
        {#if showTotal}
          <div class="py-1.5 flex items-center justify-between">
            <span class="text-xs">Show totals value</span>
            <Switch
              small
              checked={fieldConfig?.showTotal}
              on:click={() => {
                onChange("showTotal", !fieldConfig?.showTotal);
              }}
            />
          </div>
        {/if}
        {#if showAxisRange}
          <div class="py-1.5 flex items-center justify-between">
            <span class="text-xs">Min</span>
            <Input
              size="sm"
              width="120px"
              id="axis-min-value-select"
              inputType="number"
              placeholder="Enter a number"
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
            <span class="text-xs">Max</span>
            <Input
              size="sm"
              width="120px"
              id="axis-min-value-select"
              inputType="number"
              placeholder="Enter a number"
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
          <span class="text-xs">Label angle</span>
          <Input
            size="sm"
            width="72px"
            id="label-angle-select"
            inputType="number"
            bind:value={labelAngle}
            onBlur={() => {
              onChange("labelAngle", labelAngle);
            }}
            onEnter={() => {
              onChange("labelAngle", labelAngle);
            }}
          />
        </div>
      {/if}
    </div>
  </Popover.Content>
</Popover.Root>
