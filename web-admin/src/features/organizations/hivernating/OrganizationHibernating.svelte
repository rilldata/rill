<script lang="ts">
  import OrganizationHibernatingForAdmins from "@rilldata/web-admin/features/organizations/hivernating/OrganizationHibernatingForAdmins.svelte";
  import OrganizationHibernatingForViewers from "@rilldata/web-admin/features/organizations/hivernating/OrganizationHibernatingForViewers.svelte";
  import OrganizationAccessControls from "@rilldata/web-admin/features/organizations/OrganizationAccessControls.svelte";

  export let organization: string;
</script>

<OrganizationAccessControls {organization}>
  <OrganizationHibernatingForAdmins {organization} slot="manage-org" />
  <OrganizationHibernatingForViewers slot="read-projects" />
</OrganizationAccessControls>
