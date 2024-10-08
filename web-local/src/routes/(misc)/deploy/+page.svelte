<script lang="ts">
  import CancelCircleInverse from "@rilldata/web-common/components/icons/CancelCircleInverse.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import OrgSelector from "@rilldata/web-common/features/project/OrgSelector.svelte";
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
  const metadata = deployer.metadata;
  const user = deployer.user;
  const project = deployer.project;
  const deployerStatus = deployer.getStatus();
  const promptOrgSelection = deployer.promptOrgSelection;

  function onOrgSelect(org: string) {
    promptOrgSelection.set(false);
    return deployer.deploy(org);
  }

  function onContactUs() {
    window.Pylon("show");
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
    {#if $promptOrgSelection}
      <OrgSelector
        orgs={$user.data?.rillUserOrgs ?? []}
        onSelect={onOrgSelect}
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
      <CancelCircleInverse size="7rem" className="text-gray-200" />
      <CTAHeader variant="bold">Oops! An error occurred</CTAHeader>
      <CTAMessage>{$deployerStatus.error}</CTAMessage>
      <CTAButton variant="secondary" on:click={onContactUs}>
        Contact us
      </CTAButton>
      <CTANeedHelp />
    {/if}
  </CTAContentContainer>
</CTALayoutContainer>
