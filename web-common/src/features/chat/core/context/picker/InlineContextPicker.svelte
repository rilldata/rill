<script lang="ts">
  import { writable } from "svelte/store";
  import { type InlineContext } from "@rilldata/web-common/features/chat/core/context/inline-context.ts";
  import PickerGroup from "@rilldata/web-common/features/chat/core/context/picker/PickerGroup.svelte";
  import { PickerOptionsHighlightManager } from "@rilldata/web-common/features/chat/core/context/picker/highlight-manager.ts";
  import { getFilteredPickerOptions } from "@rilldata/web-common/features/chat/core/context/picker/data.ts";
  import {
    autoUpdate,
    computePosition,
    offset,
    flip,
    shift,
    inline,
  } from "@floating-ui/dom";

  export let selectedChatContext: InlineContext | null = null;
  export let searchText: string = "";
  export let refNode: HTMLElement;
  export let onSelect: (ctx: InlineContext) => void;
  export let focusEditor: () => void;

  const searchTextStore = writable("");
  $: searchTextStore.set(searchText.replace(/^@/, ""));

  const filteredOptions = getFilteredPickerOptions(searchTextStore);

  const highlightManager = new PickerOptionsHighlightManager();
  const highlightedContext = highlightManager.highlightedContext;
  $: highlightManager.filterOptionsUpdated(
    $filteredOptions,
    selectedChatContext,
  );

  function handleKeyDown(event: KeyboardEvent) {
    switch (event.key) {
      case "ArrowUp": {
        highlightManager.highlightPreviousContext();
        break;
      }
      case "ArrowDown": {
        highlightManager.highlightNextContext();
        break;
      }
      case "Enter":
        if ($highlightedContext) onSelect($highlightedContext);
        event.preventDefault();
        event.stopPropagation();
        event.stopImmediatePropagation();
        break;
    }
  }

  function positionHandler(node: Node, ref: HTMLElement) {
    if (!(node instanceof HTMLElement)) return;

    let refNode = ref;
    let cleanup: (() => void) | null = null;

    // Temporary minimal implementation of https://github.com/romkor/svelte-portal/blob/master/src/Portal.svelte
    // We wont need this once we upgrade to svelte5 and switch to using dropdown component.
    const update = (newRef: HTMLElement) => {
      cleanup?.();
      document.body.appendChild(node);

      refNode = newRef;
      cleanup = autoUpdate(refNode, node, compute);
      compute();
    };

    const compute = () => {
      void computePosition(refNode, node, {
        placement: "top-start",
        middleware: [offset(10), flip(), shift(), inline()],
      }).then(({ x, y }) => {
        Object.assign(node.style, {
          left: `${x}px`,
          top: `${y}px`,
        });
      });
    };

    update(ref);
    return {
      update,
      destroy() {
        cleanup?.();
        if (node.parentNode) {
          node.parentNode.removeChild(node);
        }
      },
    };
  }
</script>

<svelte:window on:keydown={handleKeyDown} />

<!-- bits-ui dropdown component captures focus, so chat text cannot be edited when it is open.
     Newer versions of bits-ui have "trapFocus=false" param but it needs svelte5 upgrade.
     TODO: move to dropdown component after upgrade. -->
<div class="inline-chat-context-dropdown" use:positionHandler={refNode}>
  {#each $filteredOptions as section (section.type)}
    <div class="border-b last:border-b-0">
      {#each section.options as parentOption (parentOption.context.value)}
        <PickerGroup
          {parentOption}
          {selectedChatContext}
          {highlightManager}
          {searchTextStore}
          {onSelect}
          {focusEditor}
        />
      {/each}
    </div>
  {:else}
    <div class="contents-empty">No matches found</div>
  {/each}
</div>

<style lang="postcss">
  .inline-chat-context-dropdown {
    @apply flex flex-col absolute top-0 left-0 p-1.5 z-50 w-[400px] max-h-[500px] overflow-auto;
    @apply border rounded-md bg-popover text-popover-foreground shadow-md;
  }

  .contents-empty {
    @apply px-2 py-1.5 w-full ui-copy-inactive;
  }
</style>
