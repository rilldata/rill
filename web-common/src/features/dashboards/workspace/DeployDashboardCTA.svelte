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
  import { createLocalServiceDeployValidation } from "@rilldata/web-common/runtime-client/local-service";
  import { Button } from "../../../components/button";

  export let type: "primary" | "secondary" = "primary";

  $: deployValidation = createLocalServiceDeployValidation({
    query: {
      refetchOnWindowFocus: true,
    },
  });
  $: isDeployed = !!$deployValidation.data?.deployedProjectId;

  let open = false;
  function onShowDeploy() {
    if (isDeployed) {
      onDeploy();
    } else {
      open = true;
    }
    void behaviourEvent?.fireDeployIntentEvent();
  }

  function onDeploy() {
    open = false;
    window.open(`${$page.url.protocol}//${$page.url.host}/deploy`);
  }
</script>

<Tooltip distance={8}>
  <Button on:click={onShowDeploy} {type} loading={$deployValidation.isFetching}>
    {isDeployed ? "Redeploy" : "Deploy"}
  </Button>
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
          <AlertDialogTitle>Deploy this project for free</AlertDialogTitle>
          <AlertDialogDescription>
            You’re about to start a 30-day FREE trial of Rill Cloud, where you
            can set alerts, share dashboards, and more.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter class="mt-5">
          <Button type="secondary" on:click={() => (open = false)}>Back</Button>
          <Button type="primary" on:click={onDeploy}>Continue</Button>
        </AlertDialogFooter>
      </div>
    </div>
  </AlertDialogContent>
</AlertDialog>
