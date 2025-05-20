<script lang="ts">
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import SettingsSlider from "@rilldata/web-common/components/icons/SettingsSlider.svelte";
  import * as Popover from "@rilldata/web-common/components/popover";
  import type {
    ChartLegend,
    ChartSortDirection,
    FieldConfig,
  } from "@rilldata/web-common/features/canvas/components/charts/types";
  import type { ChartFieldInput } from "@rilldata/web-common/features/canvas/inspector/types";

  export let fieldConfig: FieldConfig;
  export let onChange: (property: keyof FieldConfig, value: any) => void;
  export let chartFieldInput: ChartFieldInput | undefined = undefined;
  export let label: string;

  $: isDimension = fieldConfig?.type === "nominal";
  $: isMeasure = fieldConfig?.type === "quantitative";

  let limit = fieldConfig?.limit || 5000;
  let labelAngle =
    fieldConfig?.labelAngle ?? (fieldConfig?.type === "temporal" ? 0 : -90);
  let isDropdownOpen = false;

  const sortOptions: { label: string; value: ChartSortDirection }[] = [
    { label: "Ascending", value: "x" },
    { label: "Descending", value: "-x" },
    { label: "Y-axis ascending", value: "y" },
    { label: "Y-axis descending", value: "-y" },
  ];

  const legendOptions: { label: string; value: ChartLegend }[] = [
    { label: "Top", value: "top" },
    { label: "Right", value: "right" },
    { label: "Bottom", value: "bottom" },
    { label: "Left", value: "left" },
    { label: "None", value: "none" },
  ];

  $: showAxisTitle = chartFieldInput?.axisTitleSelector ?? false;
  $: showOrigin = chartFieldInput?.originSelector ?? false;
  $: showSort = chartFieldInput?.sortSelector ?? false;
  $: showLimit = chartFieldInput?.limitSelector ?? false;
  $: showNull = chartFieldInput?.nullSelector ?? false;
  $: showLabelAngle = chartFieldInput?.labelAngleSelector ?? false;
  $: showLegend = chartFieldInput?.defaultLegendOrientation ?? false;
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
        {#if showSort}
          <div class="py-1 flex items-center justify-between">
            <span class="text-xs">Sort</span>
            <Select
              size="sm"
              id="sort-select"
              width={180}
              options={sortOptions}
              value={fieldConfig?.sort || "x"}
              on:change={(e) => onChange("sort", e.detail)}
            />
          </div>
        {/if}
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
      {#if isMeasure && showOrigin}
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
    </div>
  </Popover.Content>
</Popover.Root>
