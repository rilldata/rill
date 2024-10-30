<script lang="ts">
  import type { PivotQueryError } from "./types";

  export let errors: PivotQueryError[];
</script>

<div class="flex flex-col items-center w-full h-full">
  <span class="text-2xl font-normal m-2">Query Error</span>
  <span class="text-md text-gray-600 mb-4">
    Unable to generate the pivot table due to the following errors:
  </span>

  {#each errors as error}
    <div>
      <span class="text-red-600 font-bold">{error.statusCode}</span> :
      <span class="text-gray-700 mt-2">{error.message}</span>
    </div>
  {/each}
</div>
