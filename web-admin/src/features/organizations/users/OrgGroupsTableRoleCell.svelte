<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import CaretUpIcon from "@rilldata/web-common/components/icons/CaretUpIcon.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";

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
      {#if role}
        <DropdownMenu.Separator />
        <DropdownMenu.Item
          class="font-normal flex items-center"
          on:click={handleRevokeRole}
        >
          <span class="ml-6">Remove</span>
        </DropdownMenu.Item>
      {/if}
    </DropdownMenu.Content>
  </DropdownMenu.Root>
{:else}
  <Tooltip location="top" alignment="start" distance={8}>
    <div class="w-18 rounded-sm px-2 py-1">
      <span class="cursor-help">-</span>
    </div>
    <TooltipContent maxWidth="400px" slot="tooltip-content">
      Cannot add role for all-users group
    </TooltipContent>
  </Tooltip>
{/if}
