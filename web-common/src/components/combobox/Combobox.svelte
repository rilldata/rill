<script lang="ts">
  import AvatarListItem from "../avatar/AvatarListItem.svelte";
  import { Combobox } from "bits-ui";
  import type { Selected } from "bits-ui";
  import { Check } from "lucide-svelte";
  import { onMount, afterUpdate } from "svelte";
  import LoadingSpinner from "../icons/LoadingSpinner.svelte";

  type Option = {
    value: string;
    label: string;
  };

  export let options: Option[] = [];
  export let searchValue = "";
  export let placeholder = "Search";
  export let disabled = false;
  export let required = false;
  export let error: string | undefined = undefined;
  export let selectedValues: string[] = [];
  export let getMetadata: (
    value: string,
  ) => { name: string; photoUrl?: string } | undefined = () => undefined;
  export let enableClientFiltering = true; // When false, pass options as-is (for server-side filtering)
  export let hasMore: boolean = false;
  export let isLoadingMore: boolean = false;
  export let loadMore: (() => void | Promise<void>) | undefined = undefined;
  export let onSelectedChange: (
    value: Selected<string>[] | undefined,
  ) => void = () => {};

  const LOAD_MORE_THRESHOLD_PX = 24;

  let initialSelectedItems: Selected<string>[] = [];
  let contentEl: HTMLElement | null = null;
  let isRequestInFlight = false;
  let lastSearchValue = "";

  onMount(() => {
    // Initialize the selected state for bits-ui combobox
    initialSelectedItems = selectedValues.map((value) => ({
      value,
      label: options.find((opt) => opt.value === value)?.label || value,
    }));
  });

  $: if (!Array.isArray(options)) {
    console.error("Combobox: options must be an array");
    options = [];
  }

  $: filteredItems =
    enableClientFiltering && searchValue
      ? options.filter((option) => {
          if (!option?.value || !option?.label) return false;
          return (
            option.value.toLowerCase().includes(searchValue.toLowerCase()) ||
            option.label.toLowerCase().includes(searchValue.toLowerCase())
          );
        })
      : options;

  // Update initialSelectedItems when selectedValues changes
  $: initialSelectedItems = selectedValues.map((value) => ({
    value,
    label: options.find((opt) => opt.value === value)?.label || value,
  }));

  function handleSelectedChange(selected: Selected<string>[] | undefined) {
    if (disabled) return;
    onSelectedChange(selected);
  }

  function getValidMetadata(value: string) {
    try {
      return getMetadata(value);
    } catch {
      return undefined;
    }
  }

  function maybeLoadMoreOnScroll() {
    if (!loadMore || !contentEl) return;
    const target = contentEl;
    const { scrollTop, clientHeight, scrollHeight } = target;
    const nearBottom =
      scrollTop + clientHeight >= scrollHeight - LOAD_MORE_THRESHOLD_PX;
    if (nearBottom && hasMore && !isLoadingMore && !isRequestInFlight) {
      isRequestInFlight = true;
      Promise.resolve(loadMore()).finally(() => {
        isRequestInFlight = false;
      });
    }
  }

  // Reset scroll when search changes
  $: if (searchValue !== lastSearchValue) {
    lastSearchValue = searchValue;
    if (contentEl) contentEl.scrollTop = 0;
  }

  // If the list doesn't overflow yet and there are more pages, pull more to fill the viewport
  afterUpdate(() => {
    if (!contentEl || !loadMore) return;
    if (!hasMore || isLoadingMore || isRequestInFlight) return;
    const { clientHeight, scrollHeight } = contentEl;
    if (scrollHeight <= clientHeight + LOAD_MORE_THRESHOLD_PX) {
      isRequestInFlight = true;
      Promise.resolve(loadMore()).finally(() => {
        isRequestInFlight = false;
      });
    }
  });
</script>

<Combobox.Root
  items={filteredItems}
  onSelectedChange={handleSelectedChange}
  multiple={true}
  bind:inputValue={searchValue}
  selected={initialSelectedItems}
  {disabled}
  {required}
>
  <Combobox.Input
    class="flex justify-center items-center pl-2 w-full border border-gray-300 rounded-[2px] cursor-pointer min-h-8 h-fit focus-within:border-primary-500 focus-within:ring-2 focus-within:ring-primary-100 focus:outline-none"
    {placeholder}
    aria-label={placeholder}
    aria-invalid={!!error}
    aria-describedby={error ? "combobox-error" : undefined}
    {disabled}
    {required}
  />

  {#if error}
    <div id="combobox-error" class="text-red-500 text-sm mt-1">
      {error}
    </div>
  {/if}

  <!-- NOTE: 52px * 4 for 208px to show scroller -->
  <Combobox.Content
    class="w-full rounded-sm border border-muted bg-surface p-[6px] shadow-md outline-none"
    sideOffset={8}
  >
    <div
      class="max-h-[208px] overflow-y-auto"
      bind:this={contentEl}
      on:scroll={maybeLoadMoreOnScroll}
    >
      {#if filteredItems.length === 0}
        <div class="px-4 py-2 text-xs text-gray-500">No results found</div>
      {:else}
        {#each filteredItems as item (item.value)}
          <Combobox.Item
            class="flex h-[52px] w-full select-none items-center rounded px-4 py-2 text-sm outline-none transition-all duration-75 data-[highlighted]:bg-slate-100"
            value={item.value}
            label={item.label}
            {disabled}
          >
            <AvatarListItem
              name={getValidMetadata(item.value)?.name || item.label}
              email={item.value}
              photoUrl={getValidMetadata(item.value)?.photoUrl}
              leftSpacing={false}
            />
            <div class="grow"></div>
            <Combobox.ItemIndicator>
              <Check size="16px" />
            </Combobox.ItemIndicator>
          </Combobox.Item>
        {/each}
        {#if hasMore && isLoadingMore}
          <div
            class="px-4 py-2 text-xs text-gray-500 flex items-center justify-center"
          >
            <LoadingSpinner size="16px" />
          </div>
        {/if}
      {/if}
    </div>
  </Combobox.Content>
</Combobox.Root>
