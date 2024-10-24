<script lang="ts">
  import type { V1OrganizationPermissions } from "@rilldata/web-admin/client";
  import BillingBannerManagerForAdmins from "@rilldata/web-admin/features/billing/banner/BillingBannerManagerForAdmins.svelte";
  import BillingBannerManagerForViewers from "@rilldata/web-admin/features/billing/banner/BillingBannerManagerForViewers.svelte";

  export let organization: string;
  export let organizationPermissions: V1OrganizationPermissions;
</script>

{#if organizationPermissions.manageOrg}
  <BillingBannerManagerForAdmins {organization} />
{:else if organizationPermissions.readOrg}
  <BillingBannerManagerForViewers {organization} />
{/if}
