<script lang="ts">
  import ChangeBillingContactDialog from "@rilldata/web-admin/features/billing/contact/ChangeBillingContactDialog.svelte";
  import { getOrganizationBillingContactUser } from "@rilldata/web-admin/features/billing/contact/selectors";
  import SettingsContainer from "@rilldata/web-admin/features/organizations/settings/SettingsContainer.svelte";
  import AvatarListItem from "@rilldata/web-common/components/avatar/AvatarListItem.svelte";
  import { Button } from "@rilldata/web-common/components/button";

  export let organization: string;

  $: billingContactUser = getOrganizationBillingContactUser(organization);

  let isUpdateBillingContactDialogOpen = false;
</script>

<SettingsContainer title="Billing Contact">
  <div slot="body" class="flex flex-row items-center gap-x-1">
    {#if $billingContactUser}
      <AvatarListItem
        name={$billingContactUser.displayName}
        email={$billingContactUser.email}
        photoUrl={$billingContactUser.photoUrl}
      />
    {:else}
      This org has no billing contact.
    {/if}
  </div>
  <Button
    slot="action"
    type="secondary"
    onClick={() => (isUpdateBillingContactDialogOpen = true)}
  >
    Change billing contact
  </Button>
</SettingsContainer>

<ChangeBillingContactDialog
  bind:open={isUpdateBillingContactDialogOpen}
  {organization}
  currentBillingContact={$billingContactUser?.email}
/>
