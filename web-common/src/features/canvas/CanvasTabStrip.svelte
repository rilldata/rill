<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import { cn } from "@rilldata/web-common/lib/shadcn";
  import { ArrowLeft, ArrowRight, Pencil, Plus, Trash2 } from "lucide-svelte";
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

  let renamingIndex = -1;

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
</script>

<div class="tab-strip-wrapper" style:max-width="{maxWidth}px">
  <div
    role="tablist"
    class="inline-flex h-9 items-stretch justify-start gap-x-4 border-b border-gray-200 w-full"
  >
    {#each $tabs as tab, index (tab.name)}
      <div role="presentation" class="group/tab flex items-stretch -mb-px">
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
              "px-1 pb-2 rounded-none border-b-2 border-transparent",
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
                  disabled={index === $tabs.length - 1}
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

<style lang="postcss">
  .tab-strip-wrapper {
    @apply w-full mx-auto px-2;
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
