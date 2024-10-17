<script lang="ts">
  import OrganizationHibernatingForAdmins from "@rilldata/web-admin/features/organizations/hibernating/OrganizationHibernatingForAdmins.svelte";
  import OrganizationHibernatingForViewers from "@rilldata/web-admin/features/organizations/hibernating/OrganizationHibernatingForViewers.svelte";
  import OrganizationAccessControls from "@rilldata/web-admin/features/organizations/OrganizationAccessControls.svelte";

  export let organization: string;
</script>

<OrganizationAccessControls {organization}>
  <OrganizationHibernatingForAdmins {organization} slot="manage-org" />
  <OrganizationHibernatingForViewers slot="read-projects" />
</OrganizationAccessControls>
