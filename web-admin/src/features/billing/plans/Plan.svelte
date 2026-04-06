<script lang="ts">
  import { createAdminServiceGetBillingSubscription } from "@rilldata/web-admin/client";
  import CancelledTeamPlan from "@rilldata/web-admin/features/billing/plans/CancelledTeamPlan.svelte";
  import EnterprisePlan from "@rilldata/web-admin/features/billing/plans/EnterprisePlan.svelte";
  import POCPlan from "@rilldata/web-admin/features/billing/plans/POCPlan.svelte";
  import TeamPlan from "@rilldata/web-admin/features/billing/plans/TeamPlan.svelte";
  import TrialPlan from "@rilldata/web-admin/features/billing/plans/TrialPlan.svelte";
  import {
    isManagedPlan,
    isTeamPlan,
  } from "@rilldata/web-admin/features/billing/plans/utils";
  import { useCategorisedOrganizationBillingIssues } from "@rilldata/web-admin/features/billing/selectors";

  let {
    organization,
    showUpgradeDialog,
    billingPortalUrl,
  }: {
    organization: string;
    showUpgradeDialog: boolean;
    billingPortalUrl: string | undefined;
  } = $props();

  let subscriptionQuery = $derived(
    createAdminServiceGetBillingSubscription(organization),
  );
  let subscription = $derived($subscriptionQuery?.data?.subscription);
  let hasPayment = $derived(
    !!$subscriptionQuery?.data?.organization?.paymentCustomerId,
  );
  let plan = $derived(subscription?.plan);

  let categorisedIssues = $derived(
    useCategorisedOrganizationBillingIssues(organization),
  );

  // fresh orgs will have a never subscribed issue associated with it
  let neverSubbed = $derived(!!$categorisedIssues.data?.neverSubscribed);
  // trial plan will have a trial issue associated with it
  let isTrial = $derived(!!$categorisedIssues.data?.trial);
  // ended subscription will have a cancelled issue associated with it
  let subHasEnded = $derived(!!$categorisedIssues.data?.cancelled);
  let subIsTeamPlan = $derived(plan && isTeamPlan(plan.name));
  let subIsManagedPlan = $derived(plan && isManagedPlan(plan.name));
  let subIsEnterprisePlan = $derived(
    plan && !isTrial && !subIsTeamPlan && !subIsManagedPlan,
  );
</script>

{#if neverSubbed}
  <!-- TODO: once mocks are in. Right now we just disable the routes. -->
{:else if isTrial}
  <TrialPlan {organization} {subscription} {showUpgradeDialog} {plan} />
{:else if subHasEnded}
  <CancelledTeamPlan
    {organization}
    {showUpgradeDialog}
    {plan}
    {billingPortalUrl}
  />
{:else if subIsTeamPlan}
  <TeamPlan {organization} {subscription} {plan} {billingPortalUrl} />
{:else if subIsManagedPlan}
  <POCPlan {organization} {hasPayment} {plan} {billingPortalUrl} />
{:else if subIsEnterprisePlan}
  <EnterprisePlan {organization} {plan} />
{/if}
