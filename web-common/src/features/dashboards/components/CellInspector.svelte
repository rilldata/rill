<script lang="ts">
  import { onDestroy } from "svelte";
  import { cellInspectorStore } from "../stores/cellInspectorStore";
  import CellInspector from "../../../components/CellInspector.svelte";

  let isOpen = false;
  let value = "";

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

  onDestroy(unsubscribe);
</script>

<CellInspector
  bind:isOpen
  {value}
  on:close={handleClose}
  on:toggle={handleToggle}
  on:open={() => cellInspectorStore.open(value)}
/>
