<script lang="ts">
  import CreateAlertDialog from "./CreateAlertDialog.svelte";

  let showAlertDialog = false;
</script>

<!-- TODO: match Figma mocks -->
<button
  class="h-6 px-1.5 py-px flex items-center gap-[3px] rounded-sm hover:bg-gray-200 text-gray-700"
  on:click={() => {
    showAlertDialog = true;
  }}
>
  Create alert
</button>

<!-- Including `showAlertDialog` in the conditional ensures we tear 
  down the form state when the dialog closes -->
{#if showAlertDialog}
  <svelte:component
    this={CreateAlertDialog}
    open={showAlertDialog}
    on:close={() => (showAlertDialog = false)}
  />
{/if}
