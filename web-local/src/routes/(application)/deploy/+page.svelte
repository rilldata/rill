<script lang="ts">
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import OrgSelectorDialog from "@rilldata/web-common/features/project/OrgSelectorDialog.svelte";
  import { ProjectDeployer } from "@rilldata/web-common/features/project/ProjectDeployer";
  import { waitUntil } from "@rilldata/web-common/lib/waitUtils";
  import { createLocalServiceDeployValidation } from "@rilldata/web-common/runtime-client/local-service";
  import { onMount } from "svelte";
  import { get } from "svelte/store";
  import CTAContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CTALayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import CTAHeader from "@rilldata/web-common/components/calls-to-action/CTAHeader.svelte";
  import CTANeedHelp from "@rilldata/web-common/components/calls-to-action/CTANeedHelp.svelte";
  import CTAMessage from "@rilldata/web-common/components/calls-to-action/CTAMessage.svelte";
  import CTAButton from "@rilldata/web-common/components/calls-to-action/CTAButton.svelte";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";

  $: deployValidation = createLocalServiceDeployValidation({
    query: {
      refetchOnWindowFocus: true,
    },
  });
  const deployer = new ProjectDeployer();
  const deployerStatus = deployer.getStatus();

  let orgSelectorOpen = false;

  async function onDeploy() {
    await waitUntil(() => !get(deployValidation).isLoading);
    if (!(await deployer.validate())) return;
    if (
      $deployValidation.data?.rillUserOrgs?.length &&
      $deployValidation.data?.rillUserOrgs?.length > 1
    ) {
      orgSelectorOpen = true;
      return;
    }

    await deployer.deploy();
  }

  function handleVisibilityChange() {
    if (document.visibilityState !== "visible" || !get(deployer.validating))
      return;
    // wait for deployValidation's refetchOnWindowFocus to trigger
    setTimeout(() => onDeploy(), 0);
  }

  function onContactUs() {
    // TODO
  }

  onMount(() => {
    void onDeploy();
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
  bind:open={orgSelectorOpen}
  orgs={$deployValidation.data?.rillUserOrgs ?? []}
  onSelect={(org) => deployer.deploy(org)}
/>
