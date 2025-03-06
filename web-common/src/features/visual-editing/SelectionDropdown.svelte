<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import Search from "@rilldata/web-common/components/search/Search.svelte";

  export let allItems: Set<string>;
  export let selectedItems: Set<string>;
  export let type: string | undefined = undefined;
  export let searchValue = "";
  export let searchableItems: string[] | undefined = undefined;
  export let excludable: boolean = false;
  export let excludeMode: boolean = false;
  export let small = false;
  export let id: string = "";
  export let onSelect: (item: string) => void;
  export let setItems: (items: string[], exclude?: boolean) => void;

  let open = false;
  let selectedProxy = new Set(selectedItems);
  let allProxy = new Set(allItems);

  $: filteredItems = (
    searchableItems && searchValue ? searchableItems : Array.from(allProxy)
  ).filter((item) => {
    return (
      !selectedProxy.has(item) &&
      item.toLowerCase().includes(searchValue.toLowerCase())
    );
  });
</script>

<DropdownMenu.Root
  bind:open
  typeahead={false}
  closeOnItemClick={false}
  onOpenChange={() => {
    if (!open) {
      selectedProxy = new Set(selectedItems);
      allProxy = new Set(allItems);
    }
  }}
>
  <DropdownMenu.Trigger asChild let:builder {id}>
    <button
      use:builder.action
      {...builder}
      class:open
      class:small
      class="dropdown-trigger"
    >
      {#if type}
        {selectedItems.size} {type}
      {:else}
        {selectedItems.size} of {allItems.size}
      {/if}

      <CaretDownIcon size="12px" className="!fill-gray-600 ml-auto flex-none" />
    </button>
  </DropdownMenu.Trigger>

  <DropdownMenu.Content sameWidth class="p-0">
    <div class="p-3 pb-1">
      <Search bind:value={searchValue} autofocus={false} />
    </div>
    <div class="max-h-64 overflow-y-auto">
      {#if !searchValue}
        {#each selectedProxy as item (item)}
          <DropdownMenu.CheckboxItem
            showXForSelected={excludeMode}
            checked={selectedItems.has(item)}
            class="mx-1 cursor-pointer"
            on:click={() => {
              onSelect(item);
            }}
          >
            <slot {item}>
              {item}
            </slot>
          </DropdownMenu.CheckboxItem>
        {/each}
      {/if}

      {#if selectedProxy.size > 0 && filteredItems.length > 0}
        <DropdownMenu.Separator />
      {/if}

      {#each filteredItems as item (item)}
        <DropdownMenu.CheckboxItem
          showXForSelected={excludeMode}
          checked={selectedItems.has(item)}
          class="pl-8 mx-1"
          on:click={() => {
            onSelect(item);
          }}
        >
          <slot {item}>
            {item}
          </slot>
        </DropdownMenu.CheckboxItem>
      {:else}
        {#if searchValue}
          <div class="ui-copy-disabled text-center p-2 w-full">no results</div>
        {/if}
      {/each}
    </div>

    <footer>
      {#if excludable}
        <Button
          on:click={() => {
            setItems(Array.from(selectedItems), !excludeMode);
          }}
          type="secondary"
        >
          {#if excludeMode}
            Include
          {:else}
            Exclude
          {/if}
        </Button>
      {/if}

      <Button
        on:click={() => {
          if (selectedItems.size === allItems.size) {
            setItems([]);
          } else {
            setItems(Array.from(allItems), excludeMode);
          }
        }}
        type="plain"
      >
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
  .dropdown-trigger {
    @apply flex items-center;
    @apply px-3 gap-x-2 h-8 max-w-full;
    @apply text-sm;
    @apply border-gray-300 border rounded-[2px];
    @apply break-all overflow-hidden;
  }

  .dropdown-trigger.small {
    @apply h-6 text-xs;
  }
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
