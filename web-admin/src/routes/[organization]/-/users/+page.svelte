<script lang="ts">
  import { page } from "$app/stores";
  import { createAdminServiceListOrganizationMemberUsers } from "@rilldata/web-admin/client";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import OrgUsersTable from "@rilldata/web-admin/features/organizations/users/OrgUsersTable.svelte";

  $: organization = $page.params.organization;

  $: organizationMemberUsers =
    createAdminServiceListOrganizationMemberUsers(organization);
</script>

<div class="flex flex-col w-full">
  {#if $organizationMemberUsers.isLoading}
    <DelayedSpinner
      isLoading={$organizationMemberUsers.isLoading}
      size="1rem"
    />
  {:else if $organizationMemberUsers.isError}
    <div class="text-red-500">
      Error loading organization members: {$organizationMemberUsers.error}
    </div>
  {:else if $organizationMemberUsers.isSuccess}
    <OrgUsersTable data={$organizationMemberUsers.data.members} />
  {/if}
</div>
