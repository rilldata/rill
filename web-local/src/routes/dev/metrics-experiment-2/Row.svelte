<!-- @component
Row.svelte enables functionality around moving the row among other rows.

-->
<script lang="ts">
  import { Divider, MenuItem } from "@rilldata/web-local/lib/components/menu";
  import { createEventDispatcher } from "svelte";
  import WithDragHandle from "./WithDragHandle.svelte";
  export let mode = "unselected";

  export let suppressTooltips = false;

  const dispatch = createEventDispatcher();

  export function select() {
    dispatch("select");
  }

  export function deactivateDragHandleMenu() {
    dragMenuActive = false;
  }

  let dragMenuActive = false;
</script>

<div
  class="container flex items-center bg-white"
  class:user-select-none={mode === "multiselect"}
  on:click={(event) => {
    dispatch("select", event.shiftKey);
  }}
>
  <div class="w-full">
    <WithDragHandle
      {suppressTooltips}
      on:mousedown={(event) => {
        dispatch("select", false);
        dispatch("draghandle-mousedown", {
          x: event.clientX,
          y: event.clientY,
        });
      }}
      bind:active={dragMenuActive}
    >
      <svelte:fragment slot="menu-items">
        <MenuItem>Add Entry Before</MenuItem>
        <MenuItem>Add Entry After</MenuItem>
        <Divider />
        <MenuItem
          on:select={() => {
            dispatch("move-to-top");
          }}>Move to Top <span slot="right">meta + up</span></MenuItem
        >
        <MenuItem on:select={() => dispatch("move-up")}
          >Move Up <span slot="right">ctrl + up</span></MenuItem
        >
        <MenuItem on:select={() => dispatch("move-down")}
          >Move Down <span slot="right">ctrl + down</span></MenuItem
        >
        <MenuItem on:select={() => dispatch("move-to-bottom")}
          >Move To Bottom <span slot="right">meta + down</span></MenuItem
        >
        <Divider />
        <MenuItem>hide</MenuItem>
        <MenuItem
          on:select={() => {
            dispatch("delete");
          }}>delete measure</MenuItem
        >
      </svelte:fragment>
      <slot />
    </WithDragHandle>
  </div>
</div>
