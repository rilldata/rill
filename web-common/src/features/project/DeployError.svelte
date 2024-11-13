<script lang="ts">
  import { page } from "$app/stores";
  import { Button } from "@rilldata/web-common/components/button";
  import CTAButton from "@rilldata/web-common/components/calls-to-action/CTAButton.svelte";
  import CTAHeader from "@rilldata/web-common/components/calls-to-action/CTAHeader.svelte";
  import CTAMessage from "@rilldata/web-common/components/calls-to-action/CTAMessage.svelte";
  import CTAPylonHelp from "@rilldata/web-common/components/calls-to-action/CTAPylonHelp.svelte";
  import CancelCircleInverse from "@rilldata/web-common/components/icons/CancelCircleInverse.svelte";
  import PricingDetails from "@rilldata/web-common/features/billing/PricingDetails.svelte";
  import { buildPlanUpgradeUrl } from "@rilldata/web-common/features/organization/utils";
  import {
    type DeployError,
    DeployErrorType,
  } from "@rilldata/web-common/features/project/deploy-errors";

  export let org: string;
  export let error: DeployError;
  export let adminUrl: string;
  export let onRetry: () => void;

  $: isQuotaError =
    error.type === DeployErrorType.ProjectLimitHit ||
    error.type === DeployErrorType.OrgLimitHit ||
    error.type === DeployErrorType.SubscriptionEnded;

  $: upgradeHref = buildPlanUpgradeUrl(org, adminUrl, $page.url);
</script>

{#if isQuotaError}
  <CTAHeader variant="bold">{error.title}</CTAHeader>
  <p class="text-base text-gray-500 text-left w-[500px]">
    <PricingDetails extraText={error.message} />
  </p>
  <Button type="primary" href={upgradeHref} wide on:click>Upgrade</Button>
  <Button type="secondary" noStroke wide href="/">Back</Button>
{:else}
  <CancelCircleInverse size="7rem" className="text-gray-200" />
  <CTAHeader variant="bold">{error.title}</CTAHeader>
  <CTAMessage>{error.message}</CTAMessage>
  {#if error.type === DeployErrorType.Unknown}
    <CTAButton variant="secondary" on:click={onRetry}>Retry</CTAButton>
  {/if}
{/if}
<CTAPylonHelp />
