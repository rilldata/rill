<script lang="ts">
  import MetricsIcon from "@rilldata/web-common/components/icons/Metrics.svelte";
  import { useDashboard } from "@rilldata/web-common/features/dashboards/selectors";
  import { V1ReconcileStatus } from "@rilldata/web-common/runtime-client";
  import { Button } from "../../../components/button";
  import Tooltip from "../../../components/tooltip/Tooltip.svelte";
  import TooltipContent from "../../../components/tooltip/TooltipContent.svelte";
  import { behaviourEvent } from "../../../metrics/initMetrics";
  import { BehaviourEventMedium } from "../../../metrics/service/BehaviourEventTypes";
  import {
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "../../../metrics/service/MetricsTypes";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { featureFlags } from "../../feature-flags";
  import ViewAsButton from "../granular-access-policies/ViewAsButton.svelte";
  import { useDashboardPolicyCheck } from "../granular-access-policies/useDashboardPolicyCheck";
  import DeployDashboardCta from "./DeployDashboardCTA.svelte";

  export let metricViewName: string;

  $: dashboardQuery = useDashboard($runtime.instanceId, metricViewName);
  $: filePath = $dashboardQuery.data?.meta?.filePaths?.[0] ?? "";

  $: dashboardPolicyCheck = useDashboardPolicyCheck(
    $runtime.instanceId,
    filePath,
  );

  const { readOnly } = featureFlags;

  $: dashboardIsIdle =
    $dashboardQuery.data?.meta?.reconcileStatus ===
    V1ReconcileStatus.RECONCILE_STATUS_IDLE;

  function fireTelemetry() {
    behaviourEvent
      .fireNavigationEvent(
        metricViewName,
        BehaviourEventMedium.Button,
        MetricsEventSpace.Workspace,
        MetricsEventScreenName.Dashboard,
        MetricsEventScreenName.MetricsDefinition,
      )
      .catch(console.error);
  }

  let showDeployDashboardModal = false;

  async function showDeployModal() {
    showDeployDashboardModal = true;
    await behaviourEvent?.fireDeployIntentEvent();
  }
</script>

<div class="flex gap-2 flex-shrink-0 ml-auto">
  {#if $dashboardPolicyCheck.data}
    <ViewAsButton />
  {/if}
  {#if !$readOnly}
    <Tooltip distance={8}>
      <Button
        href={`/files${filePath}`}
        disabled={!dashboardIsIdle}
        on:click={fireTelemetry}
        type="secondary"
      >
        Edit Metrics <MetricsIcon size="16px" />
      </Button>
      <TooltipContent slot="tooltip-content">
        {#if !dashboardIsIdle}
          Dependencies are being ingested
        {:else}
          Edit this dashboard's metrics & settings
        {/if}
      </TooltipContent>
    </Tooltip>
    <Tooltip distance={8}>
      <Button on:click={() => showDeployModal()} type="brand">Deploy</Button>
      <TooltipContent slot="tooltip-content">
        Deploy this dashboard to Rill Cloud
      </TooltipContent>
    </Tooltip>
  {/if}
</div>

<DeployDashboardCta
  on:close={() => (showDeployDashboardModal = false)}
  open={showDeployDashboardModal}
/>
