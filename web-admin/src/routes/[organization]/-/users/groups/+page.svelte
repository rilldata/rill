<script lang="ts">
  import { page } from "$app/stores";
  import { createAdminServiceListOrganizationMemberUsergroups } from "@rilldata/web-admin/client";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import OrgGroupsTable from "@rilldata/web-admin/features/organizations/users/OrgGroupsTable.svelte";

  $: organization = $page.params.organization;

  $: organizationMemberUserGroups =
    createAdminServiceListOrganizationMemberUsergroups(organization);
</script>

<div class="flex flex-col w-full">
  {#if $organizationMemberUserGroups.isLoading}
    <DelayedSpinner
      isLoading={$organizationMemberUserGroups.isLoading}
      size="1rem"
    />
  {:else if $organizationMemberUserGroups.isError}
    <div class="text-red-500">
      Error loading organization members: {$organizationMemberUserGroups.error}
    </div>
  {:else if $organizationMemberUserGroups.isSuccess}
    <OrgGroupsTable data={$organizationMemberUserGroups.data.members} />
  {/if}
</div>
