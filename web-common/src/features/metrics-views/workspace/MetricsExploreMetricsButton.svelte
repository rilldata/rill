<script lang="ts">
  import { goto } from "$app/navigation";
  import { Button } from "@rilldata/web-common/components/button";
  import ExploreIcon from "@rilldata/web-common/components/icons/Explore.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { behaviourEvent } from "@rilldata/web-common/metrics/initMetrics";
  import { BehaviourEventMedium } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
  import {
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-common/metrics/service/MetricsTypes";
  import { getContext } from "svelte";
  import type { Writable } from "svelte/store";

  export let metricsInternalRep;
  export let metricsDefName;

  let metricsConfigErrorStore = getContext(
    "rill:metrics-config:errors"
  ) as Writable<any>;

  $: measures = $metricsInternalRep.getMeasures();
  $: dimensions = $metricsInternalRep.getDimensions();

  let buttonDisabled = true;
  let buttonStatus;

  const viewDashboard = () => {
    goto(`/dashboard/${metricsDefName}`);

    behaviourEvent.fireNavigationEvent(
      metricsDefName,
      BehaviourEventMedium.Button,
      MetricsEventSpace.Workspace,
      MetricsEventScreenName.MetricsDefinition,
      MetricsEventScreenName.Dashboard
    );
  };

  $: if ($metricsInternalRep.getMetricKey("model") === "") {
    buttonDisabled = true;
    buttonStatus = ["Select a model before exploring metrics"];
  } else if (
    // check if all the measures have a valid expression
    measures?.filter((measure) => measure?.expression?.length)?.length === 0 ||
    // and if the dimensions all have a valid property
    dimensions?.filter((dimension) => dimension?.column?.length)?.length === 0
  ) {
    buttonDisabled = true;
    buttonStatus = ["Add measures and dimensions before exploring metrics"];
  } else if (Object.values($metricsConfigErrorStore).some((error) => error)) {
    buttonDisabled = true;
    buttonStatus = Object.values($metricsConfigErrorStore).filter(
      (error) => error
    );
  } else {
    buttonStatus = ["Explore the metrics dashboard"];
    buttonDisabled = false;
  }
</script>

<Tooltip alignment="middle" distance={5} location="right">
  <!-- TODO: we need to standardize these buttons. -->
  <Button
    disabled={buttonDisabled}
    label="Go to dashboard"
    on:click={() => viewDashboard()}
    type="primary"
  >
    Go to Dashboard <ExploreIcon size="16px" />
  </Button>
  <TooltipContent slot="tooltip-content">
    {#each buttonStatus as status}
      <div>{status}</div>
    {/each}
  </TooltipContent>
</Tooltip>
