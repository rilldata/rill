<script lang="ts">
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import {
    DimensionFilterMode,
    DimensionFilterModeOptions,
  } from "@rilldata/web-common/features/dashboards/filters/dimension-filters/constants";

  export let mode: DimensionFilterMode;
  export let disabled: boolean = false;
  export let size: "sm" | "md" | "lg" = "md";
  export let onModeChange: (mode: DimensionFilterMode) => void = () => {};

  function handleModeChange(newMode: DimensionFilterMode) {
    onModeChange(newMode);
  }
</script>

<Select
  id="dimension-filter-mode-selector"
  bind:value={mode}
  options={DimensionFilterModeOptions}
  onChange={handleModeChange}
  {size}
  {disabled}
  minWidth={82}
  forcedTriggerStyle="rounded-r-none"
/>
