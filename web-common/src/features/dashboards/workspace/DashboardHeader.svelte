<script lang="ts">
  import { goto } from "$app/navigation";
  import { Button } from "@rilldata/web-common/components/button";
  import MetricsIcon from "@rilldata/web-common/components/icons/Metrics.svelte";
  import PanelCTA from "@rilldata/web-common/components/panel/PanelCTA.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { calendlyModalStore } from "@rilldata/web-common/features/dashboards/dashboard-stores";
  import { featureFlags } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { behaviourEvent } from "@rilldata/web-local/lib/metrics/initMetrics";
  import { BehaviourEventMedium } from "@rilldata/web-local/lib/metrics/service/BehaviourEventTypes";
  import {
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-local/lib/metrics/service/MetricsTypes";
  import { getContext } from "svelte";
  import type { Tweened } from "svelte/motion";
  import { runtime } from "../../../runtime-client/runtime-store";
  import Filters from "../filters/Filters.svelte";
  import { useMetaQuery } from "../selectors";
  import TimeControls from "../time-controls/TimeControls.svelte";

  export let metricViewName: string;

  const navigationVisibilityTween = getContext(
    "rill:app:navigation-visibility-tween"
  ) as Tweened<number>;

  const viewMetrics = (metricViewName: string) => {
    goto(`/dashboard/${metricViewName}/edit`);

    behaviourEvent.fireNavigationEvent(
      metricViewName,
      BehaviourEventMedium.Button,
      MetricsEventSpace.Workspace,
      MetricsEventScreenName.Dashboard,
      MetricsEventScreenName.MetricsDefinition
    );
  };

  $: metaQuery = useMetaQuery($runtime.instanceId, metricViewName);
  $: displayName = $metaQuery.data?.label;
  $: isEditableDashboard = $featureFlags.readOnly === false;

  function openCalendly() {
    calendlyModalStore.set(metricViewName);
    behaviourEvent.firePublishEvent(
      metricViewName,
      BehaviourEventMedium.Button,
      MetricsEventSpace.Workspace,
      MetricsEventScreenName.Dashboard,
      MetricsEventScreenName.Dashboard,
      true
    );
  }
</script>

<section
  class="w-full flex flex-col"
  id="header"
  style:padding-left="{$navigationVisibilityTween * 24}px"
>
  <!-- top row
    title and call to action
  -->
  <div
    class="flex items-center justify-between w-full pl-1 pr-4"
    style:height="var(--header-height)"
  >
    <!-- title element -->
    <h1 style:line-height="1.1" style:margin-top="-1px">
      <div class="pl-4" style:font-family="InterDisplay" style:font-size="20px">
        {displayName || metricViewName}
      </div>
    </h1>
    <!-- top right CTAs -->
    {#if isEditableDashboard}
      <PanelCTA side="right">
        <Tooltip distance={8}>
          <Button on:click={() => viewMetrics(metricViewName)} type="secondary">
            Edit Metrics <MetricsIcon size="16px" />
          </Button>
          <TooltipContent slot="tooltip-content">
            Edit this dashboard's metrics & settings
          </TooltipContent>
        </Tooltip>
        <Tooltip distance={8}>
          <Button on:click={openCalendly} type="primary">Publish</Button>
          <TooltipContent slot="tooltip-content">
            Schedule time to chat with Rill about early access to hosted
            dashboards.
          </TooltipContent>
        </Tooltip>
      </PanelCTA>
    {/if}
  </div>
  <!-- bottom row -->
  <div class="px-2 pt-1">
    <TimeControls {metricViewName} />
    {#key metricViewName}
      <Filters {metricViewName} />
    {/key}
  </div>
</section>
