<script lang="ts">
  import { LIST_SLIDE_DURATION } from "@rilldata/web-local/lib/application-config";
  import Callout from "@rilldata/web-local/lib/components/callout/Callout.svelte";
  import Hide from "@rilldata/web-local/lib/components/icons/Hide.svelte";
  import Show from "@rilldata/web-local/lib/components/icons/Show.svelte";
  import { Divider, MenuItem } from "@rilldata/web-local/lib/components/menu";
  import Tooltip from "@rilldata/web-local/lib/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-local/lib/components/tooltip/TooltipContent.svelte";
  import { createEventDispatcher } from "svelte";
  import { slide } from "svelte/transition";
  import PrimaryCell from "./PrimaryCell.svelte";
  import SecondaryCell from "./SecondaryCell.svelte";
  import WithDragHandle from "./WithDragHandle.svelte";

  export let expression: string;
  export let displayName: string;
  export let description: string;
  export let format;
  export let selected = false;
  export let mode = "unselected";
  export let visible = true;
  export let error: string = undefined;

  export let isDragging = false;

  const dispatch = createEventDispatcher();

  let expressionElement: HTMLElement;
  let displayNameElement: HTMLElement;
  let descriptionElement: HTMLElement;
  let formatterElement: HTMLElement;

  export function blurAllFields() {
    displayNameElement?.blur();
    expressionElement?.blur();
    descriptionElement?.blur();
  }

  export function currentFocuse() {
    if (document.activeElement === displayNameElement) return "display";
    if (document.activeElement === expressionElement) return "expression";
    if (document.activeElement === descriptionElement) return "description";
    return undefined;
  }

  export function elementInFocus() {
    return (
      document.activeElement === displayNameElement ||
      document.activeElement === expressionElement ||
      document.activeElement === descriptionElement
    );
  }

  export function select() {
    dispatch("select");
  }

  export function deactivateDragHandleMenu() {
    dragMenuActive = false;
  }

  function handleEdit(key) {
    return (event) => dispatch("edit", { key, value: event.target.value });
  }

  $: if (mode === "multiselect") {
    blurAllFields();
  }

  let isEditing = false;
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
      suppressTooltips={isDragging}
      on:mousedown={() => {
        dispatch("select", false);
        dispatch("draghandle-mousedown");
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
      <div class:ring-1={error} class:ring-yellow-500={error} class="w-full">
        <div
          class="input-container flex items-center"
          class:disabled={!visible}
          class:bg-gray-50={selected && !error}
          class:bg-yellow-50={error}
          class:ring-2={selected && !isEditing}
          class:ring-1={selected && isEditing}
        >
          <PrimaryCell>
            <Tooltip location="bottom" alignment="start" distance={8}>
              <button
                class="pl-2 pr-2 show-hide"
                class:opacity-0={visible}
                class:text-gray-400={!visible}
                style:font-size="20px"
                on:click={() => dispatch("toggle-visibility")}
              >
                {#if visible}
                  <Show />
                {:else}
                  <Hide />
                {/if}
              </button>
              <TooltipContent slot="tooltip-content">
                toggle visibility in dashboard
              </TooltipContent>
            </Tooltip>
            <input
              bind:this={expressionElement}
              class="resize-none bg-transparent box-border pl-0 w-full"
              value={expression}
              placeholder="DuckDB SQL expression like count(*), sum(revenue), etc."
              readonly={mode === "multiselect"}
              on:focus={(event) => {
                isEditing = true;
              }}
              on:input={handleEdit("expression")}
            />
          </PrimaryCell>
          <!-- displayName -->
          <SecondaryCell>
            <input
              bind:this={displayNameElement}
              value={displayName || ""}
              style:width="280px"
              on:focus={(event) => {
                isEditing = true;
              }}
              placeholder="display name"
              on:input={handleEdit("displayName")}
            />
          </SecondaryCell>
          <!-- description -->
          <SecondaryCell>
            <input
              bind:this={descriptionElement}
              value={description || ""}
              style:width="280px"
              on:focus={(event) => {
                dispatch("focus", "description");
                isEditing = true;
              }}
              placeholder="description"
              on:input={handleEdit("description")}
            />
          </SecondaryCell>
          <!-- formatter -->
          <SecondaryCell>
            <input
              bind:this={formatterElement}
              value={format || ""}
              on:input={(event) => dispatch("formatter")}
            />
          </SecondaryCell>
        </div>
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

<style lang="postcss">
  .container:hover .show-hide {
    opacity: 1;
  }

  .show-hide:focus {
    opacity: 1;
  }

  input {
    min-height: 36px;
    background-color: transparent;
    @apply px-4 py-2;
  }

  input:focus {
    @apply ring-1 outline-0;
  }

  .input-container.disabled input {
    color: gray;
    font-style: italic;
  }
</style>
