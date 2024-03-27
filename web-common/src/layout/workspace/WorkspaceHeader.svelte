<script lang="ts">
  import { page } from "$app/stores";
  import { IconButton } from "@rilldata/web-common/components/button";
  import HideRightSidebar from "@rilldata/web-common/components/icons/HideRightSidebar.svelte";
  import SlidingWords from "@rilldata/web-common/components/tooltip/SlidingWords.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { dynamicTextInputWidth } from "@rilldata/web-common/lib/actions/dynamic-text-input-width";
  import SourceUnsavedIndicator from "../../features/sources/editor/SourceUnsavedIndicator.svelte";
  import { workspaces } from "./workspace-stores";
  import { navigationOpen } from "../navigation/Navigation.svelte";

  export let onChangeCallback: (
    e: Event & {
      currentTarget: EventTarget & HTMLInputElement;
    },
  ) => void = () => {};
  export let titleInput: string;
  export let editable = true;
  export let showInspectorToggle = true;

  let titleInputElement: HTMLInputElement;
  let editingTitle = false;
  let tooltipActive: boolean;
  let width: number;

  $: workspaceLayout = $workspaces;

  $: visible = workspaceLayout.inspector.visible;

  function onKeydown(
    event: KeyboardEvent & {
      currentTarget: EventTarget & Window;
    },
  ) {
    if (editingTitle && event.key === "Enter") {
      titleInputElement.blur();
    }
  }

  function onInput() {
    if (editable) {
      editingTitle = true;
    }
  }
</script>

<svelte:window on:keydown={onKeydown} />

<header
  class="grid items-center content-stretch justify-between pl-4 border-b border-gray-300"
  style:grid-template-columns="[title] minmax(0, 1fr) [controls] auto"
  style:height="var(--header-height)"
  bind:clientWidth={width}
>
  <div class:pl-4={!$navigationOpen} class="slide">
    {#if titleInput !== undefined && titleInput !== null}
      <h1
        style:font-size="16px"
        class="grid grid-flow-col justify-start items-center gap-x-1 overflow-hidden"
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
            class="bg-transparent border-transparent border-2 rounded pl-2 pr-2"
            class:editable
            class:font-bold={editingTitle === false}
            value={titleInput}
            on:input={onInput}
            on:change={onChangeCallback}
            on:focus={() => {
              editingTitle = true;
            }}
            on:blur={() => {
              editingTitle = false;
            }}
            use:dynamicTextInputWidth
            bind:this={titleInputElement}
          />
          <TooltipContent slot="tooltip-content">
            <div class="flex items-center gap-x-2">Edit</div>
          </TooltipContent>
        </Tooltip>

        {#if $page.url.pathname.startsWith("/source")}
          <SourceUnsavedIndicator sourceName={titleInput} />
        {/if}
      </h1>
    {/if}
  </div>

  <div class="flex items-center mr-4">
    <slot name="workspace-controls" {width} />
    {#if showInspectorToggle}
      <IconButton on:click={workspaceLayout.inspector.toggle}>
        <span class="text-gray-500">
          <HideRightSidebar size="18px" />
        </span>
        <svelte:fragment slot="tooltip-content">
          <SlidingWords active={$visible} direction="horizontal" reverse
            >inspector</SlidingWords
          >
        </svelte:fragment>
      </IconButton>
    {/if}

    <div class="pl-4">
      <slot name="cta" {width} />
    </div>
  </div>
</header>

<style lang="postcss">
  .editable {
    @apply cursor-pointer;
  }

  .editable:hover {
    @apply border-gray-400;
  }
</style>
