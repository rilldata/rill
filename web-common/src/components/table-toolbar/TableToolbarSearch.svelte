<script lang="ts">
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

  let expanded = $state(false);
  let inputRef: HTMLInputElement | undefined = $state();

  async function open() {
    if (disabled) return;
    expanded = true;
    await tick();
    inputRef?.focus();
  }

  function close() {
    searchText = "";
    expanded = false;
  }

  function handleKeyDown(e: KeyboardEvent) {
    if (e.key === "Escape") {
      close();
    }
  }

  function handleInput(e: Event) {
    searchText = (e.target as HTMLInputElement).value;
  }
</script>

{#if expanded}
  <div
    class="flex flex-row items-center gap-x-1.5 h-9 border rounded-sm bg-input px-2 min-w-[200px]"
  >
    <SearchIcon size="16" className="text-fg-secondary shrink-0" />
    <input
      bind:this={inputRef}
      value={searchText}
      oninput={handleInput}
      type="text"
      class="outline-none bg-transparent text-sm text-fg-primary placeholder-fg-secondary flex-1 min-w-0"
      placeholder="Search..."
      onkeydown={handleKeyDown}
    />
    <button
      class="text-fg-secondary hover:text-fg-primary shrink-0"
      onclick={close}
      aria-label="Close search"
    >
      <X size={14} />
    </button>
  </div>
{:else}
  <button
    class="flex items-center justify-center h-9 w-9 border rounded-sm bg-input hover:bg-surface-hover text-fg-secondary {disabled
      ? 'opacity-50 cursor-not-allowed pointer-events-none'
      : ''}"
    onclick={open}
    aria-label="Search"
  >
    <SearchIcon size="16" className="text-fg-secondary" />
  </button>
{/if}
