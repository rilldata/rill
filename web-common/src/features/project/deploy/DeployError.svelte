<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import CTAButton from "@rilldata/web-common/components/calls-to-action/CTAButton.svelte";
  import CTAHeader from "@rilldata/web-common/components/calls-to-action/CTAHeader.svelte";
  import CTAMessage from "@rilldata/web-common/components/calls-to-action/CTAMessage.svelte";
  import CTAPylonHelp from "@rilldata/web-common/components/calls-to-action/CTAPylonHelp.svelte";
  import CancelCircleInverse from "@rilldata/web-common/components/icons/CancelCircleInverse.svelte";
  import PricingDetails from "@rilldata/web-common/features/billing/PricingDetails.svelte";
  import {
    DeployErrorType,
    getPrettyDeployError,
  } from "@rilldata/web-common/features/project/deploy/deploy-errors";

  export let error: Error;
  export let isOrgOnTrial: boolean;
  export let planUpgradeUrl: string;
  export let githubAccessUrl: string = "";
  export let onRetry: () => void;
  export let onBack: () => void;

  $: deployError = getPrettyDeployError(error, isOrgOnTrial);

  $: isQuotaError =
    deployError.type === DeployErrorType.ProjectLimitHit ||
    deployError.type === DeployErrorType.OrgLimitHit ||
    deployError.type === DeployErrorType.TrialEnded ||
    deployError.type === DeployErrorType.SubscriptionEnded;
  $: isGithubNoAccessError =
    deployError.type === DeployErrorType.GithubNoAccess && !!githubAccessUrl;
</script>

{#if isQuotaError}
  <CTAHeader variant="bold">{deployError.title}</CTAHeader>
  <p class="text-base text-gray-500 text-left w-[500px]">
    <PricingDetails extraText={deployError.message} />
  </p>
  <Button type="primary" href={planUpgradeUrl} wide>Upgrade</Button>
  <Button type="secondary" noStroke wide onClick={onBack}>Back</Button>
{:else if isGithubNoAccessError}
  <CancelCircleInverse size="7rem" className="text-gray-200" />
  <CTAHeader variant="bold">{deployError.title}</CTAHeader>
  <CTAMessage>{deployError.message}</CTAMessage>
  <CTAButton variant="secondary" href={githubAccessUrl}>
    Retry connection
  </CTAButton>
{:else}
  <CancelCircleInverse size="7rem" className="text-gray-200" />
  <CTAHeader variant="bold">{deployError.title}</CTAHeader>
  <CTAMessage>{deployError.message}</CTAMessage>
  {#if deployError.type === DeployErrorType.Unknown}
    <CTAButton variant="secondary" onClick={onRetry}>Retry</CTAButton>
  {/if}
{/if}
<CTAPylonHelp />
