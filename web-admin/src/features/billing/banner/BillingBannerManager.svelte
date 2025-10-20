<script lang="ts">
  import { onNavigate } from "$app/navigation";
  import type { V1OrganizationPermissions } from "@rilldata/web-admin/client";
  import BillingBannerManagerForAdmins from "@rilldata/web-admin/features/billing/banner/BillingBannerManagerForAdmins.svelte";
  import BillingBannerManagerForViewers from "@rilldata/web-admin/features/billing/banner/BillingBannerManagerForViewers.svelte";
  import { BillingBannerID } from "@rilldata/web-common/components/banner/constants";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";

  export let organization: string;
  export let organizationPermissions: V1OrganizationPermissions;

  onNavigate(({ from, to }) => {
    const changedOrganization =
      !from || !to || from.params.organization !== to.params.organization;
    if (changedOrganization) {
      eventBus.emit("remove-banner", BillingBannerID);
    }
  });
</script>

{#if organizationPermissions.manageOrg}
  <BillingBannerManagerForAdmins {organization} />
{:else if organizationPermissions.readOrg && organizationPermissions.readProjects}
  <BillingBannerManagerForViewers {organization} />
{/if}
