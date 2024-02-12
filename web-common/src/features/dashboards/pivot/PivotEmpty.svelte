<script>
  import EmptyTableIcon from "./EmptyTableIcon.svelte";

  export let isFetching = false;
</script>

<div class="flex flex-col items-center w-full h-full justify-center gap-y-3">
  {#if isFetching}
    <!-- Add building pivot table svg -->
    <div class="font-semibold text-gray-800 mt-1 text-lg">
      Hold tight, we are building your pivot table.
    </div>
  {:else}
    <EmptyTableIcon />
    <div class="font-semibold text-gray-800 mt-1 text-lg">
      Your pivot table looks lonely
    </div>
    <div class="text-gray-600 text-base">
      Give it some data to keep it company.
    </div>
    <div class="text-gray-600">Learn more about pivot tables in our docs.</div>
  {/if}
</div>
