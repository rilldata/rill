<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import CaretUpIcon from "@rilldata/web-common/components/icons/CaretUpIcon.svelte";

  /** All tag values available to choose from. Dropdown renders nothing if empty. */
  export let tags: string[];
  /** Two-way bindable selected tags */
  export let selectedTags: string[] = [];
  /** Keep the menu open after each selection (matches the kind/status dropdowns in admin) */
  export let closeOnSelect = true;

  let open = false;

  function toggle(tag: string) {
    if (selectedTags.includes(tag)) {
      selectedTags = selectedTags.filter((t) => t !== tag);
    } else {
      selectedTags = [...selectedTags, tag];
    }
  }
</script>

{#if tags.length > 0}
  <DropdownMenu.Root bind:open>
    <DropdownMenu.Trigger
      class="min-w-fit min-h-9 flex flex-row gap-1 items-center rounded-sm border bg-input {open
        ? 'bg-gray-200'
        : 'hover:bg-surface-hover'} px-2 py-1"
    >
      <span class="text-fg-secondary font-medium">
        {#if selectedTags.length === 0}
          All tags
        {:else if selectedTags.length === 1}
          {selectedTags[0]}
        {:else}
          {selectedTags[0]}, +{selectedTags.length - 1} other{selectedTags.length >
          2
            ? "s"
            : ""}
        {/if}
      </span>
      {#if open}
        <CaretUpIcon size="12px" />
      {:else}
        <CaretDownIcon size="12px" />
      {/if}
    </DropdownMenu.Trigger>
    <DropdownMenu.Content align="start" class="w-48 max-h-72 overflow-auto">
      {#each tags as tag}
        <DropdownMenu.CheckboxItem
          {closeOnSelect}
          checked={selectedTags.includes(tag)}
          onCheckedChange={() => toggle(tag)}
        >
          {tag}
        </DropdownMenu.CheckboxItem>
      {/each}
    </DropdownMenu.Content>
  </DropdownMenu.Root>
{/if}
