<script lang="ts">
  import { goto } from "$app/navigation";
  import {
    Button,
    IconSpaceFixer,
  } from "@rilldata/web-common/components/button";
  import type { LineStatus } from "@rilldata/web-common/components/editor/line-status/state";
  import Forward from "@rilldata/web-common/components/icons/Forward.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { behaviourEvent } from "@rilldata/web-common/metrics/initMetrics";
  import { BehaviourEventMedium } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
  import {
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-common/metrics/service/MetricsTypes";
  import { getModelOutOfPossiblyMalformedYAML } from "../utils";

  export let yaml;
  export let metricsDefName;
  export let error: LineStatus;

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

  const TOOLTIP_CTA = "Fix this error to enable your dashboard.";
  $: possibleModel = getModelOutOfPossiblyMalformedYAML(yaml);
  $: if (!yaml?.length) {
    buttonDisabled = true;
    buttonStatus = ["WHAT.", TOOLTIP_CTA];
  } else if (error) {
    buttonDisabled = true;
    buttonStatus = [error.message, TOOLTIP_CTA];
  } else if (possibleModel === null) {
    buttonDisabled = true;
    buttonStatus = ["Select a model before exploring metrics"];
  } else {
    buttonStatus = ["Explore your metrics dashboard"];
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
    <IconSpaceFixer pullLeft>
      <Forward /></IconSpaceFixer
    > Go to Dashboard
  </Button>
  <TooltipContent slot="tooltip-content">
    {#each buttonStatus as status}
      <div>{status}</div>
    {/each}
  </TooltipContent>
</Tooltip>
