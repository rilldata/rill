<script lang="ts">
  import {
    getPaymentIssues,
    PaymentBillingIssueTypes,
  } from "@rilldata/web-admin/features/billing/banner/handlePaymentBillingIssues";
  import { fetchPaymentsPortalURL } from "@rilldata/web-admin/features/billing/plans/selectors";
  import { isTrialPlan } from "@rilldata/web-admin/features/billing/plans/utils";
  import { getPlanForOrg } from "@rilldata/web-admin/features/billing/selectors";
  import SettingsContainer from "@rilldata/web-admin/features/organizations/settings/SettingsContainer.svelte";
  import { Button } from "@rilldata/web-common/components/button";

  export let organization: string;

  $: paymentIssues = getPaymentIssues(organization);
  $: paymentIssueTexts =
    $paymentIssues.data?.map((i) => PaymentBillingIssueTypes[i.type]) ?? [];

  $: plan = getPlanForOrg(organization);
  $: isTrial = $plan.data && isTrialPlan($plan.data);
  async function handleManagePayment() {
    window.open(
      await fetchPaymentsPortalURL(organization, window.location.href),
      "_self",
    );
  }
</script>

{#if $plan.data && !isTrial}
  <SettingsContainer
    title="Payment Method"
    titleIcon={paymentIssueTexts.length ? "error" : "none"}
  >
    <div slot="body">
      {#if paymentIssueTexts.length}
        {paymentIssueTexts.join("")} Please click <b>Manage</b> below to correct.
      {:else}
        Your payment method is valid and good to go.
      {/if}
    </div>
    <Button slot="action" type="secondary" on:click={handleManagePayment}>
      Manage
    </Button>
  </SettingsContainer>
{/if}
