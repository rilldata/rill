<script lang="ts">
  import { createAdminServiceGetBillingSubscription } from "@rilldata/web-admin/client";
  import CancelledTeamPlan from "@rilldata/web-admin/features/billing/plans/CancelledTeamPlan.svelte";
  import EnterprisePlan from "@rilldata/web-admin/features/billing/plans/EnterprisePlan.svelte";
  import FreePlan from "@rilldata/web-admin/features/billing/plans/FreePlan.svelte";
  import GrowthPlan from "@rilldata/web-admin/features/billing/plans/GrowthPlan.svelte";
  import POCPlan from "@rilldata/web-admin/features/billing/plans/POCPlan.svelte";
  import TeamPlan from "@rilldata/web-admin/features/billing/plans/TeamPlan.svelte";
  import TrialPlan from "@rilldata/web-admin/features/billing/plans/TrialPlan.svelte";
  import {
    isFreePlan,
    isGrowthPlan,
    isManagedPlan,
    isTeamPlan,
  } from "@rilldata/web-admin/features/billing/plans/utils";
  import { useCategorisedOrganizationBillingIssues } from "@rilldata/web-admin/features/billing/selectors";

  export let organization: string;
  export let showUpgradeDialog: boolean;

  $: subscriptionQuery = createAdminServiceGetBillingSubscription(organization);
  $: subscription = $subscriptionQuery?.data?.subscription;
  $: hasPayment = !!$subscriptionQuery?.data?.organization?.paymentCustomerId;
  $: plan = subscription?.plan;

  $: categorisedIssues = useCategorisedOrganizationBillingIssues(organization);

  // fresh orgs will have a never subscribed issue associated with it
  $: neverSubbed = !!$categorisedIssues.data?.neverSubscribed;
  // trial plan will have a trial issue associated with it
  $: isTrial = !!$categorisedIssues.data?.trial;
  // ended subscription will have a cancelled issue associated with it
  $: subHasEnded = !!$categorisedIssues.data?.cancelled;
  $: subIsFreePlan = plan && isFreePlan(plan.name);
  $: subIsGrowthPlan = plan && isGrowthPlan(plan.name);
  $: subIsTeamPlan = plan && isTeamPlan(plan.name);
  $: subIsManagedPlan = plan && isManagedPlan(plan.name);
  $: subIsEnterprisePlan =
    plan &&
    !isTrial &&
    !subIsFreePlan &&
    !subIsGrowthPlan &&
    !subIsTeamPlan &&
    !subIsManagedPlan;
</script>

{#if neverSubbed}
  <!-- TODO: once mocks are in. Right now we just disable the routes. -->
{:else if isTrial}
  <TrialPlan {organization} {subscription} {showUpgradeDialog} {plan} />
{:else if subHasEnded}
  <CancelledTeamPlan {organization} {showUpgradeDialog} {plan} />
{:else if subIsFreePlan}
  <FreePlan {organization} {plan} />
{:else if subIsGrowthPlan}
  <GrowthPlan {organization} {subscription} {plan} />
{:else if subIsTeamPlan}
  <TeamPlan {organization} {subscription} {plan} />
{:else if subIsManagedPlan}
  <POCPlan {organization} {hasPayment} {plan} />
{:else if subIsEnterprisePlan}
  <EnterprisePlan {organization} {plan} />
{/if}
