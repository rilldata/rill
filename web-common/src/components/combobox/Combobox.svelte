<script lang="ts">
  import { Combobox, Selected } from "bits-ui";
  import { Check } from "lucide-svelte";

  export let options: { value: string; label: string }[] = [];
  export let inputValue = "";
  export let name = "";
  export let placeholder = "Search";
  export let label = "";
  export let id = "";
  export let onSelectedChange: (value: Selected<string> | undefined) => void;

  // FIXME: fuzzy
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

  <Combobox.Root items={filteredItems} bind:inputValue {onSelectedChange}>
    <Combobox.Input
      class="flex justify-center items-center pl-2 w-full border border-gray-300 rounded-[2px] cursor-pointer min-h-8 h-fit &:focus-within:border-primary-500 &:focus-within:ring-2 &:focus-within:ring-primary-100"
      {placeholder}
      aria-label={placeholder}
    />

    <Combobox.Content
      class="w-full rounded-sm border border-muted bg-background px-1 py-1 shadow-popover outline-none"
      sideOffset={8}
    >
      <!-- TODO: brb to polish -->
      {#each filteredItems as item (item.value)}
        <Combobox.Item
          class="flex h-10 w-full select-none items-center rounded p-4 text-sm outline-none transition-all duration-75 data-[highlighted]:bg-muted"
          value={item.value}
          label={item.label}
        >
          {item.label}
          <Combobox.ItemIndicator class="ml-auto" asChild={false}>
            <Check />
          </Combobox.ItemIndicator>
        </Combobox.Item>
      {:else}
        <span class="block px-5 py-2 text-sm text-muted-foreground">
          No results found
        </span>
      {/each}
    </Combobox.Content>
    <Combobox.HiddenInput {name} />
  </Combobox.Root>
</div>
