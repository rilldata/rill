<script lang="ts">
  import {
    createAdminServiceGetPaymentsPortalURL,
    createAdminServiceListOrganizationBillingIssues,
  } from "@rilldata/web-admin/client";
  import { PaymentBillingIssueTypes } from "@rilldata/web-admin/features/billing/banner/handlePaymentBillingIssues";
  import { getCategorisedPlans } from "@rilldata/web-admin/features/billing/plans/selectors";
  import { getPlanForOrg } from "@rilldata/web-admin/features/billing/selectors";
  import SettingsContainer from "@rilldata/web-admin/features/organizations/settings/SettingsContainer.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import { page } from "$app/stores";

  export let organization: string;

  $: paymentUrl = createAdminServiceGetPaymentsPortalURL(organization, {
    returnUrl: $page.url.toString(),
  });
  $: issues = createAdminServiceListOrganizationBillingIssues(organization);
  $: paymentIssues =
    $issues.data?.issues
      ?.filter((i) => i.type in PaymentBillingIssueTypes)
      .map((i) => PaymentBillingIssueTypes[i.type]) ?? [];

  $: plan = getPlanForOrg(organization);
  const categorisedPlans = getCategorisedPlans();
  $: isTrial = $plan?.id === $categorisedPlans.data?.trialPlan?.id;
</script>

{#if !isTrial}
  <SettingsContainer
    title="Payment Method"
    titleIcon={paymentIssues.length ? "error" : "none"}
  >
    <div slot="body">
      {#if paymentIssues.length}
        {paymentIssues.join("")} Please click <b>Manage</b> below to correct.
      {:else}
        Your payment method is valid and good to go.
      {/if}
    </div>
    <Button slot="action" type="secondary" href={$paymentUrl.data?.url}>
      Manage
    </Button>
  </SettingsContainer>
{/if}
