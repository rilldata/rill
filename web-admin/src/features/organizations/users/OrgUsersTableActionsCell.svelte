<script lang="ts">
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import { OrgUserRoles } from "@rilldata/web-common/features/users/roles.ts";
  import { Trash2Icon } from "lucide-svelte";
  import RemoveUserFromOrgConfirmDialog from "./RemoveUserFromOrgConfirmDialog.svelte";
  import {
    createAdminServiceRemoveOrganizationMemberUser,
    getAdminServiceListOrganizationInvitesQueryKey,
    getAdminServiceListOrganizationMemberUsersQueryKey,
    getAdminServiceListUsergroupMemberUsersQueryKey,
  } from "@rilldata/web-admin/client";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { page } from "$app/stores";

  export let email: string;
  export let role: string;
  export let isCurrentUser: boolean;
  export let currentUserRole: string;
  export let isBillingContact: boolean;
  // Changing billing contact is not an action for this user. So handle it upstream
  // This also avoids rendering the modal per row.
  export let onAttemptRemoveBillingContactUser: () => void;
  export let onConvertToMember: (user: string) => void;

  let isDropdownOpen = false;
  let isRemoveConfirmOpen = false;

  $: organization = $page.params.organization;
  $: isAdmin = currentUserRole === OrgUserRoles.Admin;
  $: isEditor = currentUserRole === OrgUserRoles.Editor;
  $: canManageUser = true;
  // $: canManageUser =
  //   !isCurrentUser &&
  //   (isAdmin ||
  //     (isEditor &&
  //       (role === OrgUserRoles.Editor ||
  //         role === OrgUserRoles.Viewer ||
  //         role === OrgUserRoles.Guest)));

  const queryClient = useQueryClient();
  const removeOrganizationMemberUser =
    createAdminServiceRemoveOrganizationMemberUser();

  function onRemoveClick() {
    if (isBillingContact) {
      // If the user is a billing contact we cannot remove without update contact to a different user 1st.
      onAttemptRemoveBillingContactUser();
    } else {
      // Else show the confirmation for remove
      isRemoveConfirmOpen = true;
    }
  }

  async function handleRemove(email: string) {
    try {
      await $removeOrganizationMemberUser.mutateAsync({
        organization: organization,
        email: email,
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
</script>

{#if canManageUser}
  <DropdownMenu.Root bind:open={isDropdownOpen}>
    <DropdownMenu.Trigger class="flex-none">
      <IconButton rounded active={isDropdownOpen}>
        <ThreeDot size="16px" />
      </IconButton>
    </DropdownMenu.Trigger>
    <DropdownMenu.Content align="start">
      <DropdownMenu.Item
        class="font-normal flex items-center"
        type="destructive"
        on:click={onRemoveClick}
      >
        <Trash2Icon size="12px" />
        <span class="ml-2">Remove</span>
      </DropdownMenu.Item>
      {#if role === OrgUserRoles.Guest}
        <DropdownMenu.Item
          class="font-normal flex items-center"
          on:click={() => onConvertToMember(email)}
        >
          Convert to member
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
