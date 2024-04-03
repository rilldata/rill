<script lang="ts">
  import { page } from "$app/stores";
  import { IconButton } from "@rilldata/web-common/components/button";
  import HideRightSidebar from "@rilldata/web-common/components/icons/HideRightSidebar.svelte";
  import SlidingWords from "@rilldata/web-common/components/tooltip/SlidingWords.svelte";
  import { workspaces } from "./workspace-stores";
  import { navigationOpen } from "../navigation/Navigation.svelte";
  import { scale } from "svelte/transition";
  import { cubicOut } from "svelte/easing";

  export let titleInput: string;
  export let editable = true;
  export let showInspectorToggle = true;
  export let isSourceUnsaved = false;

  let width: number;
  let value = titleInput;
  let titleWidth: number;

  $: context = $page.url.pathname;
  $: workspaceLayout = workspaces.get(context);
  $: visible = workspaceLayout.inspector.visible;
</script>

<header class="slide" bind:clientWidth={width}>
  <div class="wrapper slide" class:pl-4={!$navigationOpen}>
    <label aria-hidden for="model-title-input" bind:clientWidth={titleWidth}>
      {value}
    </label>
    <input
      id="model-title-input"
      name="title"
      type="text"
      autocomplete="off"
      spellcheck="false"
      disabled={!editable}
      style:width="{titleWidth}px"
      bind:value
      on:change
      on:keydown={(e) => {
        if (e.key === "Enter") {
          e.preventDefault();
          e.currentTarget.blur();
        }
      }}
      on:blur={() => {
        if (value.length === 0) {
          value = titleInput;
        }
      }}
    />

    {#if context.startsWith("/source") && isSourceUnsaved}
      <span
        class="w-1.5 h-1.5 bg-gray-300 rounded flex-none"
        transition:scale={{ duration: 200, easing: cubicOut }}
      />
    {/if}
  </div>

  <div class="flex items-center mr-4 flex-none">
    <slot name="workspace-controls" {width} />

    {#if showInspectorToggle}
      <IconButton on:click={workspaceLayout.inspector.toggle}>
        <span class="text-gray-500">
          <HideRightSidebar size="18px" />
        </span>
        <svelte:fragment slot="tooltip-content">
          <SlidingWords active={$visible} direction="horizontal" reverse>
            inspector
          </SlidingWords>
        </svelte:fragment>
      </IconButton>
    {/if}

    <div class="pl-4 flex-none">
      <slot name="cta" {width} />
    </div>
  </div>
</header>

<style lang="postcss">
  input:focus,
  input:not(:disabled):hover {
    @apply border-2 border-gray-400 outline-none;
  }

  input,
  label {
    @apply whitespace-pre rounded border-2 border-transparent;
  }

  input {
    @apply absolute text-left;
    text-indent: 6px;
  }

  label {
    @apply w-fit min-w-8 px-2 text-transparent max-w-full;
  }

  header {
    @apply justify-between;
    @apply flex items-center pl-4 border-b border-gray-300 overflow-hidden;
    height: var(--header-height);
  }

  .wrapper {
    @apply overflow-hidden max-w-full gap-x-2;
    @apply size-full relative flex items-center;
    @apply font-bold pr-2 self-start;
    font-size: 16px;
  }
</style>
