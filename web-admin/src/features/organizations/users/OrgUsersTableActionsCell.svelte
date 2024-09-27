<script lang="ts">
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import { Trash2Icon, UserCogIcon } from "lucide-svelte";
  import RemoveUserFromOrgConfirmDialog from "./RemoveUserFromOrgConfirmDialog.svelte";
  import type { V1MemberUsergroup } from "@rilldata/web-admin/client";

  export let email: string;
  export let role: string;
  export let pendingAcceptance: boolean;
  export let isCurrentUser: boolean;
  export let userGroups: V1MemberUsergroup[];
  export let onRemove: (email: string) => void;
  export let onSetRole: (email: string, role: string) => void;
  export let onAddUsergroupMemberUser: (
    email: string,
    usergroup: string,
  ) => void;

  let isDropdownOpen = false;
  let isRemoveConfirmOpen = false;

  function handleRemove() {
    onRemove(email);
  }

  function handleUpdateRole(role: string) {
    onSetRole(email, role);
  }

  function handleAddUsergroupMemberUser(usergroup: string) {
    onAddUsergroupMemberUser(email, usergroup);
  }
</script>

{#if !isCurrentUser}
  <DropdownMenu.Root bind:open={isDropdownOpen}>
    <DropdownMenu.Trigger class="flex-none">
      <IconButton rounded active={isDropdownOpen}>
        <ThreeDot size="16px" />
      </IconButton>
    </DropdownMenu.Trigger>
    <DropdownMenu.Content align="start">
      {#if role}
        <DropdownMenu.Sub>
          <DropdownMenu.SubTrigger class="font-normal flex items-center">
            <UserCogIcon size="12px" />
            <span class="ml-2">Change role</span>
          </DropdownMenu.SubTrigger>
          <DropdownMenu.SubContent>
            <DropdownMenu.CheckboxItem
              class="font-normal flex items-center"
              checked={role === "admin"}
              on:click={() => {
                handleUpdateRole("admin");
              }}
            >
              <span>Admin</span>
            </DropdownMenu.CheckboxItem>
            <DropdownMenu.CheckboxItem
              class="font-normal flex items-center"
              checked={role === "viewer"}
              on:click={() => {
                handleUpdateRole("viewer");
              }}
            >
              <span>Viewer</span>
            </DropdownMenu.CheckboxItem>
            <DropdownMenu.CheckboxItem
              class="font-normal flex items-center"
              checked={role === "collaborator"}
              on:click={() => {
                handleUpdateRole("collaborator");
              }}
            >
              <span>Collaborator</span>
            </DropdownMenu.CheckboxItem>
          </DropdownMenu.SubContent>
        </DropdownMenu.Sub>
        {#if !pendingAcceptance && userGroups.length > 0}
          <DropdownMenu.Sub>
            <DropdownMenu.SubTrigger class="font-normal flex items-center">
              <UserCogIcon size="12px" />
              <span class="ml-2">Add to user group</span>
            </DropdownMenu.SubTrigger>
            <DropdownMenu.SubContent>
              {#each userGroups as usergroup}
                <DropdownMenu.Item
                  class="font-normal flex items-center"
                  on:click={() => {
                    handleAddUsergroupMemberUser(usergroup.groupName);
                  }}
                >
                  <span>{usergroup.groupName}</span>
                </DropdownMenu.Item>
              {/each}
            </DropdownMenu.SubContent>
          </DropdownMenu.Sub>
        {/if}
      {/if}

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
    </DropdownMenu.Content>
  </DropdownMenu.Root>
{/if}

<RemoveUserFromOrgConfirmDialog
  bind:open={isRemoveConfirmOpen}
  {email}
  onRemove={handleRemove}
/>
