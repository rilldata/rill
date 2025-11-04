<script lang="ts">
  import {
    canManageOrgUser,
    invalidateAfterUserDelete,
  } from "@rilldata/web-admin/features/organizations/user-management/utils.ts";
  import IconButton from "web-common/src/components/button/IconButton.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import { OrgUserRoles } from "@rilldata/web-common/features/users/roles.ts";
  import { Trash2Icon } from "lucide-svelte";
  import RemoveUserFromOrgConfirmDialog from "@rilldata/web-admin/features/organizations/user-management/dialogs/RemoveUserFromOrgConfirmDialog.svelte";
  import {
    createAdminServiceRemoveOrganizationMemberUser,
    type V1OrganizationPermissions,
  } from "@rilldata/web-admin/client";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";
  import { page } from "$app/stores";

  export let email: string;
  export let role: string;
  export let isCurrentUser: boolean;
  export let organizationPermissions: V1OrganizationPermissions;
  export let isBillingContact: boolean;
  // Changing billing contact is not an action for this user. So handle it upstream
  // This also avoids rendering the modal per row.
  export let onAttemptRemoveBillingContactUser: () => void;
  export let onConvertToMember: () => void;

  let isDropdownOpen = false;
  let isRemoveConfirmOpen = false;

  $: organization = $page.params.organization;
  $: isGuest = role === OrgUserRoles.Guest;
  $: canManageUser =
    // TODO: backend doesnt restrict removing oneself, revisit this UI check.
    !isCurrentUser && canManageOrgUser(organizationPermissions, role);

  const queryClient = useQueryClient();
  const removeOrganizationMemberUser =
    createAdminServiceRemoveOrganizationMemberUser();

  function onRemoveClick() {
    if (isBillingContact) {
      // If the user is a billing contact we cannot remove without update contact to a different user 1st.
      onAttemptRemoveBillingContactUser();
    } else if (isGuest) {
      void handleRemove(email);
    } else {
      // Else show the confirmation for remove
      isRemoveConfirmOpen = true;
    }
  }

  async function handleRemove(email: string) {
    try {
      await $removeOrganizationMemberUser.mutateAsync({
        org: organization,
        email,
      });

      await invalidateAfterUserDelete(queryClient, organization);

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
      {#if role === OrgUserRoles.Guest}
        <DropdownMenu.Item
          class="font-normal flex items-center"
          on:click={onConvertToMember}
        >
          Convert to member
        </DropdownMenu.Item>
      {/if}
      <DropdownMenu.Item
        class="font-normal flex items-center"
        type="destructive"
        on:click={onRemoveClick}
      >
        <Trash2Icon size="12px" />
        <span class="ml-2">Remove</span>
      </DropdownMenu.Item>
    </DropdownMenu.Content>
  </DropdownMenu.Root>
{/if}

<RemoveUserFromOrgConfirmDialog
  bind:open={isRemoveConfirmOpen}
  {email}
  onRemove={handleRemove}
/>
