<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import { cn } from "@rilldata/web-common/lib/shadcn";
  import {
    ArrowLeft,
    ArrowRight,
    ChevronLeft,
    ChevronRight,
    Pencil,
    Plus,
    Trash2,
  } from "lucide-svelte";
  import { onMount, tick } from "svelte";
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
  // When a component is being dragged, tabs become drop targets for cross-tab moves.
  export let dragComponent: BaseCanvasComponent | null = null;
  export let onDropOnTab: ((tabIndex: number) => void) | undefined = undefined;

  $: tabs = group.tabs;
  $: activeTabIndex = group.activeTabIndex;
  $: dragging = !!dragComponent;
  $: tabsLength = $tabs.length;

  let renamingIndex = -1;
  let tabStripWrapper: HTMLDivElement | undefined;
  let tabList: HTMLDivElement | undefined;
  let canScrollToStart = false;
  let canScrollToEnd = false;
  let dragOverTabIndex = -1;

  $: if (!dragging) dragOverTabIndex = -1;

  function updateOverflow() {
    if (!tabStripWrapper) return;

    const maxScrollLeft =
      tabStripWrapper.scrollWidth - tabStripWrapper.clientWidth;
    canScrollToStart = tabStripWrapper.scrollLeft > 1;
    canScrollToEnd =
      maxScrollLeft > 1 && tabStripWrapper.scrollLeft < maxScrollLeft - 1;
  }

  async function updateOverflowAfterRender() {
    await tick();
    updateOverflow();
  }

  function scrollToEnd() {
    tabStripWrapper?.scrollTo({
      left: tabStripWrapper.scrollWidth,
      behavior: "smooth",
    });
  }

  function scrollToStart() {
    tabStripWrapper?.scrollTo({
      left: 0,
      behavior: "smooth",
    });
  }

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

  $: overflowKey = `${maxWidth}:${editable}:${renamingIndex}:${$tabs
    .map((tab) => tab.displayName)
    .join("\0")}`;
  $: if (overflowKey) {
    void updateOverflowAfterRender();
  }

  onMount(() => {
    void updateOverflowAfterRender();

    if (typeof ResizeObserver === "undefined") return;

    const resizeObserver = new ResizeObserver(updateOverflow);
    if (tabStripWrapper) resizeObserver.observe(tabStripWrapper);
    if (tabList) resizeObserver.observe(tabList);

    return () => resizeObserver.disconnect();
  });
</script>

