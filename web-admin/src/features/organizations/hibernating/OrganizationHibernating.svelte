<script lang="ts">
  import type { V1OrganizationPermissions } from "@rilldata/web-admin/client";
  import OrganizationHibernatingForAdmins from "@rilldata/web-admin/features/organizations/hibernating/OrganizationHibernatingForAdmins.svelte";
  import OrganizationHibernatingForViewers from "@rilldata/web-admin/features/organizations/hibernating/OrganizationHibernatingForViewers.svelte";

  export let organization: string;
  export let organizationPermissions: V1OrganizationPermissions;
</script>

{#if organizationPermissions.manageOrg}
  <OrganizationHibernatingForAdmins {organization} />
{:else if organizationPermissions.readOrg}
  <OrganizationHibernatingForViewers />
{/if}
