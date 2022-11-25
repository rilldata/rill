<script lang="ts">
  import { goto } from "$app/navigation";
  import { RootConfig } from "@rilldata/web-local/common/config/RootConfig";
  import { BehaviourEventMedium } from "@rilldata/web-local/common/metrics-service/BehaviourEventTypes";
  import {
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-local/common/metrics-service/MetricsTypes";
  import { Button } from "../button";
  import ExploreIcon from "../icons/Explore.svelte";
  import Tooltip from "../tooltip/Tooltip.svelte";
  import TooltipContent from "../tooltip/TooltipContent.svelte";
  import { navigationEvent } from "../../metrics/initMetrics";
  import { getMetricsDefReadableById } from "../../redux-store/metrics-definition/metrics-definition-readables";
  import { useMetaQuery } from "../../svelte-query/queries/metrics-views/metadata";
  import { getContext } from "svelte";

  export let metricsInternalRep;
  export let metricsDefName;

  const config = getContext<RootConfig>("config");

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
    $metricsInternalRep.getMetricKey("from") === "" ||
    $metricsInternalRep.getMetricKey("timeseries") === ""
  ) {
    buttonDisabled = true;
    buttonStatus = "MISSING_MODEL_OR_TIMESTAMP";
  } else if (measures?.length === 0 || dimensions?.length === 0) {
    buttonDisabled = true;
    buttonStatus = "MISSING_MEASURES_OR_DIMENSIONS";
  } else {
    buttonDisabled = false;
  }
</script>

<Tooltip location="right" alignment="middle" distance={5}>
  <!-- TODO: we need to standardize these buttons. -->
  <Button
    type="primary"
    disabled={buttonDisabled}
    on:click={() => viewDashboard()}
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
