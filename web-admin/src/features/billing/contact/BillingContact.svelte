<script lang="ts">
  import {
    createAdminServiceGetOrganization,
    type V1Subscription,
  } from "@rilldata/web-admin/client";
  import UpdateBillingContactDialog from "@rilldata/web-admin/features/billing/contact/UpdateBillingContactDialog.svelte";
  import { getPaymentIssueErrorText } from "@rilldata/web-admin/features/billing/issues/getMessageForPaymentIssues";
  import SettingsContainer from "@rilldata/web-admin/features/organizations/settings/SettingsContainer.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import CancelCircle from "@rilldata/web-common/components/icons/CancelCircle.svelte";

  export let organization: string;

  $: org = createAdminServiceGetOrganization(organization);
  $: billingContact = $org.data?.organization?.billingEmail;

  let openUpdateBillingContactDialog = false;
</script>

<SettingsContainer title="Billing Contact">
  <div slot="body" class="flex flex-row items-center gap-x-1">
    {#if billingContact}
      Current billing contact is "{billingContact}"
    {:else}
      This org has no billing contact.
    {/if}
  </div>
  <Button
    slot="action"
    type="secondary"
    on:click={() => (openUpdateBillingContactDialog = true)}
  >
    Update
  </Button>
</SettingsContainer>

<UpdateBillingContactDialog
  bind:open={openUpdateBillingContactDialog}
  {organization}
  currentBillingContact={billingContact}
/>
