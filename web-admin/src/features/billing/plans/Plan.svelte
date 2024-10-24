<script lang="ts">
  import { createAdminServiceGetBillingSubscription } from "@rilldata/web-admin/client";
  import CancelledTeamPlan from "@rilldata/web-admin/features/billing/plans/CancelledTeamPlan.svelte";
  import EnterprisePlan from "@rilldata/web-admin/features/billing/plans/EnterprisePlan.svelte";
  import POCPlan from "@rilldata/web-admin/features/billing/plans/POCPlan.svelte";
  import TeamPlan from "@rilldata/web-admin/features/billing/plans/TeamPlan.svelte";
  import TrialPlan from "@rilldata/web-admin/features/billing/plans/TrialPlan.svelte";
  import {
    isPOCPlan,
    isTeamPlan,
  } from "@rilldata/web-admin/features/billing/plans/utils";
  import { useCategorisedOrganizationBillingIssues } from "@rilldata/web-admin/features/billing/selectors";

  export let organization: string;
  export let showUpgrade: boolean;

  $: subscriptionQuery = createAdminServiceGetBillingSubscription(organization);
  $: subscription = $subscriptionQuery?.data?.subscription;
  $: hasPayment = !!$subscriptionQuery?.data?.organization?.paymentCustomerId;

  $: categorisedIssues = useCategorisedOrganizationBillingIssues(organization);

  // fresh orgs will have a never subscribed issue associated with it
  $: neverSubbed = !!$categorisedIssues.data?.neverSubscribed;
  // trial plan will have a trial issue associated with it
  $: isTrial = !!$categorisedIssues.data?.trial;
  // ended subscription will have a cancelled issue associated with it
  $: hasEnded = !!$categorisedIssues.data?.cancelled;
  $: subIsTeamPlan = subscription?.plan && isTeamPlan(subscription.plan);
  $: subIsPOCPlan = subscription?.plan && isPOCPlan(subscription.plan);
</script>

{#if neverSubbed}
  No subscription (TODO)
{:else if isTrial}
  <TrialPlan {organization} {subscription} {showUpgrade} />
{:else if hasEnded}
  <CancelledTeamPlan {organization} {subscription} {showUpgrade} />
{:else if subIsTeamPlan}
  <TeamPlan {organization} {subscription} />
{:else if subIsPOCPlan}
  <POCPlan {organization} plan={subscription.plan} {hasPayment} />
{:else if subscription?.plan}
  <EnterprisePlan {organization} plan={subscription.plan} />
{/if}
