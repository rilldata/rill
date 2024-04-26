<script lang="ts">
  import DataTypeIcon from "@rilldata/web-common/components/data-types/DataTypeIcon.svelte";
  import type { V1MetricsViewColumn } from "@rilldata/web-common/runtime-client";

  export let name: V1MetricsViewColumn["name"];
  export let type: V1MetricsViewColumn["type"];
  export let sorted: boolean = false;
</script>

<div class="header-cell" class:sorted>
  <DataTypeIcon suppressTooltip color={"text-gray-500"} {type} />

  <p class="truncate w-full grow text-left">
    {name}
  </p>

  <slot name="pin-button" />
</div>

<style lang="postcss">
  .header-cell {
    @apply flex gap-x-2 px-4 items-center relative size-full;
    @apply bg-gray-100 font-bold;
  }
</style>
