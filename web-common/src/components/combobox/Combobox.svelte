<script lang="ts">
  import AvatarListItem from "@rilldata/web-admin/features/organizations/users/AvatarListItem.svelte";
  import { Combobox } from "bits-ui";
  import type { Selected } from "bits-ui";
  import { Check } from "lucide-svelte";

  export let options: {
    value: string;
    label: string;
    name: string;
  }[] = [];
  export let inputValue = "";
  export let name = "";
  export let placeholder = "Search";
  export let label = "";
  export let id = "";
  export let onSelectedChange: (value: Selected<string> | undefined) => void;
  export let emptyText = "No results found";

  function handleSelectedChange(selected: Selected<string> | undefined) {
    onSelectedChange(selected);
  }

  $: filteredItems = inputValue
    ? options.filter((fruit) => fruit.value.includes(inputValue.toLowerCase()))
    : options;
</script>

<div class="flex flex-col gap-y-1">
  {#if label}
    <label for={id} class="line-clamp-1 text-sm font-medium text-gray-800">
      {label}
    </label>
  {/if}

  <Combobox.Root
    items={filteredItems}
    bind:inputValue
    onSelectedChange={handleSelectedChange}
  >
    <Combobox.Input
      class="flex justify-center items-center pl-2 w-full border border-gray-300 rounded-[2px] cursor-pointer min-h-8 h-fit focus-within:border-primary-500 focus-within:ring-2 focus-within:ring-primary-100 focus:outline-none"
      {placeholder}
      aria-label={placeholder}
    />

    <!-- NOTE: 52px * 4 for 208px to show scroller -->
    <Combobox.Content
      class="w-full rounded-sm border border-muted bg-surface p-[6px] shadow-md outline-none max-h-[208px] overflow-y-auto"
      sideOffset={8}
    >
      {#each filteredItems as item (item.value)}
        <Combobox.Item
          class="flex h-[52px] w-full select-none items-center rounded px-4 py-2 text-sm outline-none transition-all duration-75 data-[highlighted]:bg-slate-100"
          value={item.value}
          label={item.label}
        >
          <AvatarListItem name={item.name} email={item.value} />
          <Combobox.ItemIndicator class="ml-auto" asChild={false}>
            <Check size="16px" />
          </Combobox.ItemIndicator>
        </Combobox.Item>
      {:else}
        <span class="block px-5 py-4 text-xs text-gray-500">
          {emptyText}
        </span>
      {/each}
    </Combobox.Content>
    <Combobox.HiddenInput {name} />
  </Combobox.Root>
</div>
