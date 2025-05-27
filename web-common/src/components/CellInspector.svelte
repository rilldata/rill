<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import { createEventDispatcher } from "svelte";
  import { cellInspectorStore } from "../features/dashboards/stores/cellInspectorStore";
  import { formatInteger } from "../lib/formatters";

  export let value: string = "";
  export let isOpen: boolean = false;

  let hovered = false;
  let hoveredValue: string | null = null;
  let container: HTMLElement;
  let content: HTMLElement;

  // Subscribe to the cellInspectorStore to keep the component in sync
  const unsubscribe = cellInspectorStore.subscribe((state) => {
    isOpen = state.isOpen;
    if (state.value) {
      value = state.value;
    }
  });

  const dispatch = createEventDispatcher();

  // Handle keyboard events for toggling the cell inspector
  function handleKeyDown(event: KeyboardEvent) {
    // Only handle Space key when not in an input, textarea, or other form element
    const target = event.target as HTMLElement;
    const tagName = target.tagName.toLowerCase();
    const isFormElement =
      tagName === "input" || tagName === "textarea" || tagName === "select";

    if (event.code === "Space" && !event.repeat && !isFormElement) {
      event.preventDefault();
      event.stopPropagation();

      // Toggle the cell inspector visibility
      cellInspectorStore.toggle(value);
    } else if (event.key === "Escape" && isOpen) {
      event.preventDefault();
      event.stopPropagation();
      cellInspectorStore.close();
    }
  }

  function handleClickOutside(event: MouseEvent) {
    if (isOpen && container && !container.contains(event.target as Node)) {
      dispatch("close");
    }
  }

  onMount(() => {
    // Handle click outside events
    document.addEventListener("click", handleClickOutside, true);
    // Add keyboard event listener for spacebar toggle
    window.addEventListener("keydown", handleKeyDown, true);

    return () => {
      document.removeEventListener("click", handleClickOutside, true);
      window.removeEventListener("keydown", handleKeyDown, true);
    };
  });

  onDestroy(() => {
    // Clean up the subscription
    unsubscribe();
  });

  // Format number values for display using the tooltip formatter
  function formatValue(value: string): string {
    // Check if the value is a number
    const num = Number(value);
    if (!isNaN(num)) {
      return formatInteger(num);
    }
    return value;
  }

  // Only update the value on hover, but don't open the inspector
  $: if (hovered && hoveredValue && isOpen) {
    cellInspectorStore.open(hoveredValue);
  }
</script>

{#if isOpen}
  <div
    bind:this={container}
    class="fixed top-4 left-[calc(50%+120px)] z-50 transition-opacity shadow-lg rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800"
    class:invisible={!isOpen && !hovered}
    class:opacity-0={!isOpen && !hovered}
    class:opacity-100={isOpen || hovered}
    role="dialog"
    aria-labelledby="cell-inspector-title"
    aria-describedby="cell-inspector-content"
    aria-modal="true"
  >
    <div
      class="w-full max-w-2xl max-h-[80vh] overflow-hidden flex flex-col rounded-lg"
      role="document"
      bind:this={content}
    >
      <div
        class="flex items-center justify-between p-2 border-b border-gray-200 dark:border-gray-700"
        id="cell-inspector-content"
      >
        <div class="flex items-center" id="cell-inspector-title">
          <pre
            class="whitespace-pre-wrap break-words font-mono text-sm text-gray-800 dark:text-gray-200 flex-1">{formatValue(
              value,
            )}</pre>
        </div>
      </div>
      <div
        class="p-3 bg-gray-50 dark:bg-gray-700/50 text-xs text-gray-500 dark:text-gray-400 dark:border-gray-700 flex justify-center items-center"
      >
        <div class="flex space-x-4">
          <span
            >Press <kbd
              class="px-2 py-1 bg-gray-200 dark:bg-gray-600 text-xs font-mono"
              >Space</kbd
            > to close</span
          >
        </div>
      </div>
    </div>
  </div>
{/if}
