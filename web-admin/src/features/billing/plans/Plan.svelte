<script lang="ts">
  import {
    createAdminServiceGetBillingSubscription,
    type V1OrganizationQuotas,
  } from "@rilldata/web-admin/client";
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
  export let showUpgradeDialog: boolean;
  export let organizationQuotas: V1OrganizationQuotas;

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
  $: subIsEnterprisePlan =
    subscription?.plan && !isTrial && !subIsTeamPlan && !subIsPOCPlan;
</script>

{#if neverSubbed}
  <!-- TODO: once mocks are in. Right now we just disable the routes. -->
{:else if isTrial}
  <TrialPlan
    {organization}
    {subscription}
    {showUpgradeDialog}
    {organizationQuotas}
  />
{:else if hasEnded}
  <CancelledTeamPlan {organization} {showUpgradeDialog} {organizationQuotas} />
{:else if subIsTeamPlan}
  <TeamPlan {organization} {subscription} {organizationQuotas} />
{:else if subIsPOCPlan}
  <POCPlan {organization} {hasPayment} {organizationQuotas} />
{:else if subIsEnterprisePlan}
  <EnterprisePlan {organization} {organizationQuotas} />
{/if}
