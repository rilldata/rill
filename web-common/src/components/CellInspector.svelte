<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import { fly } from "svelte/transition";
  import { formatInteger } from "../lib/formatters";
  import { copyToClipboard } from "@rilldata/web-common/lib/actions/copy-to-clipboard";
  import { cellInspectorStore } from "../features/dashboards/stores/cell-inspector-store";
  import { cubicOut } from "svelte/easing";
  import Kbd from "./Kbd.svelte";

  export let value: any = null;
  export let isOpen: boolean = false;

  let hovered = false;
  let hoveredValue: string | null = null;
  let container: HTMLElement;
  let content: HTMLElement;
  let copied = false;
  let isJson = false;
  let parsedJson: any = null;

  const isMac = window.navigator.userAgent.includes("Macintosh");

  // Subscribe to the cellInspectorStore to keep the component in sync
  const unsubscribe = cellInspectorStore.subscribe((state) => {
    isOpen = state.isOpen;
    if (state.value && state.isOpen) {
      value = state.value;
    }
  });

  function handleKeyDown(event: KeyboardEvent) {
    // Only handle Space key when not in an input, textarea, or other form element
    const target = event.target as HTMLElement;
    const tagName = target.tagName.toLowerCase();
    const isFormElement =
      tagName === "input" || tagName === "textarea" || tagName === "select";

    if (event.code === "Space" && !event.repeat && !isFormElement) {
      event.preventDefault();
      event.stopPropagation();
      cellInspectorStore.toggle(value);
    } else if (event.key === "Escape" && isOpen) {
      event.preventDefault();
      event.stopPropagation();
      cellInspectorStore.close();
    } else if (
      (event.metaKey || event.ctrlKey) &&
      event.key === "c" &&
      isOpen
    ) {
      event.preventDefault();
      event.stopPropagation();
      onCopy();
    }
  }

  function handleClickOutside(event: MouseEvent) {
    if (isOpen && container && !container.contains(event.target as Node)) {
      cellInspectorStore.close();
    }
  }

  // FIXME: Hoist the keyboard event listener to the top level; centralize the hotkeys
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

  // Clean up the subscription
  onDestroy(() => {
    unsubscribe();
  });

  export function formatValue(value: any): string {
    // If the value is null or undefined, return an empty string
    if (value === null || value === undefined) {
      return "";
    }

    // If the value is JSON, pretty print it
    if (isJson && parsedJson !== null) {
      return JSON.stringify(parsedJson, null, 2);
    }

    // Handle both number type and string numbers
    if (typeof value === "number") {
      return formatInteger(value);
    }

    // For strings, check if it's a valid number without leading zeros
    if (typeof value === "string") {
      const num = Number(value);
      // Only format if it's a valid number and doesn't have leading zeros
      if (!isNaN(num) && value.trim() === String(num)) {
        return formatInteger(num);
      }
    }

    // For all other cases, return as string
    return String(value);
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

  // Parse the value as JSON if it is a valid JSON string
  $: if (isOpen) {
    try {
      parsedJson = JSON.parse(value);
      isJson = typeof parsedJson === "object" && parsedJson !== null;
    } catch {
      parsedJson = null;
      isJson = false;
    }
  }
</script>

{#if isOpen}
  <div
    bind:this={container}
    class="cell-inspector fixed top-12 right-4 z-50 transition-opacity shadow-lg rounded-lg border border-gray-200 dark:border-gray-700 bg-surface dark:bg-gray-800"
    class:invisible={!isOpen && !hovered}
    class:opacity-0={!isOpen && !hovered}
    class:opacity-100={isOpen || hovered}
    role="dialog"
    aria-labelledby="cell-inspector-title"
    aria-describedby="cell-inspector-content"
    aria-modal="true"
    transition:fly={{ duration: 200, x: 200, easing: cubicOut }}
  >
    <div
      class="w-full min-w-64 max-w-2xl max-h-[80vh] overflow-hidden flex flex-col rounded-lg"
      role="document"
      bind:this={content}
    >
      <div
        class="flex justify-between p-2 border-gray-200 dark:border-gray-700 gap-1 overflow-y-auto"
        class:items-start={isJson}
        class:items-center={!isJson}
      >
        {#if value === null}
          <span class="text-sm text-gray-500 dark:text-gray-400 italic"
            >No value</span
          >
        {:else}
          <span
            class="whitespace-pre-wrap break-words text-sm text-gray-700 dark:text-gray-200 flex-1"
            class:font-mono={isJson}>{formatValue(value)}</span
          >
        {/if}
      </div>
      <div
        class="flex justify-between p-2 border-t border-gray-200 gap-1 text-[11px] text-gray-500"
      >
        {#if !copied}
          <span>
            <Kbd>{isMac ? "⌘" : "Ctrl"}</Kbd>
            <Kbd>C</Kbd>
            to copy</span
          >
        {:else}
          <span>Copied</span>
        {/if}
        <span>
          <Kbd>Space</Kbd>
          to close</span
        >
      </div>
    </div>
  </div>
{/if}
