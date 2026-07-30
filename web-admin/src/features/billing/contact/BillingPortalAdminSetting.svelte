<script lang="ts">
  import ChangeBillingPortalAdminDialog from "@rilldata/web-admin/features/billing/contact/ChangeBillingPortalAdminDialog.svelte";
  import { getOrganizationBillingPortalAdminUser } from "@rilldata/web-admin/features/billing/contact/selectors";
  import AvatarListItem from "@rilldata/web-common/components/avatar/AvatarListItem.svelte";
  import InfoCircle from "@rilldata/web-common/components/icons/InfoCircle.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  let { organization, canEdit }: { organization: string; canEdit: boolean } =
    $props();

  let billingPortalAdminUser = $derived(
    getOrganizationBillingPortalAdminUser(organization),
  );

  let isUpdateBillingPortalAdminDialogOpen = $state(false);
</script>

<section>
  <h2 class="section-header">
    {m.billing_portal_admin_header()}
    <Tooltip location="right" alignment="middle" distance={8}>
      <span class="text-fg-muted flex">
        <InfoCircle size="16px" />
      </span>
      <TooltipContent maxWidth="240px" slot="tooltip-content">
        {m.billing_portal_admin_tooltip()}
      </TooltipContent>
    </Tooltip>
  </h2>
  <div class="section-card">
    <div class="card-content">
      {#if $billingPortalAdminUser}
        <AvatarListItem
          name={$billingPortalAdminUser.displayName ||
            $billingPortalAdminUser.email ||
            ""}
          email={$billingPortalAdminUser.email}
          photoUrl={$billingPortalAdminUser.photoUrl}
        />
      {:else}
        <span class="text-sm text-fg-tertiary"
          >{m.billing_no_billing_portal_admin()}</span
        >
      {/if}
    </div>
    {#if canEdit}
      <button
        class="manage-btn"
        onclick={() => (isUpdateBillingPortalAdminDialogOpen = true)}
      >
        {m.billing_change_billing_portal_admin()}
      </button>
    {/if}
  </div>
</section>

<ChangeBillingPortalAdminDialog
  bind:open={isUpdateBillingPortalAdminDialogOpen}
  {organization}
  currentBillingPortalAdmin={$billingPortalAdminUser?.email}
/>

<style lang="postcss">
  .section-header {
    @apply text-lg font-medium text-fg-primary mb-3;
    @apply flex items-center gap-x-1.5;
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
