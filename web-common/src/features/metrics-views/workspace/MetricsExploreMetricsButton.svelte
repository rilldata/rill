<script lang="ts">
  import { goto } from "$app/navigation";
  import {
    Button,
    IconSpaceFixer,
  } from "@rilldata/web-common/components/button";
  import Forward from "@rilldata/web-common/components/icons/Forward.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { navigationEvent } from "@rilldata/web-local/lib/metrics/initMetrics";
  import { BehaviourEventMedium } from "@rilldata/web-local/lib/metrics/service/BehaviourEventTypes";
  import {
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-local/lib/metrics/service/MetricsTypes";

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

  $: if ($metricsInternalRep.getMetricKey("model") === "") {
    buttonDisabled = true;
    buttonStatus = "MISSING_MODEL";
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
    <IconSpaceFixer pullLeft pullRight={false}>
      <Forward />
    </IconSpaceFixer>
    Go to Dashboard
  </Button>
  <TooltipContent slot="tooltip-content">
    <div>
      {#if buttonStatus === "MISSING_MODEL"}
        Select a model before exploring metrics
      {:else if buttonStatus === "MISSING_MEASURES_OR_DIMENSIONS"}
        Add measures and dimensions before exploring metrics
      {:else}
        Explore the metrics dashboard
      {/if}
    </div>
  </TooltipContent>
</Tooltip>
