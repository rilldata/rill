<script lang="ts">
  import EditIcon from "@rilldata/web-local/lib/components/icons/EditIcon.svelte";
  import ModelIcon from "@rilldata/web-local/lib/components/icons/Model.svelte";
  import Tooltip from "@rilldata/web-local/lib/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-local/lib/components/tooltip/TooltipContent.svelte";
  import WorkspaceHeaderStatusSpinner from "./WorkspaceHeaderStatusSpinner.svelte";

  export let onChangeCallback;
  export let titleInput;
  export let showStatus = true;

  let titleInputElement;
  let editingTitle = false;
  let titleInputValue;
  let tooltipActive;

  function onKeydown(event) {
    if (editingTitle && event.key === "Enter") {
      titleInputElement.blur();
    }
  }

  $: inputSize =
    Math.max((editingTitle ? titleInputValue : titleInput)?.length || 0, 5) + 1;
</script>

<svelte:window on:keydown={onKeydown} />

<header
  style:height="var(--header-height)"
  class="grid items-center content-stretch justify-between bg-gray-100 pl-6 pr-6"
  style:grid-template-columns="[title] auto [controls] auto"
>
  <div>
    {#if titleInput !== undefined && titleInput !== null}
      <h1
        style:font-size="16px"
        class="grid grid-flow-col justify-start items-center gap-x-1"
      >
        <slot name="icon">
          <ModelIcon />
        </slot>

        <Tooltip
          distance={8}
          bind:active={tooltipActive}
          suppress={editingTitle}
        >
          <input
            autocomplete="off"
            id="model-title-input"
            bind:this={titleInputElement}
            on:input={(evt) => {
              titleInputValue = evt.target.value;
              editingTitle = true;
            }}
            class="bg-gray-100 border border-transparent border-2 hover:border-gray-400 rounded pl-2 pr-2 cursor-pointer"
            class:ui-copy-strong={editingTitle === false}
            on:blur={() => {
              editingTitle = false;
            }}
            value={titleInput}
            size={inputSize}
            on:change={onChangeCallback}
          />
          <TooltipContent slot="tooltip-content">
            <div class="flex items-center gap-x-2">
              <EditIcon size=".75em" />Edit
            </div>
          </TooltipContent>
        </Tooltip>
      </h1>
    {/if}
  </div>
  <div>
    <slot name="right" />
    {#if showStatus}
      <WorkspaceHeaderStatusSpinner />
    {/if}
  </div>
</header>
