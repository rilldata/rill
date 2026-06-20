<script lang="ts">
  import { cn } from "@rilldata/web-common/lib/shadcn";
  import { ChevronLeft, ChevronRight, Plus, X } from "lucide-svelte";
  import type { BaseCanvasComponent } from "./components/BaseCanvasComponent";
  import type { TabGroup } from "./stores/tab-group";

  export let group: TabGroup;
  export let maxWidth: number;
  // Called with the selected tab's stable name so the parent can sync URL state.
  export let onSelect: ((tabName: string) => void) | undefined = undefined;
  // Edit-mode affordances.
  export let editable: boolean = false;
  export let onAddTab: (() => void) | undefined = undefined;
  export let onRenameTab:
    | ((tabIndex: number, label: string) => void)
    | undefined = undefined;
  export let onDeleteTab: ((tabIndex: number) => void) | undefined = undefined;
  export let onMoveTab:
    | ((tabIndex: number, direction: -1 | 1) => void)
    | undefined = undefined;
  // Reorder a tab by dragging its title (from index -> to index).
  export let onReorderTab: ((from: number, to: number) => void) | undefined =
    undefined;
  // When a component is being dragged, tabs become drop targets for cross-tab moves.
  export let dragComponent: BaseCanvasComponent | null = null;
  export let onDropOnTab: ((tabIndex: number) => void) | undefined = undefined;

  $: tabs = group.tabs;
  $: activeTabIndex = group.activeTabIndex;
  $: dragging = !!dragComponent;

  let renamingIndex = -1;
  // Index of the tab currently being dragged to reorder, or -1 when not reordering.
  let reorderFrom = -1;

  function select(index: number) {
    if (index === $activeTabIndex) return;
    group.activeTabIndex.set(index);
    const tab = $tabs[index];
    if (tab && onSelect) onSelect(tab.name);
  }

  function commitRename(index: number, value: string) {
    const trimmed = value.trim();
    renamingIndex = -1;
    if (trimmed && trimmed !== $tabs[index]?.displayName) {
      onRenameTab?.(index, trimmed);
    }
  }

  function dropReorder(index: number) {
    if (reorderFrom !== -1 && reorderFrom !== index) {
      onReorderTab?.(reorderFrom, index);
    }
    reorderFrom = -1;
  }
</script>

<div class="tab-strip-wrapper" style:max-width="{maxWidth}px">
  <div
    role="tablist"
    class="inline-flex h-9 items-center justify-start gap-x-4 border-b border-gray-200 w-full"
  >
    {#each $tabs as tab, index (tab.name)}
      <div
        role="presentation"
        class="group/tab flex items-center"
        class:cursor-grab={editable && !dragging && renamingIndex !== index}
        class:opacity-50={reorderFrom === index}
        draggable={editable && !dragging && renamingIndex !== index}
        on:dragstart={(e) => {
          reorderFrom = index;
          // Firefox only initiates a drag when drag data is set.
          e.dataTransfer?.setData("text/plain", String(index));
        }}
        on:dragover={(e) => {
          if (reorderFrom !== -1) e.preventDefault();
        }}
        on:drop|preventDefault={() => dropReorder(index)}
        on:dragend={() => (reorderFrom = -1)}
      >
        {#if editable && renamingIndex === index}
          <!-- svelte-ignore a11y-autofocus -->
          <input
            class="text-sm font-medium border-b-2 border-primary-500 bg-transparent outline-none w-24"
            value={tab.displayName}
            autofocus
            on:blur={(e) => commitRename(index, e.currentTarget.value)}
            on:keydown={(e) => {
              if (e.key === "Enter") e.currentTarget.blur();
              else if (e.key === "Escape") {
                renamingIndex = -1;
              }
            }}
          />
        {:else}
          {#if editable}
            <button
              class="mb-2 opacity-0 group-hover/tab:opacity-100 text-fg-secondary hover:text-fg-primary transition-opacity disabled:opacity-0"
              title="Move tab left"
              disabled={index === 0}
              on:click={() => onMoveTab?.(index, -1)}
            >
              <ChevronLeft size="12px" />
            </button>
          {/if}
          <button
            role="tab"
            aria-selected={index === $activeTabIndex}
            class={cn(
              "inline-flex items-center justify-center whitespace-nowrap text-sm font-medium text-fg-secondary transition-all",
              "pb-2 rounded-none border-b-2 border-transparent -mb-px",
              "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
              index === $activeTabIndex && "border-primary-500 text-fg-primary",
              dragging &&
                "rounded-t bg-primary-50 ring-1 ring-primary-300 ring-inset",
            )}
            on:click={() => select(index)}
            on:dblclick={() => {
              if (editable) renamingIndex = index;
            }}
            on:mouseup={() => {
              if (dragging) onDropOnTab?.(index);
            }}
          >
            {tab.displayName}
          </button>
          {#if editable}
            <button
              class="opacity-0 group-hover/tab:opacity-100 text-fg-secondary hover:text-fg-primary transition-opacity disabled:opacity-0"
              title="Move tab right"
              disabled={index === $tabs.length - 1}
              on:click={() => onMoveTab?.(index, 1)}
            >
              <ChevronRight size="12px" />
            </button>
            <button
              class="ml-1 mb-2 opacity-0 group-hover/tab:opacity-100 text-fg-secondary hover:text-fg-primary transition-opacity"
              title="Delete tab"
              on:click={() => onDeleteTab?.(index)}
            >
              <X size="12px" />
            </button>
          {/if}
        {/if}
      </div>
    {/each}

    {#if editable}
      <button
        class="mb-2 text-fg-secondary hover:text-fg-primary"
        title="Add tab"
        on:click={() => onAddTab?.()}
      >
        <Plus size="14px" />
      </button>
    {/if}
  </div>
</div>

<style lang="postcss">
  .tab-strip-wrapper {
    @apply w-full mx-auto px-2;
  }
</style>
