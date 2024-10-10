<script lang="ts">
  import BillingBannerForAdmins from "@rilldata/web-admin/features/billing/banner/BillingBannerForAdmins.svelte";
  import BillingBannerForViewers from "@rilldata/web-admin/features/billing/banner/BillingBannerForViewers.svelte";
  import OrganizationAccessControls from "@rilldata/web-admin/features/organizations/OrganizationAccessControls.svelte";

  export let organization: string;
</script>

<OrganizationAccessControls {organization}>
  <BillingBannerForAdmins {organization} slot="manage-org" />
  <BillingBannerForViewers {organization} slot="read-projects" />
</OrganizationAccessControls>
