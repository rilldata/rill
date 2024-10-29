<script lang="ts">
  import { createAdminServiceGetOrganization } from "@rilldata/web-admin/client";
  import { getPaymentIssueErrorText } from "@rilldata/web-admin/features/billing/issues/getMessageForPaymentIssues";
  import { fetchPaymentsPortalURL } from "@rilldata/web-admin/features/billing/plans/selectors";
  import { useCategorisedOrganizationBillingIssues } from "@rilldata/web-admin/features/billing/selectors";
  import SettingsContainer from "@rilldata/web-admin/features/organizations/settings/SettingsContainer.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import CancelCircle from "@rilldata/web-common/components/icons/CancelCircle.svelte";

  export let organization: string;

  $: org = createAdminServiceGetOrganization(organization);
  $: categorisedIssues = useCategorisedOrganizationBillingIssues(organization);
  $: paymentIssues = $categorisedIssues.data?.payment;
  $: neverSubscribed = $categorisedIssues.data?.neverSubscribed;

  async function handleManagePayment() {
    window.open(
      await fetchPaymentsPortalURL(organization, window.location.href),
      "_self",
    );
  }
</script>

<!-- Presence of paymentCustomerId signifies that the org's payment is managed through stripe -->
{#if !$categorisedIssues.isLoading && !neverSubscribed && $org.data?.organization?.paymentCustomerId}
  <SettingsContainer title="Payment Method">
    <div slot="body" class="flex flex-row items-center gap-x-1">
      {#if paymentIssues?.length}
        <CancelCircle className="text-red-600" size="14px" />
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
