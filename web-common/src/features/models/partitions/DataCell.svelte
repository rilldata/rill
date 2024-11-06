<script lang="ts">
  import type { V1ModelPartitionData } from "../../../runtime-client";

  export let data: V1ModelPartitionData;
</script>

<div class="data-cell">
  {#each Object.entries(data) as [key, value] (key)}
    <div class="data-item">
      <span class="data-key">{key}:</span>
      <span class="data-value">{value}</span>
    </div>
  {/each}
</div>

<style lang="postcss">
  .data-cell {
    @apply flex flex-col gap-y-2;
  }
  .data-key {
    @apply font-semibold;
  }
  .data-value {
    @apply font-medium;
  }
</style>
