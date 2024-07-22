<script lang="ts">
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import OrgSelectorDialog from "@rilldata/web-common/features/project/OrgSelectorDialog.svelte";
  import { ProjectDeployer } from "@rilldata/web-common/features/project/ProjectDeployer";
  import { onMount } from "svelte";
  import CTAContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CTALayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import CTAHeader from "@rilldata/web-common/components/calls-to-action/CTAHeader.svelte";
  import CTANeedHelp from "@rilldata/web-common/components/calls-to-action/CTANeedHelp.svelte";
  import CTAMessage from "@rilldata/web-common/components/calls-to-action/CTAMessage.svelte";
  import CTAButton from "@rilldata/web-common/components/calls-to-action/CTAButton.svelte";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";

  const deployer = new ProjectDeployer();
  const deployValidation = deployer.validation;
  const deployerStatus = deployer.getStatus();
  const promptOrgSelection = deployer.promptOrgSelection;

  function handleVisibilityChange() {
    if (document.visibilityState !== "visible") return;
    void deployer.checkDeployStatus();
  }

  function onContactUs() {
    // TODO
  }

  onMount(() => {
    void deployer.deploy();
  });
</script>

<svelte:window on:visibilitychange={handleVisibilityChange} />

<CTAContentContainer>
  <CTALayoutContainer>
    {#if $deployerStatus.isLoading}
      <div class="h-36">
        <Spinner status={EntityStatus.Running} size="7rem" duration={725} />
      </div>
      <CTAHeader variant="bold">
        Hang tight! We're deploying your project...
      </CTAHeader>
    {:else if $deployerStatus.error}
      <div class="h-36">
        <Spinner status={EntityStatus.Running} size="7rem" duration={725} />
      </div>
      <CTAHeader variant="bold">Oops! An error occurred</CTAHeader>
      <CTAMessage>{$deployerStatus.error}</CTAMessage>
      <CTAButton variant="secondary" on:click={onContactUs}>
        Contact us
      </CTAButton>
    {/if}
    <CTANeedHelp />
  </CTALayoutContainer>
</CTAContentContainer>

<OrgSelectorDialog
  bind:open={$promptOrgSelection}
  orgs={$deployValidation.data?.rillUserOrgs ?? []}
  onSelect={(org) => deployer.deploy(org)}
/>
