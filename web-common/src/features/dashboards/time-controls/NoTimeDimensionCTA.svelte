<script lang="ts">
  import Calendar from "@rilldata/web-common/components/icons/Calendar.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { featureFlags } from "../../feature-flags";

  const { readOnly } = featureFlags;

  $: isReadOnlyDashboard = $readOnly === true;
</script>

<Tooltip location="bottom" distance={8}>
  <div class="px-3 py-2 flex items-center gap-x-3">
    <span class="ui-copy-icon"><Calendar size="16px" /></span>
    <span class="ui-copy-disabled">No time dimension specified</span>
  </div>
  <TooltipContent slot="tooltip-content" maxWidth="250px">
    {#if isReadOnlyDashboard}
      No time dimension available for this dashboard.
    {:else}
      Add a time dimension to your dashboard definition to enable time series
      plots.
    {/if}
  </TooltipContent>
</Tooltip>
