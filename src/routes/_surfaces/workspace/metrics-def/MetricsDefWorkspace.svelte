<script lang="ts">
  import SectionDraghandle from "./SectionDragHandle.svelte";

  import { layout } from "$lib/application-state-stores/layout-store";

  import PreviewTable from "$lib/components/table/PreviewTable.svelte";
  let innerHeight;

  const dummyColNames = new Array(10)
    .fill(0)
    .map((x, i) => ({ name: "col_" + i, type: "INTEGER" }));

  const dummyRowsData = new Array(50).fill(0).map((x, i) => {
    return Object.fromEntries(dummyColNames.map((cn, j) => [cn, i * j]));
  });

  const tableContainerDivClass =
    "rounded border border-gray-200 border-2  overflow-auto ";
</script>

<svelte:window bind:innerHeight />

<div class="editor-pane bg-gray-100">
  <div
    style:height="calc({innerHeight}px - {$layout.modelPreviewHeight}px -
    var(--header-height))"
    style="display: flex; flex-flow: column;"
    class="p-6 pt-0"
  >
    <div>
      <div style:height="40px">foo</div>
      <div style:height="40px">foo</div>
    </div>
    <div style:flex="1" class={tableContainerDivClass}>
      <PreviewTable rows={dummyRowsData} columnNames={dummyColNames} />
    </div>
  </div>

  <SectionDraghandle />

  <div style:height="{$layout.modelPreviewHeight}px" class="p-6 ">
    <div class={tableContainerDivClass + " h-full"}>
      <PreviewTable rows={dummyRowsData} columnNames={dummyColNames} />
    </div>
  </div>
</div>