<div class="tab-strip-shell" style:max-width="{maxWidth}px">
  {#if canScrollToStart}
    <button
      class="scroll-button scroll-to-start"
      aria-label="Show previous tabs"
      title="Show previous tabs"
      on:click={scrollToStart}
    >
      <ChevronLeft size="16px" />
    </button>
  {/if}

  <div
    bind:this={tabStripWrapper}
    on:scroll={updateOverflow}
    class="tab-strip-wrapper"
  >
    <div
      bind:this={tabList}
      role="tablist"
      class="inline-flex h-9 min-w-full w-max items-stretch justify-start gap-x-4 border-b border-gray-200"
    >
      {#each $tabs as tab, index (tab.name)}
        <div
          role="presentation"
          class="group/tab flex flex-none items-stretch -mb-px"
        >
          {#if editable && renamingIndex === index}
            <!-- svelte-ignore a11y-autofocus -->
            <input
              class="text-sm font-medium border-b-2 border-primary-500 bg-transparent outline-none pb-2 px-1"
              size={Math.max(tab.displayName.length, 8)}
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
            <button
              role="tab"
              aria-selected={index === $activeTabIndex}
              class={cn(
                "inline-flex items-center justify-center whitespace-nowrap text-sm font-medium text-fg-secondary transition-all",
                "px-1 pb-2 rounded-t-sm border-b-2 border-transparent",
                "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
                dragging && "cursor-copy border-primary-200/60 text-fg-primary",
                index === $activeTabIndex &&
                  "border-primary-500 text-fg-primary",
                dragging &&
                  dragOverTabIndex === index &&
                  "bg-primary-50 border-primary-500 shadow-[inset_0_0_0_1px_var(--color-primary-300)]",
              )}
              on:click={() => select(index)}
              on:dblclick={() => {
                if (editable) renamingIndex = index;
              }}
              on:mouseenter={() => {
                if (dragging) dragOverTabIndex = index;
              }}
              on:mousemove={() => {
                if (dragging && dragOverTabIndex !== index) {
                  dragOverTabIndex = index;
                }
              }}
              on:mouseleave={() => {
                if (dragOverTabIndex === index) dragOverTabIndex = -1;
              }}
              on:mouseup={() => {
                if (dragging) onDropOnTab?.(index);
                dragOverTabIndex = -1;
              }}
            >
              {tab.displayName}
            </button>
            {#if editable && !dragging}
              <DropdownMenu.Root>
                <DropdownMenu.Trigger>
                  {#snippet child({ props })}
                    <button
                      {...props}
                      class="ctrl opacity-0 group-hover/tab:opacity-100 data-[state=open]:opacity-100"
                      title="Tab options"
                      aria-label="Tab options for {tab.displayName}"
                    >
                      <ThreeDot size="14px" />
                    </button>
                  {/snippet}
                </DropdownMenu.Trigger>
                <DropdownMenu.Content align="start" class="min-w-[160px]">
                  <DropdownMenu.Item
                    class="flex flex-row gap-x-2 text-fg-primary"
                    onclick={() => (renamingIndex = index)}
                  >
                    <Pencil size="14px" />
                    Rename
                  </DropdownMenu.Item>
                  <DropdownMenu.Item
                    class="flex flex-row gap-x-2 text-fg-primary"
                    disabled={index === 0}
                    onclick={() => onMoveTab?.(index, -1)}
                  >
                    <ArrowLeft size="14px" />
                    Move left
                  </DropdownMenu.Item>
                  <DropdownMenu.Item
                    class="flex flex-row gap-x-2 text-fg-primary"
                    disabled={index === tabsLength - 1}
                    onclick={() => onMoveTab?.(index, 1)}
                  >
                    <ArrowRight size="14px" />
                    Move right
                  </DropdownMenu.Item>
                  <DropdownMenu.Separator />
                  <DropdownMenu.Item
                    class="flex flex-row gap-x-2 text-red-600"
                    onclick={() => onDeleteTab?.(index)}
                  >
                    <Trash2 size="14px" />
                    Delete
                  </DropdownMenu.Item>
                </DropdownMenu.Content>
              </DropdownMenu.Root>
            {/if}
          {/if}
        </div>
      {/each}

      {#if editable}
        <button class="add-tab" title="Add tab" on:click={() => onAddTab?.()}>
          <Plus size="14px" />
        </button>
      {/if}
    </div>
  </div>

  {#if canScrollToEnd}
    <button
      class="scroll-button scroll-to-end"
      aria-label="Show more tabs"
      title="Show more tabs"
      on:click={scrollToEnd}
    >
      <ChevronRight size="16px" />
    </button>
  {/if}
</div>

<style lang="postcss">
  .tab-strip-shell {
    @apply relative z-[100] flex w-full max-w-full mx-auto items-stretch px-2;
  }

  .tab-strip-wrapper {
    @apply min-w-0 flex-1 overflow-x-auto;
    scrollbar-width: none;
    -ms-overflow-style: none;
  }

  .tab-strip-wrapper::-webkit-scrollbar {
    @apply hidden;
  }

  .scroll-button {
    @apply flex w-7 flex-none items-center justify-center pb-2;
    @apply border-b border-gray-200 text-fg-secondary;
    @apply transition-colors hover:text-fg-primary;
  }

  /* Per-tab options trigger (the ⋯ menu). A full-height flex child so its
     icon optically centers against the tab label, sharing the strip's
     baseline via the same pb-2 the label uses. It fades in on tab hover
     via the group-hover/tab variant in markup. */
  .ctrl {
    @apply flex items-center pb-2 px-0.5 border-b-2 border-transparent;
    @apply text-fg-secondary transition-opacity hover:text-fg-primary;
  }

  /* Add-tab button sits after the last tab, separated from the cluster. */
  .add-tab {
    @apply flex items-center pb-2 ml-2 border-b-2 border-transparent;
    @apply text-fg-secondary transition-colors hover:text-fg-primary;
  }
</style>
