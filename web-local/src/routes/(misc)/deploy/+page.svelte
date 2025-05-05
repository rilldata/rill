<script lang="ts">
  import { goto } from "$app/navigation";
  import { Button } from "@rilldata/web-common/components/button";
  import TrialDetailsDialog from "@rilldata/web-common/features/billing/TrialDetailsDialog.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { getPlanUpgradeUrl } from "@rilldata/web-common/features/organization/utils";
  import OrgSelector from "@rilldata/web-common/features/project/OrgSelector.svelte";
  import {
    ProjectDeployer,
    ProjectDeployStage,
  } from "@rilldata/web-common/features/project/ProjectDeployer";
  import DeployError from "@rilldata/web-common/features/project/DeployError.svelte";
  import { onMount } from "svelte";
  import CTAContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CTALayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import CTAHeader from "@rilldata/web-common/components/calls-to-action/CTAHeader.svelte";
  import CTANeedHelp from "@rilldata/web-common/components/calls-to-action/CTANeedHelp.svelte";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import type { PageData } from "./$types";
  import CreateNewOrgForm from "@rilldata/web-common/features/organization/CreateNewOrgForm.svelte";
  import { CreateNewOrgFormId } from "@rilldata/web-common/features/organization/CreateNewOrgForm.svelte";

  export let data: PageData;

  const deployer = new ProjectDeployer(data.orgParam);
  const metadata = deployer.metadata;
  const user = deployer.user;
  const project = deployer.project;
  const deployerStatus = deployer.getStatus();
  const stage = deployer.stage;
  const org = deployer.org;

  $: planUpgradeUrl = getPlanUpgradeUrl(org);

  // This is specifically the org selected using the OrgSelector.
  // Used to retrigger the deploy after the user confirms deploy on an empty org.
  let deployConfirmOpen = false;

  function onBack() {
    if ($user.data?.rillUserOrgs?.length) {
      deployer.onSelectOrg();
    } else {
      void goto("/");
    }
  }

  onMount(() => {
    void deployer.loginOrDeploy();
  });
</script>

<!-- This seems to be necessary to trigger tanstack query to update the query object -->
<!-- TODO: find a config to avoid this -->
<div class="hidden">
  {$user.isLoading}-{$metadata.isLoading}-{$project.isLoading}
</div>

<CTALayoutContainer>
  <CTAContentContainer>
    {#if $stage === ProjectDeployStage.CreateNewOrg}
      <div class="text-xl">Let’s create your first organization</div>
      <div class="text-base text-gray-500">
        Create an organization to deploy this project to. <a
          href="https://docs.rilldata.com/reference/cli/org/create"
          target="_blank">See docs</a
        >
      </div>
      <CreateNewOrgForm onCreate={(org) => deployer.setOrgAndName(org)} />
      <Button
        wide
        forcedStyle="min-width:500px !important;"
        type="primary"
        submitForm
        form={CreateNewOrgFormId}
      >
        Continue
      </Button>
    {:else if $stage === ProjectDeployStage.SelectOrg}
      <OrgSelector
        orgs={$user.data?.rillUserOrgs ?? []}
        onSelect={(org) => deployer.setOrgAndName(org)}
      />
    {:else if $deployerStatus.isLoading}
      <div class="h-36">
        <Spinner status={EntityStatus.Running} size="7rem" duration={725} />
      </div>
      <CTAHeader variant="bold">
        Hang tight! We're deploying your project...
      </CTAHeader>
      <CTANeedHelp />
    {:else if $deployerStatus.error}
      <DeployError
        error={$deployerStatus.error}
        planUpgradeUrl={$planUpgradeUrl}
        onRetry={() => deployer.deploy()}
        {onBack}
      />
    {/if}
  </CTAContentContainer>
</CTALayoutContainer>

<TrialDetailsDialog
  bind:open={deployConfirmOpen}
  onContinue={() => deployer.deploy()}
/>
