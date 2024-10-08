<script lang="ts">
  import { getPaymentIssueErrorText } from "@rilldata/web-admin/features/billing/banner/handlePaymentBillingIssues";
  import { fetchPaymentsPortalURL } from "@rilldata/web-admin/features/billing/plans/selectors";
  import { useCategorisedOrganizationBillingIssues } from "@rilldata/web-admin/features/billing/selectors";
  import SettingsContainer from "@rilldata/web-admin/features/organizations/settings/SettingsContainer.svelte";
  import { Button } from "@rilldata/web-common/components/button";

  export let organization: string;

  $: categorisedIssues = useCategorisedOrganizationBillingIssues(organization);

  $: showPaymentBlock =
    !$categorisedIssues.isLoading &&
    // fresh orgs do not need payment
    !$categorisedIssues.data.neverSubscribed &&
    // orgs on trial do not need payment either
    !$categorisedIssues.data.trial;

  async function handleManagePayment() {
    window.open(
      await fetchPaymentsPortalURL(organization, window.location.href),
      "_self",
    );
  }
</script>

{#if showPaymentBlock}
  <SettingsContainer
    title="Payment Method"
    titleIcon={$categorisedIssues.data?.payment?.length ? "error" : "none"}
  >
    <div slot="body">
      {#if $categorisedIssues.data?.payment}
        {getPaymentIssueErrorText($categorisedIssues.data.payment)} Please click
        <b>Manage</b> below to correct.
      {:else}
        Your payment method is valid and good to go.
      {/if}
    </div>
    <Button slot="action" type="secondary" on:click={handleManagePayment}>
      Manage
    </Button>
  </SettingsContainer>
{/if}
