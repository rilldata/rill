<script lang="ts">
  import { page } from "$app/stores";
  import {
    AlertDialog,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
    AlertDialogTrigger,
  } from "@rilldata/web-common/components/alert-dialog";
  import DeployIcon from "@rilldata/web-common/components/icons/DeployIcon.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { behaviourEvent } from "@rilldata/web-common/metrics/initMetrics";
  import { BehaviourEventAction } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
  import { createLocalServiceDeployValidation } from "@rilldata/web-common/runtime-client/local-service";
  import { Button } from "../../../components/button";

  $: deployValidation = createLocalServiceDeployValidation({
    query: {
      refetchOnWindowFocus: true,
    },
  });
  $: isDeployed = !!$deployValidation.data?.deployedProjectId;

  $: deployPageUrl = `${$page.url.protocol}//${$page.url.host}/deploy`;

  let open = false;
  function onShowDeploy() {
    if (!isDeployed) {
      open = true;
    }
    void behaviourEvent?.fireDeployEvent(BehaviourEventAction.DeployIntent);
  }
</script>

<Tooltip distance={8}>
  {#if isDeployed}
    <Button
      loading={$deployValidation.isLoading}
      on:click={onShowDeploy}
      type="primary"
      href={deployPageUrl}
      target="_blank"
    >
      Redeploy
    </Button>
  {:else}
    <Button
      loading={$deployValidation.isLoading}
      on:click={onShowDeploy}
      type="primary"
    >
      Deploy to share
    </Button>
  {/if}
  <TooltipContent slot="tooltip-content">
    Deploy this dashboard to Rill Cloud
  </TooltipContent>
</Tooltip>

<AlertDialog bind:open>
  <AlertDialogTrigger asChild>
    <div class="hidden"></div>
  </AlertDialogTrigger>
  <AlertDialogContent>
    <div class="flex flex-row">
      <DeployIcon size="150px" />
      <div class="flex flex-col">
        <AlertDialogHeader>
          <AlertDialogTitle>Deploy this project</AlertDialogTitle>
          <AlertDialogDescription>
            You’re about to deploy to Rill Cloud, where you can set alerts,
            share dashboards, and more. <a
              href="https://www.rilldata.com/pricing"
              target="_blank">See pricing details</a
            >
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter class="mt-5">
          <Button on:click={() => (open = false)} type="secondary">Back</Button>
          <Button
            on:click={() => (open = false)}
            type="primary"
            href={deployPageUrl}
            target="_blank">Continue</Button
          >
        </AlertDialogFooter>
      </div>
    </div>
  </AlertDialogContent>
</AlertDialog>
