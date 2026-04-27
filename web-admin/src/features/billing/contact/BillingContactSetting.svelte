<script lang="ts">
  import ChangeBillingContactDialog from "@rilldata/web-admin/features/billing/contact/ChangeBillingContactDialog.svelte";
  import { getOrganizationBillingContactUser } from "@rilldata/web-admin/features/billing/contact/selectors";
  import AvatarListItem from "@rilldata/web-common/components/avatar/AvatarListItem.svelte";

  let { organization }: { organization: string } = $props();

  let billingContactUser = $derived(
    getOrganizationBillingContactUser(organization),
  );

  let isUpdateBillingContactDialogOpen = $state(false);
</script>

<section>
  <h2 class="section-header">Billing Contact</h2>
  <div class="section-card">
    <div class="card-content">
      {#if $billingContactUser}
        <AvatarListItem
          name={$billingContactUser.displayName}
          email={$billingContactUser.email}
          photoUrl={$billingContactUser.photoUrl}
        />
      {:else}
        <span class="text-sm text-fg-tertiary"
          >This org has no billing contact.</span
        >
      {/if}
    </div>
    <button
      class="manage-btn"
      onclick={() => (isUpdateBillingContactDialogOpen = true)}
    >
      Change billing contact
    </button>
  </div>
</section>

<ChangeBillingContactDialog
  bind:open={isUpdateBillingContactDialogOpen}
  {organization}
  currentBillingContact={$billingContactUser?.email}
/>

<style lang="postcss">
  .section-header {
    @apply text-lg font-medium text-fg-primary mb-3;
  }

  .section-card {
    @apply flex items-center justify-between border rounded-lg p-4 bg-surface-background;
    box-shadow:
      0px 1px 2px 0px rgba(0, 0, 0, 0.06),
      0px 1px 3px 0px rgba(0, 0, 0, 0.1);
  }

  .card-content {
    @apply flex items-center;
  }

  .manage-btn {
    @apply flex items-center gap-1.5 text-sm font-medium text-primary-600 border border-primary-500 rounded-sm px-4 py-2 bg-transparent cursor-pointer;
  }

  .manage-btn:hover {
    @apply bg-primary-50;
  }
</style>
