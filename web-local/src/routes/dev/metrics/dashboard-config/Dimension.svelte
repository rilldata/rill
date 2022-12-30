<script lang="ts">
  import Hide from "@rilldata/web-common/components/icons/Hide.svelte";
  import Show from "@rilldata/web-common/components/icons/Show.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { createEventDispatcher } from "svelte";
  import Cell from "../Cell.svelte";
  import Row from "../core/Row.svelte";
  import { createInputManagementProvider } from "../input-management-provider";

  export let displayName;
  export let column;
  export let description;

  export let selected = false;
  export let mode = "unselected";
  export let visible = true;
  export let error: string = undefined;

  export let isDragging = false;

  const dispatch = createEventDispatcher();

  let expressionElement: HTMLElement;
  let displayNameElement: HTMLElement;
  let descriptionElement: HTMLElement;
  let columnElement: HTMLElement;

  /** this provider gives us a function to blur all corresponding fields used */
  const inp = createInputManagementProvider();

  let row;

  export function blurAllFields() {
    displayNameElement?.blur();
    expressionElement?.blur();
    descriptionElement?.blur();
  }

  export function deactivateDragHandleMenu() {
    row.deactivateDragHandleMenu();
  }

  $: if (mode === "multiselect") {
    blurAllFields();
  }

  let level: "warning" | "info";
  $: level = error ? "warning" : undefined;

  function handleEdit(key) {
    return (event) => dispatch("edit", { key, value: event.target.value });
  }
</script>

<div
  class="container w-max"
  class:relative={selected || error}
  class:z-10={selected || error}
>
  <Row
    bind:this={row}
    {mode}
    {selected}
    {error}
    suppressTooltips={isDragging}
    on:move-to-top
    on:move-to-bottom
    on:move-up
    on:move-down
    on:delete
    on:select
    on:draghandle-mousedown
    let:handleFocus
  >
    <div class="input-container flex items-center" class:disabled={!visible}>
      <!-- displayName -->
      <Cell {level}>
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
          class:select-none={mode === "multiselect"}
          bind:this={displayNameElement}
          value={displayName || ""}
          style:width="280px"
          on:focus={handleFocus}
          placeholder="display name"
          on:input={handleEdit("displayName")}
        />
      </Cell>
      <Cell {level}>
        <input
          class:select-none={mode === "multiselect"}
          bind:this={columnElement}
          value={column || ""}
          style:width="280px"
          on:focus={handleFocus}
          placeholder="dimension column"
          on:input={handleEdit("column")}
        />
      </Cell>
      <!-- description -->
      <Cell {level}>
        <input
          class:select-none={mode === "multiselect"}
          bind:this={descriptionElement}
          value={description || ""}
          style:width="280px"
          on:focus={handleFocus}
          placeholder="description"
          on:input={handleEdit("description")}
        />
      </Cell>
    </div>
  </Row>
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
    @apply ring-0 outline-0;
  }

  .input-container.disabled input {
    color: gray;
    font-style: italic;
  }
</style>
