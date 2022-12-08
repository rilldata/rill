<script lang="ts">
  import { goto } from "$app/navigation";
  import { BehaviourEventMedium } from "@rilldata/web-local/lib/metrics/service/BehaviourEventTypes";
  import {
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-local/lib/metrics/service/MetricsTypes";
  import { navigationEvent } from "../../metrics/initMetrics";
  import { Button } from "../button";
  import ExploreIcon from "../icons/Explore.svelte";
  import Tooltip from "../tooltip/Tooltip.svelte";
  import TooltipContent from "../tooltip/TooltipContent.svelte";

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
        select a model and a timestamp column before exploring metrics
      {:else if buttonStatus === "MISSING_MEASURES_OR_DIMENSIONS"}
        add measures and dimensions before exploring metrics
      {:else}
        explore your metrics dashboard
      {/if}
    </div>
  </TooltipContent>
</Tooltip>
