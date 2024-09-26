<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceDeleteUsergroup,
    createAdminServiceListOrganizationMemberUsergroups,
    getAdminServiceListOrganizationMemberUsergroupsQueryKey,
  } from "@rilldata/web-admin/client";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import OrgGroupsTable from "@rilldata/web-admin/features/organizations/users/OrgGroupsTable.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { useQueryClient } from "@tanstack/svelte-query";

  $: organization = $page.params.organization;

  $: organizationMemberUserGroups =
    createAdminServiceListOrganizationMemberUsergroups(organization);

  const queryClient = useQueryClient();
  const deleteUserGroup = createAdminServiceDeleteUsergroup();

  async function handleDelete(deletedUserGroupId: string) {
    try {
      await $deleteUserGroup.mutateAsync({
        organization: organization,
        usergroup: deletedUserGroupId,
      });

      await queryClient.invalidateQueries(
        getAdminServiceListOrganizationMemberUsergroupsQueryKey(organization),
      );

      eventBus.emit("notification", { message: "User group deleted" });
    } catch (error) {
      eventBus.emit("notification", {
        message: "Error deleting user group",
        type: "error",
      });
    }
  }
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
    <OrgGroupsTable
      data={$organizationMemberUserGroups.data.members}
      onDelete={handleDelete}
    />
  {/if}
</div>
