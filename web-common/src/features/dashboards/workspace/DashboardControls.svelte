<script lang="ts">
  import Filters from "../filters/Filters.svelte";
  import TimeControls from "../time-controls/TimeControls.svelte";

  export let metricViewName: string;
</script>

<div class="p-1 py-2 space-y-2">
  <TimeControls {metricViewName} />
  {#key metricViewName}
    <Filters />
  {/key}
</div>
