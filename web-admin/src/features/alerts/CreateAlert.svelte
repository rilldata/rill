<script lang="ts">
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { useMetricsView } from "@rilldata/web-common/features/dashboards/selectors";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { Button } from "@rilldata/web-common/components/button";
  import { BellPlusIcon } from "lucide-svelte";
  import CreateAlertDialog from "@rilldata/web-common/features/alerts/CreateAlertDialog.svelte";

  const {
    selectors: {
      timeRangeSelectors: { isCustomTimeRange },
    },
    runtime,
    metricsViewName,
  } = getStateManagers();

  $: metricsView = useMetricsView($runtime?.instanceId, $metricsViewName);
  $: hasTimeDimension = !!$metricsView?.data?.timeDimension;

  let showAlertDialog = false;
</script>

{#if hasTimeDimension}
  <Tooltip distance={8} location="top" suppress={!$isCustomTimeRange}>
    <Button
      compact
      disabled={$isCustomTimeRange}
      on:click={() => (showAlertDialog = true)}
      type="secondary"
    >
      <BellPlusIcon class="inline-flex" size="16px" />
    </Button>
    <TooltipContent slot="tooltip-content">
      To create an alert, set a non-custom time range.
    </TooltipContent>
  </Tooltip>
{/if}

<!-- Including `showAlertDialog` in the conditional ensures we tear 
    down the form state when the dialog closes -->
{#if showAlertDialog}
  <svelte:component
    this={CreateAlertDialog}
    open={showAlertDialog}
    on:close={() => (showAlertDialog = false)}
  />
{/if}
