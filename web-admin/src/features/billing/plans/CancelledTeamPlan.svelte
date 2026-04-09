<script lang="ts">
  import type { V1BillingPlan } from "@rilldata/web-admin/client";
  import ContactUs from "@rilldata/web-admin/features/billing/ContactUs.svelte";
  import PlanQuotas from "@rilldata/web-admin/features/billing/plans/PlanQuotas.svelte";
  import StartTeamPlanDialog from "@rilldata/web-admin/features/billing/plans/StartTeamPlanDialog.svelte";
  import { useCategorisedOrganizationBillingIssues } from "@rilldata/web-admin/features/billing/selectors";
  import SettingsContainer from "@rilldata/web-admin/features/organizations/settings/SettingsContainer.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import InfoCircleFilled from "@rilldata/web-common/components/icons/InfoCircleFilled.svelte";
  import { DateTime } from "luxon";

  let {
    organization,
    plan,
    showUpgradeDialog,
    billingPortalUrl,
  }: {
    organization: string;
    plan: V1BillingPlan;
    showUpgradeDialog: boolean;
    billingPortalUrl: string | undefined;
  } = $props();

  let categorisedIssues = $derived(
    useCategorisedOrganizationBillingIssues(organization),
  );
  let cancelledSubIssue = $derived($categorisedIssues.data?.cancelled);

  let willEndOnText = $derived.by(() => {
    if (cancelledSubIssue?.metadata.subscriptionCancelled?.endDate) {
      const endDate = DateTime.fromJSDate(
        new Date(cancelledSubIssue.metadata.subscriptionCancelled.endDate),
      );
      if (endDate.isValid && endDate.toMillis() > Date.now())
        return endDate.toLocaleString(DateTime.DATE_MED);
    }
    return "";
  });

  let open = $state(false);
  $effect(() => {
    if (showUpgradeDialog) open = true;
  });
</script>

<SettingsContainer title={plan?.displayName || "Team plan"}>
  <div>
    <div class="flex flex-row items-center gap-x-1 text-sm">
      <InfoCircleFilled className="text-yellow-500" size="14px" />
      Your plan is cancelled
      {#if willEndOnText}
        but you still have access until <b>{willEndOnText}.</b>
      {:else}
        and your subscription has ended.
      {/if}
    </div>
    {#if billingPortalUrl}
      <div>
        <a
          href={billingPortalUrl}
          target="_blank"
          rel="noreferrer noopener"
          class="invoice-link">View Invoice</a
        >
      </div>
    {/if}
    {#if plan}
      <!-- if there is no plan then quotas will be set to 0. It doesnt make sense to show this then -->
      <PlanQuotas {organization} />
    {/if}
  </div>
  {#snippet contact()}
    <span>For custom enterprise needs,</span>
    <ContactUs />
  {/snippet}

  {#snippet action()}
    <Button type="primary" onClick={() => (open = true)}>
      Renew Team plan
    </Button>
  {/snippet}
</SettingsContainer>

{#if !$categorisedIssues.isLoading}
  <StartTeamPlanDialog
    bind:open
    {organization}
    type="renew"
    endDate={cancelledSubIssue?.metadata.subscriptionCancelled?.endDate}
  />
{/if}

<style lang="postcss">
  .invoice-link {
    @apply text-sm text-primary-500 no-underline mt-2 inline-block;
  }
  .invoice-link:hover {
    @apply text-primary-600 underline;
  }
</style>
