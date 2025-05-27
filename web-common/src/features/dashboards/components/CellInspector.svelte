<script lang="ts">
  import { onDestroy } from "svelte";
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
    cellInspectorStore.toggle(value);
  }

  // No need for keyboard event handler here - it's handled in the base CellInspector component

  onDestroy(unsubscribe);
</script>

<CellInspector
  bind:isOpen
  {value}
  on:close={handleClose}
  on:toggle={handleToggle}
  on:open={() => cellInspectorStore.open(value)}
/>
