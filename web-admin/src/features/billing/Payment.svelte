<script lang="ts">
  import {
    createAdminServiceGetBillingSubscription,
    createAdminServiceGetOrganization,
  } from "@rilldata/web-admin/client";
  import {
    getPaymentIssueErrorText,
    needsPaymentSetup,
  } from "@rilldata/web-admin/features/billing/issues/getMessageForPaymentIssues";
  import { fetchPaymentsPortalURL } from "@rilldata/web-admin/features/billing/plans/selectors";
  import { useCategorisedOrganizationBillingIssues } from "@rilldata/web-admin/features/billing/selectors";
  import SettingsContainer from "@rilldata/web-admin/features/organizations/settings/SettingsContainer.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import CancelCircle from "@rilldata/web-common/components/icons/CancelCircle.svelte";
  import { isEnterprisePlan, isManagedPlan } from "./plans/utils";

  let { organization }: { organization: string } = $props();

  let org = $derived(createAdminServiceGetOrganization(organization));
  let subscriptionQuery = $derived(
    createAdminServiceGetBillingSubscription(organization),
  );
  let plan = $derived($subscriptionQuery?.data?.subscription?.plan);
  let categorisedIssues = $derived(
    useCategorisedOrganizationBillingIssues(organization),
  );
  let paymentIssues = $derived($categorisedIssues.data?.payment);
  let neverSubscribed = $derived(!!$categorisedIssues.data?.neverSubscribed);
  let onTrial = $derived(!!$categorisedIssues.data?.trial);
  let onManagedPlan = $derived(plan && isManagedPlan(plan.name));
  let onEnterprisePlan = $derived(plan && isEnterprisePlan(plan.name));
  // For enterprise and managed orgs, hide when payment details haven't been
  // entered yet (setup done via CLI). Once set up, show the Manage button.
  // neverSubscribed orgs are always hidden since the billing page is not shown.
  let pendingSetup = $derived(
    neverSubscribed ||
      ((onManagedPlan || onEnterprisePlan) &&
        needsPaymentSetup(paymentIssues ?? [])),
  );

  async function handleManagePayment() {
    const setup = paymentIssues?.length
      ? needsPaymentSetup(paymentIssues)
      : false;
    window.open(
      await fetchPaymentsPortalURL(organization, window.location.href, setup),
      "_self",
    );
  }
</script>

<!-- Presence of paymentCustomerId signifies that the org's payment is managed through stripe -->
{#if !$categorisedIssues.isLoading && $org.data?.organization?.paymentCustomerId && !onTrial && !pendingSetup}
  <SettingsContainer title="Payment Method">
    <div class="flex flex-row items-center gap-x-1">
      {#if paymentIssues?.length}
        <CancelCircle className="text-red-600" size="14px" />
        {getPaymentIssueErrorText(paymentIssues)} Please click Manage below to correct.
      {:else}
        Your payment method is valid and good to go.
      {/if}
    </div>
    {#snippet action()}
      <Button type="secondary" onClick={handleManagePayment}>Manage</Button>
    {/snippet}
  </SettingsContainer>
{/if}
