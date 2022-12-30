<!-- @component
Row.svelte enables functionality around moving the row among other rows.

-->
<script lang="ts">
  import { Callout } from "@rilldata/web-common/components/callout";
  import { Divider, MenuItem } from "@rilldata/web-common/components/menu";
  import { LIST_SLIDE_DURATION } from "@rilldata/web-local/lib/application-config";
  import { createEventDispatcher } from "svelte";
  import { slide } from "svelte/transition";
  import WithDragHandle from "./WithDragHandle.svelte";
  export let mode = "unselected";
  export let selected = false;

  export let error: string = undefined;

  export let suppressTooltips = false;

  let isEditing = false;

  const dispatch = createEventDispatcher();

  export function select() {
    dispatch("select");
  }

  export function deactivateDragHandleMenu() {
    dragMenuActive = false;
  }

  function handleEdit(key) {
    return (event) => dispatch("edit", { key, value: event.target.value });
  }

  function handleFocus() {
    isEditing = true;
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
      <div
        class="w-full"
        class:bg-gray-50={selected && !error}
        class:bg-yellow-50={error}
        class:ring-2={selected && !isEditing}
        class:ring-1={(selected && isEditing) || error}
        class:ring-yellow-500={error}
      >
        <slot {handleEdit} {handleFocus} />
        {#if error}
          <div transition:slide|local={{ duration: LIST_SLIDE_DURATION }}>
            <Callout rounded={false} border={false} level="warning"
              >{error}</Callout
            >
          </div>
        {/if}
      </div>
    </WithDragHandle>
  </div>
</div>
