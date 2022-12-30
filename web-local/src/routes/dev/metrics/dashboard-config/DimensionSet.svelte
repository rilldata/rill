<script lang="ts">
  import { guidGenerator } from "@rilldata/web-common/lib/guid";
  import ContainerTitle from "../core/ContainerTitle.svelte";
  import RowContainer from "../core/RowContainer.svelte";
  import Dimension from "./Dimension.svelte";

  let errorGUID = guidGenerator();

  let dimensions = [
    {
      displayName: "User ID",
      description: "",
      column: "user_id",
      id: guidGenerator(),
      visible: true,
    },
    ...Array.from({ length: 0 }).map((_, i) => {
      return {
        displayName: "Dimension " + i,
        description: "",
        column: `dimension=${i}`,
        id: guidGenerator(),
        visible: true,
      };
    }),
    {
      displayName: "Country",
      description: "",
      column: "country",
      id: errorGUID,
      visible: true,
    },
    {
      displayName: "Language",
      description: "",
      column: "language",
      id: guidGenerator(),
      visible: true,
    },
  ];
</script>

<ContainerTitle>Dimensions</ContainerTitle>

<RowContainer
  items={dimensions}
  addItemText="Add Dimension"
  on:update-items={(event) => {
    dimensions = event.detail;
  }}
  let:item
  let:edit
  let:moveUp
  let:moveDown
  let:moveToTop
  let:moveToBottom
  let:deleteItem
  let:select
  let:toggleVisibility
  let:selected
  let:mode
  let:isDragging
  let:dragHandleMousedown
>
  <Dimension
    visible={item.visible}
    displayName={item.displayName}
    description={item.description}
    column={item.column}
    {selected}
    {mode}
    {isDragging}
    on:draghandle-mousedown={dragHandleMousedown}
    on:select={select}
    on:toggle-visibility={toggleVisibility}
    on:edit={edit}
    on:delete={deleteItem}
    on:move-up={moveUp}
    on:move-down={moveDown}
    on:move-to-top={moveToTop}
    on:move-to-bottom={moveToBottom}
  />
</RowContainer>
