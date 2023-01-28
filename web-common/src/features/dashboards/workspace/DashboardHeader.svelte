<script lang="ts">
  import { goto } from "$app/navigation";
  import {
    Button,
    IconSpaceFixer,
  } from "@rilldata/web-common/components/button";
  import Forward from "@rilldata/web-common/components/icons/Forward.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import {
    appQueryStatusStore,
    runtimeStore,
  } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { WorkspaceHeader } from "@rilldata/web-local/lib/components/workspace";
  import { navigationEvent } from "@rilldata/web-local/lib/metrics/initMetrics";
  import { BehaviourEventMedium } from "@rilldata/web-local/lib/metrics/service/BehaviourEventTypes";
  import {
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-local/lib/metrics/service/MetricsTypes";
  import { getContext } from "svelte";
  import type { Tweened } from "svelte/motion";
  import Filters from "../filters/Filters.svelte";
  import { useMetaQuery } from "../selectors";
  import TimeControls from "../time-controls/TimeControls.svelte";

  export let metricViewName: string;

  const navigationVisibilityTween = getContext(
    "rill:app:navigation-visibility-tween"
  ) as Tweened<number>;

  const viewMetrics = (metricViewName: string) => {
    goto(`/dashboard/${metricViewName}/edit`);

    navigationEvent.fireEvent(
      metricViewName,
      BehaviourEventMedium.Button,
      MetricsEventSpace.Workspace,
      MetricsEventScreenName.Dashboard,
      MetricsEventScreenName.MetricsDefinition
    );
  };

  $: metaQuery = useMetaQuery($runtimeStore.instanceId, metricViewName);
  $: displayName = $metaQuery.data?.label;
  $: isEditableDashboard = $runtimeStore.readOnly === false;

  appQueryStatusStore;
</script>

<section class="w-full flex flex-col" id="header">
  <!-- top row
    title and call to action
  -->
  <WorkspaceHeader
    titleInput={displayName || metricViewName}
    editable={false}
    appRunning={$appQueryStatusStore}
  >
    <div slot="cta">
      {#if isEditableDashboard}
        <div style="flex-shrink: 0;" class="flex gap-x-2">
          <Tooltip distance={8}>
            <Button
              on:click={() => viewMetrics(metricViewName)}
              type="secondary"
            >
              Edit Metrics
            </Button>
            <TooltipContent slot="tooltip-content">
              Edit this dashboard's metrics & settings
            </TooltipContent>
          </Tooltip>
          <Tooltip distance={8}>
            <Button
              on:click={() => {
                goto(`/model/${$metaQuery?.data?.model}`);
              }}
              type="primary"
            >
              <IconSpaceFixer pullLeft pullRight={false}>
                <Forward size="14px" />
              </IconSpaceFixer>
              Edit Model
            </Button>
            <TooltipContent slot="tooltip-content">
              Edit the model that powers this dashboard
            </TooltipContent>
          </Tooltip>
        </div>
      {/if}
    </div>
  </WorkspaceHeader>

  <!-- bottom row -->
  <div class="px-2 pt-3">
    <TimeControls {metricViewName} />
    {#key metricViewName}
      <Filters {metricViewName} />
    {/key}
  </div>
</section>
