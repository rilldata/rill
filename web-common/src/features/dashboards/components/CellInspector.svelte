<script lang="ts">
  import { onDestroy } from "svelte";
  import { cellInspectorStore } from "../stores/cellInspectorStore";
  import CellInspector from "../../../components/CellInspector.svelte";

  let isOpen = false;
  let value = "";
  // We don't need position in this component as it's handled by the parent CellInspector

  const unsubscribe = cellInspectorStore.subscribe((state) => {
    isOpen = state.isOpen;
    value = state.value;
    // Position is handled by the parent component
  });

  function handleClose() {
    cellInspectorStore.close();
  }

  function handleToggle() {
    cellInspectorStore.toggle("", { x: 0, y: 0 });
  }

  onDestroy(unsubscribe);
</script>

<CellInspector
  bind:isOpen
  {value}
  on:close={handleClose}
  on:toggle={handleToggle}
/>
