<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import IconSpaceFixer from "@rilldata/web-common/components/button/IconSpaceFixer.svelte";
  import ResponsiveButtonText from "@rilldata/web-common/components/panel/ResponsiveButtonText.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { BehaviourEventMedium } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
  import { MetricsEventSpace } from "@rilldata/web-common/metrics/service/MetricsTypes";
  import { Wand } from "lucide-svelte";
  import { V1ReconcileStatus } from "../../../runtime-client";
  import httpClient from "@rilldata/web-common/runtime-client/http-client";
  import { allowPrimary } from "../../dashboards/workspace/DeployProjectCTA.svelte";
  import { featureFlags } from "../../feature-flags";
  import { useCreateMetricsViewFromTableUIAction } from "../../metrics-views/ai-generation/generateMetricsView";
  import { useModel } from "../selectors";

  export let modelName: string;
  export let hasError = false;
  export let collapse = false;

  const { ai } = featureFlags;

  const instanceId = httpClient.getInstanceId();

  $: modelQuery = useModel(instanceId, modelName);
  $: connector = $modelQuery.data?.model?.spec?.outputConnector;
  $: modelIsIdle =
    $modelQuery.data?.meta?.reconcileStatus ===
    V1ReconcileStatus.RECONCILE_STATUS_IDLE;

  $: createMetricsViewFromModel = useCreateMetricsViewFromTableUIAction(
    instanceId,
    connector as string,
    "",
    "",
    modelName,
    false,
    BehaviourEventMedium.Button,
    MetricsEventSpace.RightPanel,
  );
</script>

<Tooltip distance={8} location="bottom">
  <Button
    disabled={!modelIsIdle || hasError}
    onClick={createMetricsViewFromModel}
    type={$allowPrimary ? "primary" : "secondary"}
  >
    <IconSpaceFixer pullLeft pullRight={collapse}>
      <Wand size="14px" />
    </IconSpaceFixer>
    <ResponsiveButtonText {collapse}>
      Generate metrics view
      {#if $ai}
        with AI
      {/if}
    </ResponsiveButtonText>
  </Button>
  <TooltipContent slot="tooltip-content">
    {#if hasError}
      Fix the errors in your model
    {:else if !modelIsIdle}
      Model is not ready
    {:else}
      Generate metrics from this model
    {/if}
  </TooltipContent>
</Tooltip>
