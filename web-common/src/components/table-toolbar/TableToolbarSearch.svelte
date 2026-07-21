<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import SearchIcon from "@rilldata/web-common/components/icons/Search.svelte";
  import type { RuneStore } from "@rilldata/web-common/lib/store-utils/types.svelte.ts";

  let {
    searchTextStore,
  }: {
    searchTextStore?: RuneStore<string>;
  } = $props();

  function handleInput(event: Event & { currentTarget: HTMLInputElement }) {
    searchTextStore?.setter(event.currentTarget?.value);
  }
</script>

{#if searchTextStore}
  <div
    class="flex flex-row items-center gap-x-2 h-9 flex-1 min-w-0 border px-3 py-2.5"
  >
    <SearchIcon size="16" className="text-fg-muted shrink-0" />
    <input
      value={searchTextStore.value}
      type="text"
      class="outline-none bg-transparent text-sm text-fg-primary placeholder-fg-muted flex-1 min-w-0"
      placeholder={m.common_search_ellipsis()}
      oninput={handleInput}
    />
  </div>
{/if}
