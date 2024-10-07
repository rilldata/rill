<script lang="ts">
  import {
    createAdminServiceGetBillingSubscription,
    createAdminServiceListOrganizationBillingIssues,
  } from "@rilldata/web-admin/client";
  import { getCancelledSubIssue } from "@rilldata/web-admin/features/billing/banner/handleSubscriptionIssues";
  import EndedTeamPlan from "@rilldata/web-admin/features/billing/plans/EndedTeamPlan.svelte";
  import EnterprisePlan from "@rilldata/web-admin/features/billing/plans/EnterprisePlan.svelte";
  import TeamPlan from "@rilldata/web-admin/features/billing/plans/TeamPlan.svelte";
  import TrialPlan from "@rilldata/web-admin/features/billing/plans/TrialPlan.svelte";
  import { isTrialPlan } from "@rilldata/web-admin/features/billing/plans/utils";

  export let organization: string;

  $: subscriptionQuery = createAdminServiceGetBillingSubscription(organization);
  $: subscription = $subscriptionQuery?.data?.subscription;
  $: issues = createAdminServiceListOrganizationBillingIssues(organization);
  $: cancelledIssue = getCancelledSubIssue($issues.data?.issues ?? []);

  $: isTrial = subscription?.plan && isTrialPlan(subscription.plan);
  $: hasEnded = !!subscription?.endDate || !!cancelledIssue;
  $: isBilled = !!subscription?.currentBillingCycleEndDate;
</script>

{#if subscription}
  {#if isTrial}
    <TrialPlan {organization} {subscription} />
  {:else if hasEnded}
    <EndedTeamPlan {organization} {subscription} />
  {:else if isBilled}
    <TeamPlan {organization} {subscription} />
  {:else}
    <EnterprisePlan {organization} plan={subscription.plan} />
  {/if}
{:else}
  No subscription (TODO)
{/if}
