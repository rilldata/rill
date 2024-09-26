<script lang="ts">
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import { Trash2Icon, UserCogIcon } from "lucide-svelte";
  import DeleteUserGroupConfirmDialog from "./DeleteUserGroupConfirmDialog.svelte";

  export let name: string;
  export let role: string | undefined = undefined;
  export let onDelete: (deletedGroupName: string) => void;
  export let onAddRole: (groupName: string, role: string) => void;
  export let onSetRole: (groupName: string, role: string) => void;

  let isDropdownOpen = false;
  let isDeleteConfirmOpen = false;

  function handleAssignRole(role: string) {
    onAddRole(name, role);
  }

  function handleUpdateRole(role: string) {
    onSetRole(name, role);
  }
</script>

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
          <span class="ml-2">Update role</span>
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
    {:else}
      <DropdownMenu.Sub>
        <DropdownMenu.SubTrigger class="font-normal flex items-center">
          <UserCogIcon size="12px" />
          <span class="ml-2">Assign a role</span>
        </DropdownMenu.SubTrigger>
        <DropdownMenu.SubContent>
          <svelte:component
            this={role !== undefined
              ? DropdownMenu.Item
              : DropdownMenu.CheckboxItem}
            class="font-normal flex items-center"
            on:click={() => {
              role !== undefined
                ? handleUpdateRole("admin")
                : handleAssignRole("admin");
            }}
          >
            <span>Admin</span>
          </svelte:component>
          <svelte:component
            this={role !== undefined
              ? DropdownMenu.Item
              : DropdownMenu.CheckboxItem}
            class="font-normal flex items-center"
            on:click={() => {
              role !== undefined
                ? handleUpdateRole("viewer")
                : handleAssignRole("viewer");
            }}
          >
            <span>Viewer</span>
          </svelte:component>
          <svelte:component
            this={role !== undefined
              ? DropdownMenu.Item
              : DropdownMenu.CheckboxItem}
            class="font-normal flex items-center"
            on:click={() => {
              role !== undefined
                ? handleUpdateRole("collaborator")
                : handleAssignRole("collaborator");
            }}
          >
            <span>Collaborator</span>
          </svelte:component>
        </DropdownMenu.SubContent>
      </DropdownMenu.Sub>
    {/if}

    <DropdownMenu.Item
      class="font-normal flex items-center"
      type="destructive"
      on:click={() => {
        isDeleteConfirmOpen = true;
      }}
    >
      <Trash2Icon size="12px" />
      <span class="ml-2">Delete</span>
    </DropdownMenu.Item>
  </DropdownMenu.Content>
</DropdownMenu.Root>

<DeleteUserGroupConfirmDialog
  bind:open={isDeleteConfirmOpen}
  {name}
  {onDelete}
/>
