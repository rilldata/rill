<script lang="ts">
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import { Trash2Icon } from "lucide-svelte";
  import RemoveUserFromOrgConfirmDialog from "./RemoveUserFromOrgConfirmDialog.svelte";
  import {
    createAdminServiceRemoveOrganizationMemberUser,
    createAdminServiceUpdateOrganization,
    getAdminServiceListOrganizationInvitesQueryKey,
    getAdminServiceListOrganizationMemberUsersQueryKey,
    getAdminServiceListUsergroupMemberUsersQueryKey,
  } from "@rilldata/web-admin/client";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { page } from "$app/stores";

  export let organization: string;
  export let name: string;
  export let email: string;
  export let isCurrentUser: boolean;
  export let isAdminUser: boolean;
  // Used in "Assign as billing contact". So we use manageOrg instead of manageUsers
  export let currentUserCanManageOrg: boolean;

  let isDropdownOpen = false;
  let isRemoveConfirmOpen = false;

  $: organization = $page.params.organization;

  const queryClient = useQueryClient();
  const removeOrganizationMemberUser =
    createAdminServiceRemoveOrganizationMemberUser();
  const updateOrg = createAdminServiceUpdateOrganization();

  async function handleRemove(email: string) {
    try {
      await $removeOrganizationMemberUser.mutateAsync({
        organization: organization,
        email: email,
        // Uncomment if `keepProjectRoles` is needed
        // See: https://github.com/rilldata/rill/pull/2231
        // params: {
        //   keepProjectRoles: false,
        // },
      });

      await queryClient.invalidateQueries({
        queryKey:
          getAdminServiceListOrganizationMemberUsersQueryKey(organization),
      });

      await queryClient.invalidateQueries({
        queryKey: getAdminServiceListOrganizationInvitesQueryKey(organization),
      });

      await queryClient.invalidateQueries({
        queryKey: getAdminServiceListUsergroupMemberUsersQueryKey(
          organization,
          "autogroup:users",
        ),
      });

      await queryClient.invalidateQueries({
        queryKey: getAdminServiceListUsergroupMemberUsersQueryKey(
          organization,
          "autogroup:members",
        ),
      });

      await queryClient.invalidateQueries({
        queryKey: getAdminServiceListUsergroupMemberUsersQueryKey(
          organization,
          "autogroup:guests",
        ),
      });

      eventBus.emit("notification", {
        message: "User removed from organization",
      });
    } catch (error) {
      console.error("Error removing user from organization", error);
      eventBus.emit("notification", {
        message: "Error removing user from organization",
        type: "error",
      });
    }
  }

  async function handleAssignAsBillingContact() {
    try {
      await $updateOrg.mutateAsync({
        name: organization,
        data: {
          billingEmail: email,
        },
      });

      eventBus.emit("notification", {
        message: `Successfully assigned ${name} as the billing contact`,
      });
    } catch (error) {
      console.error("Error assigning user as billing contact", error);
      eventBus.emit("notification", {
        message:
          "Error: Unable to assign billing contact. Please try again or contact support if the issue persists.",
        type: "error",
      });
    }
  }

  $: showAssignAsBillingContact = currentUserCanManageOrg && isAdminUser;
  $: showDropdown = !isCurrentUser || showAssignAsBillingContact;
</script>

{#if showDropdown}
  <DropdownMenu.Root bind:open={isDropdownOpen}>
    <DropdownMenu.Trigger class="flex-none">
      <IconButton rounded active={isDropdownOpen}>
        <ThreeDot size="16px" />
      </IconButton>
    </DropdownMenu.Trigger>
    <DropdownMenu.Content align="start">
      {#if !isCurrentUser}
        <DropdownMenu.Item
          class="font-normal flex items-center"
          type="destructive"
          on:click={() => {
            isRemoveConfirmOpen = true;
          }}
        >
          <Trash2Icon size="12px" />
          <span class="ml-2">Remove</span>
        </DropdownMenu.Item>
      {/if}
      {#if showAssignAsBillingContact}
        <DropdownMenu.Item
          class="font-normal flex items-center"
          on:click={handleAssignAsBillingContact}
        >
          <span class="ml-2">Assign as billing contact</span>
        </DropdownMenu.Item>
      {/if}
    </DropdownMenu.Content>
  </DropdownMenu.Root>
{/if}

<RemoveUserFromOrgConfirmDialog
  bind:open={isRemoveConfirmOpen}
  {email}
  onRemove={handleRemove}
/>
