<script lang="ts">
  import { onMount } from "svelte";
  import Tooltip from "../../components/tooltip/Tooltip.svelte";
  import TooltipContent from "../../components/tooltip/TooltipContent.svelte";
  import { getStateManagers } from "../dashboards/state-managers/state-managers";

  const {
    selectors: {
      timeRangeSelectors: { isCustomTimeRange },
    },
  } = getStateManagers();

  let showAlertDialog = false;

  // Only import the Create Alert dialog if in the Cloud context.
  // This ensures Rill Developer doesn't try and fail to import the admin-client.
  let CreateAlertDialog;
  onMount(async () => {
    CreateAlertDialog = (await import("./CreateAlertDialog.svelte")).default;
  });
</script>

<Tooltip location="top" distance={8} suppress={!$isCustomTimeRange}>
  <button
    disabled={$isCustomTimeRange}
    class="h-6 px-1.5 py-px flex items-center gap-[3px] rounded-sm hover:bg-gray-200 text-gray-700 disabled:cursor-not-allowed disabled:text-gray-400 disabled:hover:bg-gray-100"
    on:click={() => {
      showAlertDialog = true;
    }}
  >
    Create alert
  </button>
  <TooltipContent slot="tooltip-content">
    To create an alert, set a non-custom time range.
  </TooltipContent>
</Tooltip>

<!-- Including `showAlertDialog` in the conditional ensures we tear 
  down the form state when the dialog closes -->
{#if showAlertDialog}
  <svelte:component
    this={CreateAlertDialog}
    open={showAlertDialog}
    on:close={() => (showAlertDialog = false)}
  />
{/if}
