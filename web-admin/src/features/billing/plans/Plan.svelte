<script lang="ts">
  import {
    createAdminServiceGetBillingSubscription,
    createAdminServiceListOrganizationBillingIssues,
  } from "@rilldata/web-admin/client";
  import {
    getCancelledIssue,
    getNeverSubscribedIssue,
  } from "@rilldata/web-admin/features/billing/banner/handleSubscriptionIssues";
  import { getTrialIssue } from "@rilldata/web-admin/features/billing/banner/handleTrialPlan";
  import EndedTeamPlan from "@rilldata/web-admin/features/billing/plans/EndedTeamPlan.svelte";
  import EnterprisePlan from "@rilldata/web-admin/features/billing/plans/EnterprisePlan.svelte";
  import TeamPlan from "@rilldata/web-admin/features/billing/plans/TeamPlan.svelte";
  import TrialPlan from "@rilldata/web-admin/features/billing/plans/TrialPlan.svelte";
  import { isTeamPlan } from "@rilldata/web-admin/features/billing/plans/utils";

  export let organization: string;

  $: subscriptionQuery = createAdminServiceGetBillingSubscription(organization);
  $: subscription = $subscriptionQuery?.data?.subscription;
  $: issues = createAdminServiceListOrganizationBillingIssues(organization);

  $: neverSubbedIssue = getNeverSubscribedIssue($issues.data?.issues ?? []);
  $: cancelledIssue = getCancelledIssue($issues.data?.issues ?? []);
  $: trialIssue = getTrialIssue($issues.data?.issues ?? []);

  // fresh orgs will have a never subscribed issue associated with it
  $: neverSubbed = !!neverSubbedIssue;
  // trial plan will have a trial issue associated with it
  $: isTrial = !!trialIssue;
  // ended subscription will have a cancelled issue associated with it
  $: hasEnded = !!cancelledIssue;
  $: subIsTeamPlan = subscription?.plan && isTeamPlan(subscription.plan);
</script>

{#if neverSubbed}
  No subscription (TODO)
{:else if isTrial}
  <TrialPlan {organization} {subscription} />
{:else if hasEnded}
  <EndedTeamPlan {organization} {subscription} />
{:else if subIsTeamPlan}
  <TeamPlan {organization} {subscription} />
{:else if subscription?.plan}
  <EnterprisePlan {organization} plan={subscription.plan} />
{/if}
