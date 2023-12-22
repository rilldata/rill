<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import IconSpaceFixer from "@rilldata/web-common/components/button/IconSpaceFixer.svelte";
  import Add from "@rilldata/web-common/components/icons/Add.svelte";
  import ResponsiveButtonText from "@rilldata/web-common/components/panel/ResponsiveButtonText.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import {
    useCreateDashboardFromModelUIAction,
    useModelSchemaIsReady,
  } from "@rilldata/web-common/features/models/createDashboardFromModel";
  import { BehaviourEventMedium } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
  import { MetricsEventSpace } from "@rilldata/web-common/metrics/service/MetricsTypes";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { runtime } from "../../../runtime-client/runtime-store";

  export let modelName: string;
  export let hasError = false;
  export let collapse = false;

  const queryClient = useQueryClient();

  $: modelSchemaIsReady = useModelSchemaIsReady(
    queryClient,
    $runtime.instanceId,
    modelName
  );

  $: createDashboardFromModel = useCreateDashboardFromModelUIAction(
    $runtime.instanceId,
    modelName,
    queryClient,
    BehaviourEventMedium.Button,
    MetricsEventSpace.RightPanel
  );
</script>

<Tooltip alignment="right" distance={8} location="bottom">
  <Button
    disabled={!$modelSchemaIsReady}
    on:click={createDashboardFromModel}
    type="primary"
  >
    <IconSpaceFixer pullLeft pullRight={collapse}>
      <Add />
    </IconSpaceFixer>
    <ResponsiveButtonText {collapse}>Autogenerate Dashboard</ResponsiveButtonText>
  </Button>
  <TooltipContent slot="tooltip-content">
    {#if hasError}
      Fix the errors in your model to autogenerate dashboard
    {:else}
      Autogenerate a dashboard from this model
    {/if}
  </TooltipContent>
</Tooltip>
