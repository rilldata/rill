<script lang="ts">
  import { onMount } from "svelte";
  import { writable } from "svelte/store";
  import {
    getIdForContext,
    type InlineContext,
  } from "@rilldata/web-common/features/chat/core/context/inline-context.ts";
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
  import { getFilteredPickerItems } from "@rilldata/web-common/features/chat/core/context/picker/filters.ts";
  import { buildPickerTree } from "@rilldata/web-common/features/chat/core/context/picker/picker-tree.ts";
  import { KeyboardNavigationManager } from "@rilldata/web-common/features/chat/core/context/picker/keyboard-navigation.ts";
  import ExpandableOption from "@rilldata/web-common/features/chat/core/context/picker/ExpandableOption.svelte";
  import SimpleOption from "@rilldata/web-common/features/chat/core/context/picker/SimpleOption.svelte";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";

  export let selectedChatContext: InlineContext | null = null;
  export let searchText: string = "";
  export let refNode: HTMLElement;
  export let onSelect: (ctx: InlineContext) => void;
  export let focusEditor: () => void;

  $: selectedItemId = selectedChatContext
    ? getIdForContext(selectedChatContext)
    : null;

  const runtimeClient = useRuntimeClient();

  const searchTextStore = writable("");
  $: searchTextStore.set(searchText.replace(/^@/, ""));

  const uiState = new ContextPickerUIState();
  const expandedParentsStore = uiState.expandedParentsStore;
  const filteredOptions = getFilteredPickerItems(
    runtimeClient,
    uiState,
    searchTextStore,
  );
  $: pickerTree = buildPickerTree($filteredOptions);

  const keyboardNavigationManager = new KeyboardNavigationManager(uiState);
  $: keyboardNavigationManager.setPickerItems(
    $filteredOptions,
    $expandedParentsStore,
    selectedItemId,
  );

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

  onMount(() => {
    return keyboardNavigationManager.on("select", onSelect);
  });
</script>

<svelte:window on:keydown={(e) => keyboardNavigationManager.handleKeyDown(e)} />

<!-- bits-ui dropdown component captures focus, so chat text cannot be edited when it is open.
     Newer versions of bits-ui have "trapFocus=false" param but it needs svelte5 upgrade.
     TODO: move to dropdown component after upgrade. -->
<div class="inline-chat-context-dropdown" use:positionHandler={refNode}>
  <div class="dropdown-content">
    {#each pickerTree.rootNodes as rootNode (rootNode.item.id)}
      {@const showBoundary = pickerTree.boundaryIndices.has(rootNode.item.id)}
      {#if showBoundary}
        <div class="section-boundary"></div>
      {/if}

      {#if rootNode.item.hasChildren}
        <ExpandableOption
          node={rootNode}
          {selectedChatContext}
          {keyboardNavigationManager}
          {uiState}
          {searchTextStore}
          {onSelect}
          {focusEditor}
        />
      {:else}
        <SimpleOption
          item={rootNode.item}
          {selectedChatContext}
          {keyboardNavigationManager}
          {onSelect}
        />
      {/if}
    {:else}
      <div class="contents-empty">No matches found</div>
    {/each}
  </div>
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
    @apply flex flex-col absolute top-0 left-0 p-1.5 z-50;
    @apply border rounded-md bg-popover text-popover-foreground shadow-md;
  }

  .dropdown-content {
    @apply flex flex-col w-[400px] max-h-[500px] overflow-auto;
  }

  .inline-chat-navigation {
    @apply flex flex-row items-center pt-1.5 px-1.5;
    @apply text-xs text-popover-foreground/60;
  }

  .contents-empty {
    @apply px-2 py-1.5 w-full text-fg-disabled;
  }

  .section-boundary {
    @apply border-b;
  }
</style>
