<script lang="ts">
  import * as m from "@rilldata/web-common/paraglide/messages.js";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import CaretUpIcon from "@rilldata/web-common/components/icons/CaretUpIcon.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";

  export let filterSelection: "all" | "members" | "guests" | "pending" = "all";
  export let showMembers = true;
  export let roleFilter: "all" | "admin" | "editor" | "viewer" = "all";
  export let showRoleFilter = true;

  let isDropdownOpen = false;
  let isRoleDropdownOpen = false;

  $: roleOptions = [
    { value: "all" as const, label: m.users_filter_all_roles() },
    { value: "admin" as const, label: m.users_filter_admins() },
    { value: "editor" as const, label: m.users_filter_editors() },
    { value: "viewer" as const, label: m.users_filter_viewers() },
  ];

  $: roleFilterLabel =
    roleOptions.find((opt) => opt.value === roleFilter)?.label ?? m.users_filter_all_roles();
</script>

<DropdownMenu.Root bind:open={isDropdownOpen}>
  <DropdownMenu.Trigger
    class="min-w-[140px] flex flex-row justify-between gap-1 items-center rounded-sm border bg-input px-2 py-1"
  >
    <span class="capitalize"
      >{filterSelection === "all" ? m.users_filter_all_users() : filterSelection === "members" ? m.users_filter_members() : filterSelection === "pending" ? m.users_filter_pending_invites() : filterSelection}</span
    >
    {#if isDropdownOpen}
      <CaretUpIcon size="12px" />
    {:else}
      <CaretDownIcon size="12px" />
    {/if}
  </DropdownMenu.Trigger>
  <DropdownMenu.Content align="start" class="min-w-[140px]">
    <DropdownMenu.CheckboxItem
      class="font-normal flex items-center"
      checked={filterSelection === "all"}
      onclick={() => {
        filterSelection = "all";
      }}
    >
      <span>{m.users_filter_all()}</span>
    </DropdownMenu.CheckboxItem>
    {#if showMembers}
      <DropdownMenu.CheckboxItem
        class="font-normal flex items-center"
        checked={filterSelection === "members"}
        onclick={() => {
          filterSelection = "members";
        }}
      >
        <span>{m.users_filter_members()}</span>
      </DropdownMenu.CheckboxItem>
    {/if}
    <DropdownMenu.CheckboxItem
      class="font-normal flex items-center"
      checked={filterSelection === "pending"}
      onclick={() => {
        filterSelection = "pending";
      }}
    >
      <span>{m.users_filter_pending_invites()}</span>
    </DropdownMenu.CheckboxItem>
  </DropdownMenu.Content>
</DropdownMenu.Root>

{#if showRoleFilter}
  <DropdownMenu.Root bind:open={isRoleDropdownOpen}>
    <DropdownMenu.Trigger
      class="min-w-[120px] flex flex-row justify-between gap-1 items-center rounded-sm border bg-input px-2 py-1"
    >
      <span>{roleFilterLabel}</span>
      {#if isRoleDropdownOpen}
        <CaretUpIcon size="12px" />
      {:else}
        <CaretDownIcon size="12px" />
      {/if}
    </DropdownMenu.Trigger>
    <DropdownMenu.Content align="start" class="min-w-[120px]">
      {#each roleOptions as option}
        <DropdownMenu.CheckboxItem
          class="font-normal flex items-center"
          checked={roleFilter === option.value}
          onclick={() => {
            roleFilter = option.value;
          }}
        >
          <span>{option.label}</span>
        </DropdownMenu.CheckboxItem>
      {/each}
    </DropdownMenu.Content>
  </DropdownMenu.Root>
{/if}
