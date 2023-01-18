<script lang="ts">
  import type { LayoutElement } from "@rilldata/web-local/lib/types";
  import { getContext } from "svelte";
  import type { Tweened } from "svelte/motion";
  import type { Writable } from "svelte/store";

  import { IconButton } from "@rilldata/web-common/components/button";
  import HideRightSidebar from "@rilldata/web-common/components/icons/HideRightSidebar.svelte";
  import SlidingWords from "@rilldata/web-common/components/tooltip/SlidingWords.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { createResizeListenerActionFactory } from "@rilldata/web-common/lib/actions/create-resize-listener-factory";
  import WorkspaceHeaderStatusSpinner from "./WorkspaceHeaderStatusSpinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/lib/entity";

  export let onChangeCallback;
  export let titleInput;
  export let appRunning = true;
  export let editable = true;
  export let showInspectorToggle = true;
  export let width: number = undefined;

  let titleInputElement;
  let editingTitle = false;

  let titleInputValue;
  let tooltipActive;

  const { listenToNodeResize, observedNode } =
    createResizeListenerActionFactory();

  const inspectorLayout = getContext(
    "rill:app:inspector-layout"
  ) as Writable<LayoutElement>;

  const navigationVisibilityTween = getContext(
    "rill:app:navigation-visibility-tween"
  ) as Tweened<number>;

  function onKeydown(event) {
    if (editingTitle && event.key === "Enter") {
      titleInputElement.blur();
    }
  }

  $: applicationStatus = appRunning ? EntityStatus.Running : EntityStatus.Idle;

  $: inputSize =
    Math.max((editingTitle ? titleInputValue : titleInput)?.length || 0, 5) + 1;

  $: width = $observedNode?.getBoundingClientRect()?.width;
</script>

<svelte:window on:keydown={onKeydown} />
<header
  class="grid items-center content-stretch justify-between pl-4 border-b border-gray-300"
  style:grid-template-columns="[title] auto [controls] auto"
  style:height="var(--header-height)"
  use:listenToNodeResize
>
  <div style:padding-left="{$navigationVisibilityTween * 24}px">
    {#if titleInput !== undefined && titleInput !== null}
      <h1
        style:font-size="16px"
        class="w-full  overflow-x-hidden  grid grid-flow-col justify-start items-center gap-x-1"
      >
        <Tooltip
          distance={8}
          alignment="start"
          bind:active={tooltipActive}
          suppress={editingTitle || !editable}
        >
          <input
            autocomplete="off"
            disabled={!editable}
            id="model-title-input"
            class:text-overflow-ellipsis={!editable}
            bind:this={titleInputElement}
            on:focus={() => {
              editingTitle = true;
              titleInputValue = titleInput;
            }}
            on:input={(evt) => {
              if (editable) {
                titleInputValue = evt.target.value;
                editingTitle = true;
              }
            }}
            class="w-full text-overflow-ellipses whitespace-wrap bg-transparent border border-transparent border-2 {editable
              ? 'hover:border-gray-400 cursor-pointer'
              : ''} rounded pl-2 pr-2"
            class:font-bold={editingTitle === false}
            on:blur={() => {
              editingTitle = false;
            }}
            value={titleInput}
            size={inputSize}
            on:change={onChangeCallback}
          />
          <TooltipContent slot="tooltip-content">
            <div class="flex items-center gap-x-2">Edit</div>
          </TooltipContent>
        </Tooltip>
      </h1>
    {/if}
  </div>

  <div class="flex items-center mr-4">
    <slot name="workspace-controls" {width} />
    {#if showInspectorToggle}
      <IconButton
        on:click={() => {
          inspectorLayout.update((state) => {
            state.visible = !state.visible;
            return state;
          });
        }}
      >
        <span class="text-gray-500">
          <HideRightSidebar size="18px" />
        </span>
        <svelte:fragment slot="tooltip-content">
          <SlidingWords
            active={$inspectorLayout?.visible}
            direction="horizontal"
            reverse>inspector</SlidingWords
          >
        </svelte:fragment>
      </IconButton>
    {/if}

    <div class="pl-4">
      <slot name="cta" {width} />
    </div>
    <div class="ml-2">
      <WorkspaceHeaderStatusSpinner {applicationStatus} />
    </div>
  </div>
</header>
