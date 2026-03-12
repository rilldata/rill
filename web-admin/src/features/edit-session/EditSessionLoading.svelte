<script lang="ts">
  import { V1DeploymentStatus } from "@rilldata/web-admin/client";
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaHeader from "@rilldata/web-common/components/calls-to-action/CTAHeader.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import CtaMessage from "@rilldata/web-common/components/calls-to-action/CTAMessage.svelte";
  import LoadingSpinner from "@rilldata/web-common/components/LoadingSpinner.svelte";

  export let status: V1DeploymentStatus | undefined;
  export let statusMessage: string | undefined;

  $: message = getStatusMessage(status);

  function getStatusMessage(s: V1DeploymentStatus | undefined): string {
    switch (s) {
      case V1DeploymentStatus.DEPLOYMENT_STATUS_PENDING:
        return "Provisioning your editing environment...";
      case V1DeploymentStatus.DEPLOYMENT_STATUS_UPDATING:
        return "Updating your editing environment...";
      default:
        return "Starting your editing environment...";
    }
  }
</script>

<CtaLayoutContainer>
  <CtaContentContainer>
    <LoadingSpinner />
    <CtaHeader variant="bold">
      {message}
    </CtaHeader>
    {#if statusMessage}
      <CtaMessage>{statusMessage}</CtaMessage>
    {/if}
  </CtaContentContainer>
</CtaLayoutContainer>
