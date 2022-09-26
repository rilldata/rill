<script lang="ts">
  import { goto } from "$app/navigation";
  import { RootConfig } from "$web-local/common/config/RootConfig";
  import { BehaviourEventMedium } from "$web-local/common/metrics-service/BehaviourEventTypes";
  import {
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "$web-local/common/metrics-service/MetricsTypes";
  import { Button } from "../button";
  import ExploreIcon from "../icons/Explore.svelte";
  import Tooltip from "../tooltip/Tooltip.svelte";
  import TooltipContent from "../tooltip/TooltipContent.svelte";
  import { navigationEvent } from "../../metrics/initMetrics";
  import { getMetricsDefReadableById } from "../../redux-store/metrics-definition/metrics-definition-readables";
  import { useMetaQuery } from "../../svelte-query/queries/metrics-views/metadata";
  import { getContext } from "svelte";

  export let metricsDefId: string;

  const config = getContext<RootConfig>("config");

  // query the `/meta` endpoint to get the valid measures and dimensions
  $: metaQuery = useMetaQuery(config, metricsDefId);
  $: measures = $metaQuery.data?.measures;
  $: dimensions = $metaQuery.data?.dimensions;

  $: selectedMetricsDef = getMetricsDefReadableById(metricsDefId);

  let buttonDisabled = true;
  let buttonStatus = "OK";

  const viewDashboard = (metricsDefId: string) => {
    goto(`/dashboard/${metricsDefId}`);

    navigationEvent.fireEvent(
      metricsDefId,
      BehaviourEventMedium.Button,
      MetricsEventSpace.Workspace,
      MetricsEventScreenName.MetricsDefinition,
      MetricsEventScreenName.Dashboard
    );
  };

  $: if (
    $selectedMetricsDef?.sourceModelId === undefined ||
    $selectedMetricsDef?.timeDimension === undefined
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
    on:click={() => viewDashboard(metricsDefId)}
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
