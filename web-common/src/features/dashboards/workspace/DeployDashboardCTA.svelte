<script lang="ts">
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
  import { ProjectDeployer } from "@rilldata/web-common/features/project/ProjectDeployer";
  import OrgSelectorDialog from "@rilldata/web-common/features/project/OrgSelectorDialog.svelte";
  import { behaviourEvent } from "@rilldata/web-common/metrics/initMetrics";
  import { createLocalServiceDeployValidation } from "@rilldata/web-common/runtime-client/local-service";
  import { get } from "svelte/store";
  import { Button } from "../../../components/button";

  export let type: "primary" | "secondary" = "primary";

  $: deployValidation = createLocalServiceDeployValidation({
    query: {
      refetchOnWindowFocus: true,
    },
  });
  $: isDeployed = !!$deployValidation.data?.deployedProjectId;
  const deployer = new ProjectDeployer();
  const deployerStatus = deployer.getStatus();
  const deploying = deployer.deploying;

  let open = false;
  let orgSelectorOpen = false;
  function onShowDeploy() {
    if (deployer.isDeployed) {
      return deployer.deploy();
    }
    open = true;
    void behaviourEvent?.fireDeployIntentEvent();
  }

  async function onDeploy() {
    if (!(await deployer.validate())) return;
    if (
      $deployValidation.data?.rillUserOrgs?.length &&
      $deployValidation.data?.rillUserOrgs?.length > 1
    ) {
      orgSelectorOpen = true;
      return;
    }

    await deployer.deploy();
    open = false;
  }

  async function onDeployToOrg(org: string) {
    await deployer.deploy(org);
    open = false;
  }

  function handleVisibilityChange() {
    if (document.visibilityState !== "visible" || !get(deploying)) return;
    void deployer.validate();
  }
</script>

<svelte:window on:visibilitychange={handleVisibilityChange} />

<Tooltip distance={8}>
  <Button on:click={onShowDeploy} {type} loading={$deployerStatus.isLoading}>
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
          <Button
            type="primary"
            on:click={onDeploy}
            loading={$deployerStatus.isLoading || $deploying}
          >
            Continue
          </Button>
        </AlertDialogFooter>
      </div>
    </div>
  </AlertDialogContent>
</AlertDialog>

<OrgSelectorDialog
  bind:open={orgSelectorOpen}
  orgs={$deployValidation.data?.rillUserOrgs ?? []}
  onSelect={onDeployToOrg}
/>
