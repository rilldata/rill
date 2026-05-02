<script lang="ts">
  import SearchIcon from "@rilldata/web-common/components/icons/Search.svelte";

  let {
    searchText = $bindable(""),
    disabled = false,
    placeholder = "Search",
  }: {
    searchText?: string;
    disabled?: boolean;
    placeholder?: string;
  } = $props();
</script>

<div
  class="flex flex-row items-center gap-x-2 h-9 flex-1 min-w-0 border px-3 py-2.5"
>
  <SearchIcon size="16" className="text-fg-muted shrink-0" />
  <input
    bind:value={searchText}
    type="text"
    class="outline-none bg-transparent text-sm text-fg-primary placeholder-fg-muted flex-1 min-w-0"
    {placeholder}
    {disabled}
  />
</div>
