<script lang="ts">
  import {
    createAdminServiceGetOrganization,
    type V1OrganizationPermissions,
  } from "@rilldata/web-admin/client";
  import type { CreateQueryResult } from "@tanstack/svelte-query";

  export let organization: string;

  let orgPermissions: CreateQueryResult<V1OrganizationPermissions>;
  $: orgPermissions = createAdminServiceGetOrganization(organization, {
    query: {
      select: (data) => data.permissions,
    },
  });
</script>

{#if $orgPermissions?.data}
  {#if $orgPermissions.data.manageOrgMembers}
    <slot name="manage-org-members" />
  {/if}
{/if}
