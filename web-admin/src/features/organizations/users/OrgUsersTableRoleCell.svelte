<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";

  export let email: string;
  export let role: string;
  export let isCurrentUser: boolean;
  export let onSetRole: (email: string, role: string) => void;

  let isDropdownOpen = false;

  function handleUpdateRole(role: string) {
    onSetRole(email, role);
  }
</script>

{#if !isCurrentUser}
  <DropdownMenu.Root bind:open={isDropdownOpen}>
    <DropdownMenu.Trigger class="flex-none">
      {role ? `Org ${role}` : "-"}
    </DropdownMenu.Trigger>
    <DropdownMenu.Content align="start">
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
    </DropdownMenu.Content>
  </DropdownMenu.Root>
{:else}
  <span>Org {role}</span>
{/if}
