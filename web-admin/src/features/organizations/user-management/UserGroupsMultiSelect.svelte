<script lang="ts">
  import { getOrgUsergroupsInfinite } from "@rilldata/web-admin/features/organizations/user-management/selectors.ts";
  import * as Dropdown from "@rilldata/web-common/components/dropdown-menu";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import CaretUpIcon from "@rilldata/web-common/components/icons/CaretUpIcon.svelte";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  // A dropdown of the organization's user groups with a checkbox per group.
  // Managed groups (autogroup:*) are never listed since their membership cannot be edited.
  export let organization: string;
  export let selected: string[] = [];
  export let disabled = false;
  export let id: string | undefined = undefined;

  let isOpen = false;
  let searchText = "";

  $: groupsQuery = getOrgUsergroupsInfinite(organization);

  // Groups are few enough per org that we load every page up front rather than paginate a dropdown.
  $: if ($groupsQuery.hasNextPage && !$groupsQuery.isFetchingNextPage) {
    void $groupsQuery.fetchNextPage();
  }

  $: groupNames =
    $groupsQuery.data?.pages
      .flatMap((page) => page.members ?? [])
      .filter((group) => !group.groupManaged)
      .map((group) => group.groupName ?? "")
      .filter(Boolean) ?? [];

  $: filteredGroupNames = searchText
    ? groupNames.filter((name) =>
        name.toLowerCase().includes(searchText.toLowerCase()),
      )
    : groupNames;

  $: label = (() => {
    if (selected.length === 0) return m.users_select_groups();
    if (selected.length === 1) return selected[0];
    return m.users_group_count({ count: selected.length });
  })();

  function toggle(name: string) {
    if (selected.includes(name)) {
      selected = selected.filter((n) => n !== name);
    } else {
      selected = [...selected, name];
    }
    // Keep the menu open so several groups can be picked in one go
    isOpen = true;
  }
</script>

{#if $groupsQuery.isLoading}
  <DelayedSpinner isLoading size="1rem" />
{:else if groupNames.length === 0}
  <div class="text-xs text-fg-secondary">{m.users_no_groups_available()}</div>
{:else}
  <Dropdown.Root bind:open={isOpen}>
    <Dropdown.Trigger
      {id}
      {disabled}
      class="min-w-[260px] min-h-[32px] flex flex-row justify-between gap-1 items-center rounded-sm border border-gray-300 bg-surface-background text-sm px-3 {isOpen
        ? 'bg-gray-200'
        : 'hover:bg-surface-hover'}"
    >
      <span class="truncate" title={selected.join(", ")}>{label}</span>
      {#if isOpen}
        <CaretUpIcon size="12px" />
      {:else}
        <CaretDownIcon size="12px" />
      {/if}
    </Dropdown.Trigger>
    <Dropdown.Content align="start" class="w-[260px]">
      {#if groupNames.length > 8}
        <input
          type="text"
          class="w-full mb-1 px-2 py-1 text-xs border border-gray-300 rounded-sm bg-surface-background focus:outline-none focus:border-primary-500"
          placeholder={m.users_search_groups()}
          bind:value={searchText}
          onkeydown={(e) => e.stopPropagation()}
        />
      {/if}
      <div class="max-h-[240px] overflow-y-auto">
        {#each filteredGroupNames as name (name)}
          <Dropdown.CheckboxItem
            closeOnSelect={false}
            class="font-normal flex items-center overflow-hidden"
            checked={selected.includes(name)}
            onCheckedChange={() => toggle(name)}
          >
            <span class="truncate w-full" title={name}>{name}</span>
          </Dropdown.CheckboxItem>
        {:else}
          <div class="px-2 py-1 text-xs text-fg-secondary">
            {m.common_no_results()}
          </div>
        {/each}
      </div>
    </Dropdown.Content>
  </Dropdown.Root>
{/if}
