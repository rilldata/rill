<script lang="ts">
  import { beforeNavigate } from "$app/navigation";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import CaretUpIcon from "@rilldata/web-common/components/icons/CaretUpIcon.svelte";
  import { Search } from "@rilldata/web-common/components/search";

  export let availableTags: string[] = [];
  export let selectedTags: string[] = [];
  export let onTagsChange: (tags: string[]) => void;
  export let searchText = "";

  let tagsOpen = false;

  beforeNavigate(() => {
    searchText = "";
  });

  function toggleTag(tag: string) {
    onTagsChange(
      selectedTags.includes(tag)
        ? selectedTags.filter((t) => t !== tag)
        : [...selectedTags, tag],
    );
  }

  $: tagsLabel =
    selectedTags.length === 0
      ? "All tags"
      : selectedTags.length === 1
        ? selectedTags[0]
        : `${selectedTags[0]}, +${selectedTags.length - 1} other${selectedTags.length > 2 ? "s" : ""}`;
</script>

<div class="flex flex-row items-center gap-x-2">
  {#if availableTags.length > 0}
    <DropdownMenu.Root bind:open={tagsOpen}>
      <DropdownMenu.Trigger
        class="min-w-fit min-h-9 flex flex-row gap-1 items-center rounded-sm border bg-input {tagsOpen
          ? 'bg-gray-200'
          : 'hover:bg-surface-hover'} px-2 py-1"
      >
        <span class="text-fg-secondary font-medium text-sm">{tagsLabel}</span>
        {#if tagsOpen}
          <CaretUpIcon size="12px" />
        {:else}
          <CaretDownIcon size="12px" />
        {/if}
      </DropdownMenu.Trigger>
      <DropdownMenu.Content align="start" class="w-48 max-h-72 overflow-y-auto">
        {#each availableTags as tag (tag)}
          <DropdownMenu.CheckboxItem
            checked={selectedTags.includes(tag)}
            onCheckedChange={() => toggleTag(tag)}
          >
            {tag}
          </DropdownMenu.CheckboxItem>
        {/each}
      </DropdownMenu.Content>
    </DropdownMenu.Root>
  {/if}

  <div class="flex-1 min-w-0">
    <Search
      placeholder="Search"
      autofocus={false}
      bind:value={searchText}
      rounded="lg"
    />
  </div>
</div>
