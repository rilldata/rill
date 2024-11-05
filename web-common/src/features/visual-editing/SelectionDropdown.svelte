<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import Search from "@rilldata/web-common/components/search/Search.svelte";

  export let allItems: Set<string>;
  export let selectedItems: Set<string>;
  export let onSelect: (item: string) => void;
  export let onToggleSelectAll: () => void;

  let searchValue = "";
  let open = false;

  $: filteredItems = Array.from(allItems).filter((item) => {
    return (
      !selectedItems.has(item) &&
      item.toLowerCase().includes(searchValue.toLowerCase())
    );
  });
</script>

<DropdownMenu.Root bind:open typeahead={false} closeOnItemClick={false}>
  <DropdownMenu.Trigger asChild let:builder>
    <button
      use:builder.action
      {...builder}
      class:open
      class="flex px-3 gap-x-2 h-8 max-w-full items-center text-sm border-gray-300 border rounded-[2px] break-all overflow-hidden"
    >
      {selectedItems.size} of {allItems.size}
      <CaretDownIcon size="12px" className="!fill-gray-600 ml-auto flex-none" />
    </button>
  </DropdownMenu.Trigger>

  <DropdownMenu.Content sameWidth class="p-0">
    <div class="p-3 pb-1">
      <Search bind:value={searchValue} autofocus={false} />
    </div>
    <div class="max-h-64 overflow-y-auto">
      {#if !searchValue}
        {#each selectedItems as item (item)}
          <DropdownMenu.CheckboxItem
            checked
            class="mx-1 cursor-pointer"
            on:click={() => {
              onSelect(item);
            }}
          >
            {item}
          </DropdownMenu.CheckboxItem>
        {/each}
      {/if}

      {#if selectedItems.size > 0 && filteredItems.length > 0}
        <DropdownMenu.Separator />
      {/if}

      {#each filteredItems as item (item)}
        <DropdownMenu.Item
          class="pl-8 mx-1"
          on:click={() => {
            onSelect(item);
          }}
        >
          {item}
        </DropdownMenu.Item>
      {:else}
        {#if searchValue}
          <div class="ui-copy-disabled text-center p-2 w-full">no results</div>
        {/if}
      {/each}
    </div>

    <footer>
      <Button on:click={onToggleSelectAll} type="plain">
        {#if selectedItems.size === allItems.size}
          Deselect all
        {:else}
          Select all
        {/if}
      </Button>
    </footer>
  </DropdownMenu.Content>
</DropdownMenu.Root>

<style lang="postcss">
  .open {
    @apply ring-2 ring-primary-100 border-primary-600;
  }

  footer {
    @apply mt-1;
    height: 42px;
    @apply border-t border-slate-300;
    @apply bg-slate-100;
    @apply flex flex-row flex-none items-center justify-end;
    @apply gap-x-2 p-2 px-3.5;
  }
</style>
