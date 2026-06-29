<script lang="ts">
  import * as m from "@rilldata/web-common/paraglide/messages.js";
  import SearchIcon from "@rilldata/web-common/components/icons/Search.svelte";
  import { X } from "lucide-svelte";
  import { tick } from "svelte";

  let {
    searchText = $bindable(""),
    disabled = false,
  }: {
    searchText?: string;
    disabled?: boolean;
  } = $props();

  let manualExpanded = $state(false);
  let expanded = $derived(manualExpanded || searchText.length > 0);
  let inputRef: HTMLInputElement | undefined = $state();

  async function open() {
    if (disabled) return;
    manualExpanded = true;
    await tick();
    inputRef?.focus();
  }

  function close() {
    searchText = "";
    manualExpanded = false;
  }

  function handleKeyDown(e: KeyboardEvent) {
    if (e.key === "Escape") {
      close();
    }
  }
</script>

{#if expanded}
  <div
    class="flex flex-row items-center gap-x-1.5 h-9 border rounded-sm bg-input px-2 min-w-[200px]"
  >
    <SearchIcon size="16" className="text-fg-secondary shrink-0" />
    <input
      bind:this={inputRef}
      bind:value={searchText}
      type="text"
      class="outline-none bg-transparent text-sm text-fg-primary placeholder-fg-secondary flex-1 min-w-0"
      placeholder={m.common_search()}
      onkeydown={handleKeyDown}
    />
    <button
      type="button"
      class="text-fg-secondary hover:text-fg-primary shrink-0"
      onclick={close}
      aria-label={m.common_close_search()}
    >
      <X size={14} />
    </button>
  </div>
{:else}
  <button
    type="button"
    class="flex items-center justify-center h-9 w-4 text-fg-primary hover:text-fg-secondary cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
    onclick={open}
    aria-label={m.common_search()}
    {disabled}
  >
    <SearchIcon size="16" className="text-fg-secondary" />
  </button>
{/if}
