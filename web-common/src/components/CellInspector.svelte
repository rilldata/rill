<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import { createEventDispatcher } from "svelte";
  import { formatInteger } from "../lib/formatters";
  import CopyIcon from "@rilldata/web-common/components/icons/CopyIcon.svelte";
  import Check from "@rilldata/web-common/components/icons/Check.svelte";
  import { copyToClipboard } from "@rilldata/web-common/lib/actions/copy-to-clipboard";
  import { cellInspectorStore } from "../features/dashboards/stores/cell-inspector-store";

  export let value: string = "";
  export let isOpen: boolean = false;

  let hovered = false;
  let hoveredValue: string | null = null;
  let container: HTMLElement;
  let content: HTMLElement;
  let copied = false;

  // Subscribe to the cellInspectorStore to keep the component in sync
  const unsubscribe = cellInspectorStore.subscribe((state) => {
    isOpen = state.isOpen;
    if (state.value) {
      value = state.value;
    }
  });

  const dispatch = createEventDispatcher();

  const iconColor = "#15141A";

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

  function onCopy() {
    copyToClipboard(value, undefined, false);
    copied = true;

    setTimeout(() => {
      copied = false;
    }, 2_000);
  }
</script>

{#if isOpen}
  <div
    bind:this={container}
    class="cell-inspector fixed top-4 right-4 z-50 transition-opacity shadow-lg rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800"
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
        class="flex items-center justify-between p-2 border-gray-200 dark:border-gray-700 gap-1"
        id="cell-inspector-content"
      >
        <div class="flex items-center" id="cell-inspector-title">
          <pre
            class="whitespace-pre-wrap break-words font-mono text-sm text-gray-800 dark:text-gray-200 flex-1">{formatValue(
              value,
            )}</pre>
        </div>
        <button
          class="hover:bg-slate-100 rounded p-1 active:bg-slate-200 group"
          on:click={onCopy}
        >
          {#if copied}
            <Check size="14px" color={iconColor} />
          {:else}
            <CopyIcon size="14px" className="text-gray-500" />
          {/if}
        </button>
      </div>
    </div>
  </div>
{/if}
