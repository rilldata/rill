<script lang="ts">
  import { page } from "$app/stores";
  import HideRightSidebar from "@rilldata/web-common/components/icons/HideRightSidebar.svelte";
  import SlidingWords from "@rilldata/web-common/components/tooltip/SlidingWords.svelte";
  import { workspaces } from "./workspace-stores";
  import { navigationOpen } from "../navigation/Navigation.svelte";
  import { scale } from "svelte/transition";
  import { cubicOut } from "svelte/easing";
  import HideBottomPane from "@rilldata/web-common/components/icons/HideBottomPane.svelte";
  import Button from "@rilldata/web-common/components/button/Button.svelte";

  export let titleInput: string;
  export let editable = true;
  export let showInspectorToggle = true;
  export let showTableToggle = false;
  export let hasUnsavedChanges: boolean;

  let width: number;
  let titleWidth: number;

  $: value = titleInput;
  $: context = $page.url.pathname;
  $: workspaceLayout = workspaces.get(context);
  $: visible = workspaceLayout.inspector.visible;
  $: tableVisible = workspaceLayout.table.visible;
</script>

<header class="slide" bind:clientWidth={width}>
  <div class="flex justify-end items-center mr-4 gap-x-2 flex-none w-full">
    <slot name="workspace-controls" {width} />

    <div class="pl-4 flex-none">
      <slot name="cta" {width} />
    </div>

    {#if showTableToggle}
      <Button
        type="secondary"
        square
        selected={true}
        on:click={workspaceLayout.table.toggle}
      >
        <HideBottomPane size="18px" />

        <svelte:fragment slot="tooltip-content">
          <SlidingWords active={$tableVisible} reverse>
            results preview
          </SlidingWords>
        </svelte:fragment>
      </Button>
    {/if}

    {#if showInspectorToggle}
      <Button
        type="secondary"
        square
        selected={true}
        on:click={workspaceLayout.inspector.toggle}
      >
        <HideRightSidebar size="18px" />

        <svelte:fragment slot="tooltip-content">
          <SlidingWords active={$visible} direction="horizontal" reverse>
            inspector
          </SlidingWords>
        </svelte:fragment>
      </Button>
    {/if}
  </div>

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

    {#if hasUnsavedChanges}
      <span
        class="w-1.5 h-1.5 bg-gray-300 rounded flex-none"
        transition:scale={{ duration: 200, easing: cubicOut }}
      />
    {/if}
  </div>
</header>

<style lang="postcss">
  input:focus,
  input:not(:disabled):hover {
    @apply border-primary-500 outline-none;
    @apply ring-2 ring-primary-100 bg-background;
  }

  input,
  label {
    @apply whitespace-pre rounded-[2px] border border-transparent;
    @apply truncate font-semibold text-xl bg-gray-100;
  }

  input {
    @apply absolute text-left;
    text-indent: 6px;
  }

  label {
    @apply w-fit min-w-8 px-2 text-transparent max-w-full;
  }

  header {
    @apply bg-gray-100;
    @apply justify-between;
    @apply flex flex-col items-center pl-4 pt-2 overflow-hidden;
    @apply h-20;
  }

  .wrapper {
    @apply overflow-hidden max-w-full gap-x-2;
    @apply size-full relative flex items-center;
    @apply font-bold pr-2 self-start pl-1;
    font-size: 16px;
  }
</style>
