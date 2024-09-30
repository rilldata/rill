<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { capitalize } from "@rilldata/web-common/components/table/utils";
  import { UserMinusIcon } from "lucide-svelte";

  export let name: string;
  export let role: string | undefined = undefined;
  export let onAddRole: (groupName: string, role: string) => void;
  export let onSetRole: (groupName: string, role: string) => void;
  export let onRevokeRole: (groupName: string) => void;

  let isDropdownOpen = false;

  function handleAddRole(role: string) {
    onAddRole(name, role);
  }

  function handleUpdateRole(role: string) {
    onSetRole(name, role);
  }

  function handleRevokeRole() {
    onRevokeRole(name);
  }
</script>

<!-- all-users — "cannot add role for all-users group" -->
{#if name !== "all-users"}
  <DropdownMenu.Root bind:open={isDropdownOpen}>
    <DropdownMenu.Trigger class="flex-none">
      {role ? `Org ${capitalize(role)}` : "-"}
    </DropdownMenu.Trigger>
    <DropdownMenu.Content align="start">
      <DropdownMenu.CheckboxItem
        class="font-normal flex items-center"
        checked={role === "admin"}
        on:click={() => {
          if (role) {
            handleUpdateRole("admin");
          } else {
            handleAddRole("admin");
          }
        }}
      >
        <span>Admin</span>
      </DropdownMenu.CheckboxItem>
      <DropdownMenu.CheckboxItem
        class="font-normal flex items-center"
        checked={role === "viewer"}
        on:click={() => {
          if (role) {
            handleUpdateRole("viewer");
          } else {
            handleAddRole("viewer");
          }
        }}
      >
        <span>Viewer</span>
      </DropdownMenu.CheckboxItem>
      <DropdownMenu.CheckboxItem
        class="font-normal flex items-center"
        checked={role === "collaborator"}
        on:click={() => {
          if (role) {
            handleUpdateRole("collaborator");
          } else {
            handleAddRole("collaborator");
          }
        }}
      >
        <span>Collaborator</span>
      </DropdownMenu.CheckboxItem>
      {#if role}
        <DropdownMenu.Separator />
        <DropdownMenu.Item
          class="font-normal flex items-center"
          on:click={handleRevokeRole}
        >
          <UserMinusIcon size="12px" />
          <span class="ml-2">Revoke</span>
        </DropdownMenu.Item>
      {/if}
    </DropdownMenu.Content>
  </DropdownMenu.Root>
{:else}
  <!-- TODO: add tooltip for all-users -->
  <!-- "cannot add role for all-users group" -->
  <span>-</span>
{/if}
