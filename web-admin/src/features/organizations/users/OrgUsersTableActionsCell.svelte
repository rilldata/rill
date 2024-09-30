<script lang="ts">
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import { Trash2Icon, UserCogIcon } from "lucide-svelte";
  import RemoveUserFromOrgConfirmDialog from "./RemoveUserFromOrgConfirmDialog.svelte";
  import type { V1MemberUsergroup } from "@rilldata/web-admin/client";

  export let email: string;
  export let pendingAcceptance: boolean;
  export let isCurrentUser: boolean;
  export let userGroups: V1MemberUsergroup[];
  export let onRemove: (email: string) => void;
  export let onAddUsergroupMemberUser: (
    email: string,
    usergroup: string,
  ) => void;

  let isDropdownOpen = false;
  let isRemoveConfirmOpen = false;

  function handleRemove() {
    onRemove(email);
  }

  function handleAddUsergroupMemberUser(usergroup: string) {
    onAddUsergroupMemberUser(email, usergroup);
  }
</script>

<DropdownMenu.Root bind:open={isDropdownOpen}>
  <DropdownMenu.Trigger class="flex-none">
    <IconButton rounded active={isDropdownOpen}>
      <ThreeDot size="16px" />
    </IconButton>
  </DropdownMenu.Trigger>
  <DropdownMenu.Content align="start">
    {#if !pendingAcceptance && userGroups.length > 0}
      <DropdownMenu.Sub>
        <DropdownMenu.SubTrigger class="font-normal flex items-center">
          <UserCogIcon size="12px" />
          <span class="ml-2">Add to group</span>
        </DropdownMenu.SubTrigger>
        <!-- TODO: if user is already in group, disable the option -->
        <!-- Otherwise, "user is already a member of the usergroup" -->
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
  </DropdownMenu.Content>
</DropdownMenu.Root>

<RemoveUserFromOrgConfirmDialog
  bind:open={isRemoveConfirmOpen}
  {email}
  onRemove={handleRemove}
/>
