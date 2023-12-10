<script lang="ts">
  import Column from "@rilldata/web-common/components/icons/Column.svelte";
  import Row from "@rilldata/web-common/components/icons/Row.svelte";
  import DragList from "./DragList.svelte";

  let items2 = [
    { id: 6, title: "Palpatine" },
    { id: 7, title: "Death" },
  ];

  let items3 = [
    { id: 8, title: "Emperor" },
    { id: 9, title: "Star" },
  ];
</script>

<div class="header">
  <div class="header-row">
    <Column size="16px" /> Columns
    <DragList items={items2} style="horizontal" />
  </div>
  <div class="header-row">
    <Row size="16px" /> Rows
    <DragList items={items3} style="horizontal" />
  </div>
</div>

<style lang="postcss">
  .header {
    border-bottom: 1px solid #ddd;
    display: flex;
    flex-direction: column;
  }
  .header-row {
    @apply flex items-center gap-x-1 p-2;
  }
</style>
