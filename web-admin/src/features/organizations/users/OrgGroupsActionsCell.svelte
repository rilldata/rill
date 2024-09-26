<script lang="ts">
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import { Trash2Icon, UserCogIcon } from "lucide-svelte";
  import DeleteUserGroupConfirmDialog from "./DeleteUserGroupConfirmDialog.svelte";

  export let name: string;
  export let onDelete: (deletedGroupName: string) => void;
  export let onSetRole: (groupName: string, role: string) => void;

  let isDropdownOpen = false;
  let isDeleteConfirmOpen = false;

  function handleSetRole(role: string) {
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
    <DropdownMenu.Sub>
      <DropdownMenu.SubTrigger class="font-normal flex items-center">
        <UserCogIcon size="12px" />
        <span class="ml-2">Set a role</span>
      </DropdownMenu.SubTrigger>
      <DropdownMenu.SubContent>
        <DropdownMenu.Item
          class="font-normal flex items-center"
          on:click={() => {
            handleSetRole("admin");
          }}
        >
          <span>Admin</span>
        </DropdownMenu.Item>
        <DropdownMenu.Item
          class="font-normal flex items-center"
          on:click={() => {
            handleSetRole("viewer");
          }}
        >
          <span>Viewer</span>
        </DropdownMenu.Item>
        <DropdownMenu.Item
          class="font-normal flex items-center"
          on:click={() => {
            handleSetRole("collaborator");
          }}
        >
          <span>Collaborator</span>
        </DropdownMenu.Item>
      </DropdownMenu.SubContent>
    </DropdownMenu.Sub>
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
