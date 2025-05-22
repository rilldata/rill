<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import { cellInspectorStore } from "../stores/cellInspectorStore";
  import CellInspector from "../../../components/CellInspector.svelte";

  let isOpen = false;
  let value = "";
  let position = { x: 0, y: 0 };

  const unsubscribe = cellInspectorStore.subscribe((state) => {
    isOpen = state.isOpen;
    value = state.value;
    position = state.position || { x: 0, y: 0 };
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
