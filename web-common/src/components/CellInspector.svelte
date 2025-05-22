<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import { createEventDispatcher } from "svelte";
  import { cellInspectorStore } from "../features/dashboards/stores/cellInspectorStore";

  export let value: string = "";
  export let isOpen: boolean = false;

  const dispatch = createEventDispatcher();
  let hovered = false;
  let hoveredValue: string | null = null;
  let container: HTMLElement;
  let content: HTMLElement;

  function handleKeyDown(event: KeyboardEvent) {
    if (event.code === "Space" && !event.repeat) {
      event.preventDefault();
      event.stopPropagation();
      dispatch("toggle");
    } else if (event.key === "Escape") {
      event.preventDefault();
      event.stopPropagation();
      if (isOpen) {
        dispatch("close");
      }
    }
  }

  function handleClickOutside(event: MouseEvent) {
    if (isOpen && container && !container.contains(event.target as Node)) {
      dispatch("close");
    }
  }

  onMount(() => {
    window.addEventListener("keydown", handleKeyDown, true);
    document.addEventListener("click", handleClickOutside, true);

    // Handle hover events
    const handleHover = (e: MouseEvent) => {
      const target = e.target as HTMLElement;
      const cellValue = target.getAttribute("data-cell-value");

      if (cellValue) {
        hoveredValue = cellValue;
        hovered = true;
      } else {
        // Check if we're still hovering over a cell by looking at parent elements
        let current = target.parentElement;
        let found = false;

        while (current && !found) {
          if (current.hasAttribute("data-cell-value")) {
            hoveredValue = current.getAttribute("data-cell-value");
            found = true;
          }
          current = current.parentElement;
        }

        if (!found) {
          hovered = false;
        }
      }
    };

    document.addEventListener("mousemove", handleHover);

    return () => {
      window.removeEventListener("keydown", handleKeyDown, true);
      document.removeEventListener("click", handleClickOutside, true);
      document.removeEventListener("mousemove", handleHover);
    };
  });

  function stopPropagation(e: Event) {
    e.stopPropagation();
  }

  function copyToClipboard() {
    if (value) {
      navigator.clipboard
        .writeText(value)
        .then(() => {
          // Could add a toast notification here if desired
        })
        .catch((err) => {
          console.error("Failed to copy text: ", err);
        });
    }
  }

  // Only update the value on hover, but don't open the inspector
  $: if (hovered && hoveredValue && isOpen) {
    cellInspectorStore.open(hoveredValue, { x: 0, y: 0 });
  }
</script>

{#if isOpen}
  <div
    bind:this={container}
    class="fixed bottom-4 right-4 z-50 transition-opacity shadow-lg rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800"
    class:invisible={!isOpen && !hovered}
    class:opacity-0={!isOpen && !hovered}
    class:opacity-100={isOpen || hovered}
    role="dialog"
    aria-labelledby="cell-inspector-title"
    aria-describedby="cell-inspector-content"
    aria-modal="true"
  >
    <div
      class="w-full max-w-2xl max-h-[80vh] overflow-hidden flex flex-col"
      role="document"
      bind:this={content}
    >
      <div
        class="flex items-center justify-between p-0 h-8 border-b border-gray-200 dark:border-gray-700"
      >
        <div class="flex items-center px-2" id="cell-inspector-title">
          <svg
            xmlns="http://www.w3.org/2000/svg"
            class="h-5 w-5 text-slate-600 dark:text-gray-300 mr-2"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
            />
          </svg>
          <span class="text-sm font-medium text-slate-600 dark:text-gray-300"
            >Cell Value</span
          >
        </div>
        <button
          class="text-gray-400 hover:text-gray-500 dark:hover:text-gray-300 p-1 -mr-1"
          on:click|preventDefault={() => dispatch("close")}
          on:keydown={(e) => {
            if (e.key === "Enter" || e.key === " ") {
              e.preventDefault();
              dispatch("close");
            }
          }}
          aria-label="Close inspector"
          aria-controls="cell-inspector-content"
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            class="h-5 w-5"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M6 18L18 6M6 6l12 12"
            />
          </svg>
        </button>
      </div>
      <div id="cell-inspector-content" class="flex-1 overflow-auto p-2">
        <div class="flex justify-end">
          <pre
            class="whitespace-pre-wrap break-words font-mono text-sm text-gray-800 dark:text-gray-200 w-fit ml-auto">{value}</pre>
        </div>
      </div>
      <div
        class="p-3 bg-gray-50 dark:bg-gray-700/50 text-xs text-gray-500 dark:text-gray-400 border-t border-gray-200 dark:border-gray-700 flex justify-between items-center"
      >
        <div class="flex space-x-4">
          <span
            >Press <kbd
              class="px-2 py-1 bg-gray-200 dark:bg-gray-600 rounded text-xs font-mono"
              >Space</kbd
            > to close</span
          >
          <span
            ><kbd
              class="px-2 py-1 bg-gray-200 dark:bg-gray-600 rounded text-xs font-mono"
              >Shift + Click</kbd
            > to copy</span
          >
        </div>
        <button
          on:click|preventDefault={(e) =>
            e.shiftKey ? copyToClipboard() : dispatch("close")}
          class="text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200"
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            class="h-4 w-4"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M6 18L18 6M6 6l12 12"
            />
          </svg>
        </button>
      </div>
    </div>
  </div>
{/if}
