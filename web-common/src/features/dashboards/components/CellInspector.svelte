<script lang="ts">
  import { onDestroy, onMount } from "svelte";
  import { cellInspectorStore } from "../stores/cellInspectorStore";
  import CellInspector from "../../../components/CellInspector.svelte";

  let isOpen = false;
  let value = "";
  // We don't need position in this component

  const unsubscribe = cellInspectorStore.subscribe((state) => {
    isOpen = state.isOpen;
    value = state.value;
  });

  function handleClose() {
    cellInspectorStore.close();
  }

  function handleToggle() {
    cellInspectorStore.toggle("");
  }

  // Global keyboard event handler for toggling cell inspector visibility
  function handleGlobalKeyDown(event: KeyboardEvent) {
    // Only handle Space key when not in an input, textarea, or other form element
    const target = event.target as HTMLElement;
    const tagName = target.tagName.toLowerCase();
    const isFormElement =
      tagName === "input" || tagName === "textarea" || tagName === "select";

    if (event.code === "Space" && !event.repeat && !isFormElement) {
      event.preventDefault();
      event.stopPropagation();

      // Toggle the cell inspector visibility
      if (isOpen) {
        cellInspectorStore.close();
      } else if (value) {
        // Only open if we have a value to display
        cellInspectorStore.open(value);
      }
    } else if (event.key === "Escape" && isOpen) {
      event.preventDefault();
      event.stopPropagation();
      cellInspectorStore.close();
    }
  }

  onMount(() => {
    // Add global keyboard event listener
    window.addEventListener("keydown", handleGlobalKeyDown, true);

    return () => {
      // Remove global keyboard event listener
      window.removeEventListener("keydown", handleGlobalKeyDown, true);
    };
  });

  onDestroy(unsubscribe);
</script>

<CellInspector
  bind:isOpen
  {value}
  on:close={handleClose}
  on:toggle={handleToggle}
/>
