<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import IconSpaceFixer from "@rilldata/web-common/components/button/IconSpaceFixer.svelte";
  import Add from "@rilldata/web-common/components/icons/Add.svelte";
  import ResponsiveButtonText from "@rilldata/web-common/components/panel/ResponsiveButtonText.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { BehaviourEventMedium } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
  import { MetricsEventSpace } from "@rilldata/web-common/metrics/service/MetricsTypes";
  import { V1ReconcileStatus } from "../../../runtime-client";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { useCreateDashboardFromTableUIAction } from "../../metrics-views/ai-generation/generateMetricsView";
  import { useModel } from "../selectors";
  import { featureFlags } from "../../feature-flags";
  import { WandIcon } from "lucide-svelte";

  export let modelName: string;
  export let hasError = false;
  export let collapse = false;

  const { ai } = featureFlags;

  $: modelQuery = useModel($runtime.instanceId, modelName);
  $: connector = $modelQuery.data?.model?.spec?.outputConnector;
  $: modelIsIdle =
    $modelQuery.data?.meta?.reconcileStatus ===
    V1ReconcileStatus.RECONCILE_STATUS_IDLE;

  $: createDashboardFromModel = useCreateDashboardFromTableUIAction(
    $runtime.instanceId,
    connector as string,
    "",
    "",
    modelName,
    "dashboards",
    BehaviourEventMedium.Button,
    MetricsEventSpace.RightPanel,
  );
</script>

<Tooltip alignment="right" distance={8} location="bottom">
  <Button
    disabled={!modelIsIdle || hasError}
    on:click={createDashboardFromModel}
    type="brand"
  >
    <IconSpaceFixer pullLeft pullRight={collapse}>
      <Add />
    </IconSpaceFixer>
    <ResponsiveButtonText {collapse}>
      Generate dashboard
      {#if $ai}
        with AI
        <WandIcon class="w-3 h-3" />
      {/if}
    </ResponsiveButtonText>
  </Button>
  <TooltipContent slot="tooltip-content">
    {#if hasError}
      Fix the errors in your model to autogenerate dashboard
    {:else if !modelIsIdle}
      Model is not ready to generate a dashboard
    {:else}
      Generate a dashboard from this model
    {/if}
  </TooltipContent>
</Tooltip>
