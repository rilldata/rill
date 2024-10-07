<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import CaretUpIcon from "@rilldata/web-common/components/icons/CaretUpIcon.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";

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
    <DropdownMenu.Trigger
      class="w-18 flex flex-row gap-1 items-center rounded-sm {isDropdownOpen
        ? 'bg-slate-200'
        : 'hover:bg-slate-100'} px-2 py-1"
    >
      {role ? `Org ${role}` : "-"}
      {#if isDropdownOpen}
        <CaretUpIcon size="12px" />
      {:else}
        <CaretDownIcon size="12px" />
      {/if}
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
    </DropdownMenu.Content>
  </DropdownMenu.Root>
{:else}
  <div class="w-18 flex flex-row gap-1 items-center rounded-sm px-2 py-1">
    <span>Org {role}</span>
  </div>
{/if}
