<script lang="ts">
  import { getPaymentIssueErrorText } from "@rilldata/web-admin/features/billing/issues/handlePaymentBillingIssues";
  import { fetchPaymentsPortalURL } from "@rilldata/web-admin/features/billing/plans/selectors";
  import { useCategorisedOrganizationBillingIssues } from "@rilldata/web-admin/features/billing/selectors";
  import SettingsContainer from "@rilldata/web-admin/features/organizations/settings/SettingsContainer.svelte";
  import { Button } from "@rilldata/web-common/components/button";

  export let organization: string;

  $: categorisedIssues = useCategorisedOrganizationBillingIssues(organization);
  $: paymentIssues = $categorisedIssues.data?.payment;

  async function handleManagePayment() {
    window.open(
      await fetchPaymentsPortalURL(organization, window.location.href),
      "_self",
    );
  }
</script>

{#if !$categorisedIssues.isLoading}
  <SettingsContainer
    title="Payment Method"
    titleIcon={paymentIssues?.length ? "error" : "none"}
  >
    <div slot="body">
      {#if paymentIssues?.length}
        {getPaymentIssueErrorText(paymentIssues)} Please click Manage below to correct.
      {:else}
        Your payment method is valid and good to go.
      {/if}
    </div>
    <Button slot="action" type="secondary" on:click={handleManagePayment}>
      Manage
    </Button>
  </SettingsContainer>
{/if}
