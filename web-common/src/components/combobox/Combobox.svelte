<script lang="ts">
  import AvatarListItem from "@rilldata/web-admin/features/organizations/users/AvatarListItem.svelte";
  import { Combobox } from "bits-ui";
  import type { Selected } from "bits-ui";

  export let options: {
    value: string;
    label: string;
  }[] = [];
  export let inputValue = "";
  export let placeholder = "Search";
  export let onSelectedChange: (value: Selected<string> | undefined) => void;
  export let getMetadata: (
    value: string,
  ) => { name: string; photoUrl?: string } | undefined = () => undefined;

  function handleSelectedChange(selected: Selected<string> | undefined) {
    onSelectedChange(selected);
  }

  $: filteredItems = inputValue
    ? options.filter((option) =>
        option.value.toLowerCase().includes(inputValue.toLowerCase()),
      )
    : options;

  $: console.log(filteredItems);
</script>

<Combobox.Root
  items={filteredItems}
  bind:inputValue
  onSelectedChange={handleSelectedChange}
>
  <Combobox.Input
    class="flex justify-center items-center pl-2 w-full border border-gray-300 rounded-[2px] cursor-pointer min-h-8 h-fit focus-within:border-primary-500 focus-within:ring-2 focus-within:ring-primary-100 focus:outline-none"
    {placeholder}
    value={inputValue}
    aria-label={placeholder}
  />

  {#if filteredItems.length}
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
          <AvatarListItem
            name={getMetadata(item.value)?.name || item.label}
            email={item.value}
            photoUrl={getMetadata(item.value)?.photoUrl}
            leftSpacing={false}
          />
        </Combobox.Item>
      {/each}
    </Combobox.Content>
  {/if}
</Combobox.Root>
