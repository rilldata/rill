<script lang="ts">
  import { createAdminServiceGetBillingSubscription } from "@rilldata/web-admin/client";
  import EndedTeamPlan from "@rilldata/web-admin/features/billing/plans/EndedTeamPlan.svelte";
  import EnterprisePlan from "@rilldata/web-admin/features/billing/plans/EnterprisePlan.svelte";
  import TeamPlan from "@rilldata/web-admin/features/billing/plans/TeamPlan.svelte";
  import TrialPlan from "@rilldata/web-admin/features/billing/plans/TrialPlan.svelte";
  import { getPlanForOrg } from "@rilldata/web-admin/features/billing/selectors";

  export let organization: string;

  $: subscriptionQuery = createAdminServiceGetBillingSubscription(organization);
  $: plan = getPlanForOrg(organization);
  $: subscription = $subscriptionQuery?.data?.subscription;

  $: isTrial = !!subscription?.trialEndDate;
  $: hasEnded = !!subscription?.endDate;
  $: isBilled = !!subscription?.currentBillingCycleEndDate;
</script>

{#if $plan}
  {#if isTrial}
    <TrialPlan {organization} plan={$plan} {subscription} />
  {:else if isBilled}
    <TeamPlan {organization} plan={$plan} {subscription} />
  {:else if hasEnded}
    <EndedTeamPlan {organization} plan={$plan} {subscription} />
  {:else}
    <EnterprisePlan {organization} plan={$plan} />
  {/if}
{/if}
