<script lang="ts">
  import { createAdminServiceGetPaymentsPortalURL } from "@rilldata/web-admin/client";
  import {
    getPaymentIssues,
    PaymentBillingIssueTypes,
  } from "@rilldata/web-admin/features/billing/banner/handlePaymentBillingIssues";
  import { isTrialPlan } from "@rilldata/web-admin/features/billing/plans/utils";
  import { getPlanForOrg } from "@rilldata/web-admin/features/billing/selectors";
  import SettingsContainer from "@rilldata/web-admin/features/organizations/settings/SettingsContainer.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import { page } from "$app/stores";

  export let organization: string;

  $: paymentUrl = createAdminServiceGetPaymentsPortalURL(organization, {
    returnUrl: $page.url.toString(),
  });
  $: paymentIssues = getPaymentIssues(organization);
  $: paymentIssueTexts = $paymentIssues.data?.map(
    (i) => PaymentBillingIssueTypes[i.type],
  );

  $: plan = getPlanForOrg(organization);
  $: isTrial = $plan.data && isTrialPlan($plan.data);
</script>

{#if !isTrial}
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
    <Button slot="action" type="secondary" href={$paymentUrl.data?.url}>
      Manage
    </Button>
  </SettingsContainer>
{/if}
