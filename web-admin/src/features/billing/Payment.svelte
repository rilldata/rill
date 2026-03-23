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

  export let organization: string;

  $: org = createAdminServiceGetOrganization(organization);
  $: subscriptionQuery = createAdminServiceGetBillingSubscription(organization);
  $: plan = $subscriptionQuery?.data?.subscription?.plan;
  $: categorisedIssues = useCategorisedOrganizationBillingIssues(organization);
  $: paymentIssues = $categorisedIssues.data?.payment;
  $: neverSubscribed = !!$categorisedIssues.data?.neverSubscribed;
  $: onTrial = !!$categorisedIssues.data?.trial;
  $: onManagedPlan = plan && isManagedPlan(plan.name);
  $: onEnterprisePlan = plan && isEnterprisePlan(plan.name);
  // For enterprise and managed orgs, hide when payment details haven't been
  // entered yet (setup done via CLI). Once set up, show the Manage button.
  // neverSubscribed orgs are always hidden since the billing page is not shown.
  $: pendingSetup =
    neverSubscribed ||
    ((onManagedPlan || onEnterprisePlan) &&
      needsPaymentSetup(paymentIssues ?? []));

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
    <div slot="body" class="flex flex-row items-center gap-x-1">
      {#if paymentIssues?.length}
        <CancelCircle className="text-red-600" size="14px" />
        {getPaymentIssueErrorText(paymentIssues)} Please click Manage below to correct.
      {:else}
        Your payment method is valid and good to go.
      {/if}
    </div>
    <Button slot="action" type="secondary" onClick={handleManagePayment}>
      Manage
    </Button>
  </SettingsContainer>
{/if}
