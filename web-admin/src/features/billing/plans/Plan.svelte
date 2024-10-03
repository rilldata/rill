<script lang="ts">
  import { createAdminServiceGetBillingSubscription } from "@rilldata/web-admin/client";
  import EndedTeamPlan from "@rilldata/web-admin/features/billing/plans/EndedTeamPlan.svelte";
  import EnterprisePlan from "@rilldata/web-admin/features/billing/plans/EnterprisePlan.svelte";
  import TeamPlan from "@rilldata/web-admin/features/billing/plans/TeamPlan.svelte";
  import TrialPlan from "@rilldata/web-admin/features/billing/plans/TrialPlan.svelte";
  import { isTrialPlan } from "@rilldata/web-admin/features/billing/plans/utils";

  export let organization: string;

  $: subscriptionQuery = createAdminServiceGetBillingSubscription(organization);
  $: subscription = $subscriptionQuery?.data?.subscription;

  $: isTrial = subscription.plan && isTrialPlan(subscription.plan);
  $: hasEnded = !!subscription?.endDate;
  $: isBilled = !!subscription?.currentBillingCycleEndDate;
</script>

{#if subscription}
  {#if isTrial}
    <TrialPlan {organization} {subscription} />
  {:else if isBilled}
    <TeamPlan {organization} {subscription} />
  {:else if hasEnded}
    <EndedTeamPlan {organization} {subscription} />
  {:else}
    <EnterprisePlan {organization} plan={subscription.plan} />
  {/if}
{/if}
