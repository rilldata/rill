<script lang="ts">
  import { writable } from "svelte/store";
  import { type InlineContext } from "@rilldata/web-common/features/chat/core/context/inline-context.ts";
  import PickerGroup from "@rilldata/web-common/features/chat/core/context/picker/PickerGroup.svelte";
  import { PickerOptionsHighlightManager } from "@rilldata/web-common/features/chat/core/context/picker/highlight-manager.ts";
  import {
    autoUpdate,
    computePosition,
    offset,
    flip,
    shift,
    inline,
  } from "@floating-ui/dom";
  import { ArrowUp, ArrowDown, ArrowLeft, ArrowRight } from "lucide-svelte";
  import * as Kbd from "@rilldata/web-common/components/kbd";
  import { ContextPickerUIState } from "@rilldata/web-common/features/chat/core/context/picker/ui-state.ts";
  import { getFilteredPickerOptions } from "@rilldata/web-common/features/chat/core/context/picker/filters.ts";

  export let selectedChatContext: InlineContext | null = null;
  export let searchText: string = "";
  export let refNode: HTMLElement;
  export let onSelect: (ctx: InlineContext) => void;
  export let focusEditor: () => void;

  const searchTextStore = writable("");
  $: searchTextStore.set(searchText.replace(/^@/, ""));

  const uiState = new ContextPickerUIState();
  const filteredOptions = getFilteredPickerOptions(uiState, searchTextStore);

  const highlightManager = new PickerOptionsHighlightManager(uiState);
  const highlightedContext = highlightManager.highlightedContext;
  $: highlightManager.filterOptionsUpdated(
    $filteredOptions,
    selectedChatContext,
  );

  function handleKeyDown(event: KeyboardEvent) {
    switch (event.key) {
      case "ArrowUp":
        highlightManager.highlightPreviousContext();
        break;
      case "ArrowDown":
        highlightManager.highlightNextContext();
        break;
      case "ArrowLeft":
        highlightManager.collapseToClosestParent();
        break;
      case "ArrowRight":
        highlightManager.openCurrentParentOption();
        break;
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
  {#each $filteredOptions as parentOption (parentOption.context.key)}
    <PickerGroup
      {parentOption}
      {selectedChatContext}
      {uiState}
      {highlightManager}
      {searchTextStore}
      {onSelect}
      {focusEditor}
    />
  {:else}
    <div class="contents-empty">No matches found</div>
  {/each}
  <div class="inline-chat-navigation">
    <Kbd.Group>
      <Kbd.Root><ArrowUp size="12px" /></Kbd.Root>
      <Kbd.Root><ArrowDown size="12px" /></Kbd.Root>
      <span>Navigate,</span>
      <Kbd.Root><ArrowLeft size="12px" /></Kbd.Root>
      <Kbd.Root><ArrowRight size="12px" /></Kbd.Root>
      <span>Open/Close,</span>
      <Kbd.Root><span>Enter</span></Kbd.Root>
      <span>Select</span>
    </Kbd.Group>
  </div>
</div>

<style lang="postcss">
  .inline-chat-context-dropdown {
    @apply flex flex-col absolute top-0 left-0 p-1.5 z-50 w-[400px] max-h-[500px] overflow-auto;
    @apply border rounded-md bg-popover text-popover-foreground shadow-md;
  }

  .inline-chat-navigation {
    @apply flex flex-row items-center pt-1.5 px-1.5;
    @apply text-xs text-popover-foreground/60;
  }

  .contents-empty {
    @apply px-2 py-1.5 w-full ui-copy-inactive;
  }
</style>
