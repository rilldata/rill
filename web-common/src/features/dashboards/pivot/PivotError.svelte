<script lang="ts">
  import type { PivotQueryError } from "./types";

  export let errors: PivotQueryError[];

  function removeDuplicates(errors: PivotQueryError[]): PivotQueryError[] {
    const seen = new Set();
    return errors.filter((error) => {
      const key = `${error.statusCode}-${error.message}`;
      if (seen.has(key)) {
        return false;
      } else {
        seen.add(key);
        return true;
      }
    });
  }

  let uniqueErrors = removeDuplicates(errors);
</script>

<div class="flex flex-col items-center w-full h-full">
  <span class="text-3xl font-normal m-2">Sorry, unexpected query error!</span>
  <div class="text-base text-gray-600 mt-4">
    One or more APIs failed with the following error{uniqueErrors.length !== 1
      ? "s"
      : ""}:
  </div>

  {#each uniqueErrors as error}
    <div class="flex text-base gap-x-2">
      <span class="text-red-600 font-semibold">{error.statusCode} :</span>
      <span class="text-gray-700">{error.message}</span>
    </div>
  {/each}
</div>
