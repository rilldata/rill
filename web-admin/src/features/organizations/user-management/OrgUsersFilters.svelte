<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import CaretUpIcon from "@rilldata/web-common/components/icons/CaretUpIcon.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";

  export let filterSelection: "all" | "members" | "guests" | "pending" = "all";
  export let showMembers = true;

  let isDropdownOpen = false;
</script>

<DropdownMenu.Root bind:open={isDropdownOpen}>
  <DropdownMenu.Trigger
    class="min-w-[210px] flex flex-row justify-between gap-1 items-center rounded-sm border border-gray-300 {isDropdownOpen
      ? 'bg-slate-200'
      : 'hover:bg-slate-100'} px-2 py-1"
  >
    <span class="capitalize"
      >{filterSelection === "all" ? "All users" : filterSelection}</span
    >
    {#if isDropdownOpen}
      <CaretUpIcon size="12px" />
    {:else}
      <CaretDownIcon size="12px" />
    {/if}
  </DropdownMenu.Trigger>
  <DropdownMenu.Content align="start" class="w-[210px]">
    <DropdownMenu.CheckboxItem
      class="font-normal flex items-center"
      checked={filterSelection === "all"}
      on:click={() => {
        filterSelection = "all";
      }}
    >
      <span>All</span>
    </DropdownMenu.CheckboxItem>
    {#if showMembers}
      <DropdownMenu.CheckboxItem
        class="font-normal flex items-center"
        checked={filterSelection === "members"}
        on:click={() => {
          filterSelection = "members";
        }}
      >
        <span>Members</span>
      </DropdownMenu.CheckboxItem>
    {/if}
    <DropdownMenu.CheckboxItem
      class="font-normal flex items-center"
      checked={filterSelection === "pending"}
      on:click={() => {
        filterSelection = "pending";
      }}
    >
      <span>Pending invites</span>
    </DropdownMenu.CheckboxItem>
  </DropdownMenu.Content>
</DropdownMenu.Root>
