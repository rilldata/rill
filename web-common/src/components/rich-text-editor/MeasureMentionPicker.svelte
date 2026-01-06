<script lang="ts">
  import { writable } from "svelte/store";
  import {
    autoUpdate,
    computePosition,
    offset,
    flip,
    shift,
    inline,
  } from "@floating-ui/dom";

  export let availableMeasures: string[] = [];
  export let searchText: string = "";
  export let refNode: HTMLElement;
  export let onSelect: (measure: string) => void;
  export let focusEditor: () => void;

  const searchTextStore = writable("");
  $: searchTextStore.set(searchText.replace(/^@/, ""));

  $: filteredMeasures = availableMeasures
    .filter((measure) =>
      measure.toLowerCase().includes($searchTextStore.toLowerCase()),
    )
    .slice(0, 20);

  let highlightedIndex = 0;

  $: if (filteredMeasures.length > 0) {
    // Reset highlighted index when filtered results change
    highlightedIndex = Math.min(highlightedIndex, filteredMeasures.length - 1);
  }

  function handleKeyDown(event: KeyboardEvent) {
    switch (event.key) {
      case "ArrowUp":
        event.preventDefault();
        highlightedIndex = Math.max(0, highlightedIndex - 1);
        break;
      case "ArrowDown":
        event.preventDefault();
        highlightedIndex = Math.min(
          filteredMeasures.length - 1,
          highlightedIndex + 1,
        );
        break;
      case "Enter":
        if (filteredMeasures[highlightedIndex]) {
          onSelect(filteredMeasures[highlightedIndex]);
        }
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

    const update = (newRef: HTMLElement) => {
      cleanup?.();
      document.body.appendChild(node);

      refNode = newRef;
      cleanup = autoUpdate(refNode, node, compute);
      compute();
    };

    const compute = () => {
      void computePosition(refNode, node, {
        placement: "bottom-start",
        middleware: [offset(5), flip(), shift(), inline()],
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

<div class="measure-mention-picker" use:positionHandler={refNode}>
  {#if filteredMeasures.length > 0}
    {#each filteredMeasures as measure, index (measure)}
      <div
        class="measure-mention-item"
        class:highlighted={index === highlightedIndex}
        role="button"
        tabindex="0"
        on:click={() => onSelect(measure)}
        on:mouseenter={() => (highlightedIndex = index)}
      >
        <span class="measure-mention-char">@</span>
        <span class="measure-mention-name">{measure}</span>
      </div>
    {/each}
  {:else}
    <div class="measure-mention-empty">No measures found</div>
  {/if}
</div>

<style lang="postcss">
  .measure-mention-picker {
    @apply flex flex-col absolute top-0 left-0 p-1 z-50 w-[250px] max-h-[300px] overflow-auto;
    @apply border border-gray-300 rounded-md bg-white shadow-lg;
  }

  .measure-mention-item {
    @apply flex items-center gap-x-2 px-2 py-1.5 cursor-pointer rounded;
    @apply hover:bg-gray-100;
  }

  .measure-mention-item.highlighted {
    @apply bg-gray-100;
  }

  .measure-mention-char {
    @apply text-gray-500 font-medium;
  }

  .measure-mention-name {
    @apply text-sm text-gray-900;
  }

  .measure-mention-empty {
    @apply px-2 py-1.5 text-sm text-gray-500;
  }
</style>

