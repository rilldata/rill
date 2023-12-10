<script lang="ts">
  import DragList from "@rilldata/web-common/features/dashboards/pivot/DragList.svelte";

  let items1 = [
    { id: 1, title: "I" },
    { id: 2, title: "Am" },
    { id: 3, title: "Yoda" },
    { id: 4, title: "Luke" },
    { id: 5, title: "Skywalker" },
  ];
</script>

<div class="sidebar">
  <h2>MEASURES</h2>
  <!-- List of measures -->
  <h2>DIMENSIONS</h2>

  <DragList items={items1} style="vertical" />
</div>

<style lang="postcss">
  .sidebar {
    @apply bg-slate-50;
    width: 200px;
    padding: 10px;
    height: 100%;
    flex: 0 0 200px;
    padding: 1rem;
    overflow-y: auto;
  }
</style>
