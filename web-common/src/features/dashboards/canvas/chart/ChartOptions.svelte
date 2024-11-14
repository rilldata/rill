<script lang="ts">
  import AddMeasureDimensionButton from "../AddMeasureDimensionButton.svelte";

  export let chartType = "bar";

  let selectedXAxis = null;
  let selectedYAxis = null;
  function handleAddField(event) {
    const { detail: fieldName } = event;

    if (!selectedXAxis) {
      selectedXAxis = fieldName;
    } else {
      selectedYAxis = fieldName;
    }
  }
</script>

<div class="chart-options">
  <div class="channel-name">X Axis:</div>
  {selectedXAxis || "Select a field"}
  <AddMeasureDimensionButton on:addField={handleAddField} />

  <div class="channel-name">Y Axis:</div>
  {selectedYAxis || "Select a field"}
  <AddMeasureDimensionButton on:addField={handleAddField} />
</div>

<style lang="postcss">
  .chart-options {
    @apply px-4 py-2;
  }
  .channel-name {
    @apply font-medium;
  }
</style>
