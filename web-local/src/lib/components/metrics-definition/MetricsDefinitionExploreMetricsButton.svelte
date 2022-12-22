<script lang="ts">
  import { goto } from "$app/navigation";
  import { Button } from "@rilldata/web-common/components/button";
  import ExploreIcon from "@rilldata/web-common/components/icons/Explore.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { navigationEvent } from "../../metrics/initMetrics";
  import { BehaviourEventMedium } from "../../metrics/service/BehaviourEventTypes";
  import {
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "../../metrics/service/MetricsTypes";

  export let metricsInternalRep;
  export let metricsDefName;

  $: measures = $metricsInternalRep.getMeasures();
  $: dimensions = $metricsInternalRep.getDimensions();

  let buttonDisabled = true;
  let buttonStatus = "OK";

  const viewDashboard = () => {
    goto(`/dashboard/${metricsDefName}`);

    navigationEvent.fireEvent(
      metricsDefName,
      BehaviourEventMedium.Button,
      MetricsEventSpace.Workspace,
      MetricsEventScreenName.MetricsDefinition,
      MetricsEventScreenName.Dashboard
    );
  };

  $: if (
    $metricsInternalRep.getMetricKey("model") === "" ||
    $metricsInternalRep.getMetricKey("timeseries") === ""
  ) {
    buttonDisabled = true;
    buttonStatus = "MISSING_MODEL_OR_TIMESTAMP";
  } else if (
    // check if all the measures have a valid expression
    measures?.filter((measure) => measure?.expression?.length)?.length === 0 ||
    // and if the dimensions all have a valid property
    dimensions?.filter((dimension) => dimension?.property?.length)?.length === 0
  ) {
    buttonDisabled = true;
    buttonStatus = "MISSING_MEASURES_OR_DIMENSIONS";
  } else {
    buttonStatus = "NO_ERROR";
    buttonDisabled = false;
  }
</script>

<Tooltip alignment="middle" distance={5} location="right">
  <!-- TODO: we need to standardize these buttons. -->
  <Button
    disabled={buttonDisabled}
    on:click={() => viewDashboard()}
    type="primary"
  >
    Go to Dashboard <ExploreIcon size="16px" />
  </Button>
  <TooltipContent slot="tooltip-content">
    <div>
      {#if buttonStatus === "MISSING_MODEL_OR_TIMESTAMP"}
        Select a model and a timestamp column before exploring metrics
      {:else if buttonStatus === "MISSING_MEASURES_OR_DIMENSIONS"}
        Add measures and dimensions before exploring metrics
      {:else}
        Explore the metrics dashboard
      {/if}
    </div>
  </TooltipContent>
</Tooltip>
