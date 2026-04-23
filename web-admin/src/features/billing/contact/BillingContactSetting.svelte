<script lang="ts">
  import ChangeBillingContactDialog from "@rilldata/web-admin/features/billing/contact/ChangeBillingContactDialog.svelte";
  import { getOrganizationBillingContactUser } from "@rilldata/web-admin/features/billing/contact/selectors";
  import SettingsContainer from "@rilldata/web-admin/features/organizations/settings/SettingsContainer.svelte";
  import AvatarListItem from "@rilldata/web-common/components/avatar/AvatarListItem.svelte";
  import { Button } from "@rilldata/web-common/components/button";

  let { organization }: { organization: string } = $props();

  let billingContactUser = $derived(
    getOrganizationBillingContactUser(organization),
  );

  let isUpdateBillingContactDialogOpen = $state(false);
</script>

<SettingsContainer title="Billing Contact">
  <div class="flex flex-row items-center gap-x-1">
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
  {#snippet action()}
    <Button
      type="secondary"
      onClick={() => (isUpdateBillingContactDialogOpen = true)}
    >
      Change billing contact
    </Button>
  {/snippet}
</SettingsContainer>

<ChangeBillingContactDialog
  bind:open={isUpdateBillingContactDialogOpen}
  {organization}
  currentBillingContact={$billingContactUser?.email}
/>
