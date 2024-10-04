<script lang="ts">
  import HideSidebar from "@rilldata/web-common/components/icons/HideSidebar.svelte";
  import SlidingWords from "@rilldata/web-common/components/tooltip/SlidingWords.svelte";
  import { workspaces } from "./workspace-stores";
  import { navigationOpen } from "../navigation/Navigation.svelte";
  import { scale } from "svelte/transition";
  import { cubicOut } from "svelte/easing";
  import HideBottomPane from "@rilldata/web-common/components/icons/HideBottomPane.svelte";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import {
    resourceIconMapping,
    resourceColorMapping,
  } from "@rilldata/web-common/features/entity-management/resource-icon-mapping";
  import type { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import InputWithConfirm from "@rilldata/web-common/components/forms/InputWithConfirm.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";

  export let resourceKind: ResourceKind | "file";
  export let titleInput: string;
  export let editable = true;
  export let showInspectorToggle = true;
  export let showTableToggle = false;
  export let hasUnsavedChanges: boolean;
  export let filePath: string;
  export let onTitleChange: (title: string) => void = () => {};

  let width: number;

  $: value = titleInput;
  $: workspaceLayout = workspaces.get(filePath);
  $: inspectorVisible = workspaceLayout.inspector.visible;
  $: tableVisible = workspaceLayout.table.visible;
</script>

<header class="slide" bind:clientWidth={width} class:!pl-12={!$navigationOpen}>
  <div class="flex gap-x-0 items-center">
    <svelte:component
      this={resourceIconMapping[resourceKind]}
      size="19px"
      color={resourceColorMapping[resourceKind]}
    />

    <div class="wrapper slide">
      <InputWithConfirm
        size="md"
        {editable}
        id="model-title-input"
        textClass="text-xl font-semibold"
        {value}
        onConfirm={onTitleChange}
      />

      {#if hasUnsavedChanges}
        <span
          class="w-1.5 h-1.5 bg-gray-300 rounded flex-none"
          transition:scale={{ duration: 200, easing: cubicOut }}
        />
      {/if}
    </div>
  </div>

  <div class="flex items-center gap-x-2 w-fit">
    <slot name="workspace-controls" {width} />

    <div class="flex-none">
      <slot name="cta" {width} />
    </div>

    {#if showTableToggle}
      <Tooltip distance={8}>
        <Button
          type="secondary"
          square
          selected={$tableVisible}
          on:click={workspaceLayout.table.toggle}
        >
          <HideBottomPane size="18px" open={$tableVisible} />
        </Button>
        <TooltipContent slot="tooltip-content">
          <SlidingWords active={$tableVisible} reverse>
            results preview
          </SlidingWords>
        </TooltipContent>
      </Tooltip>
    {/if}

    {#if showInspectorToggle}
      <Button
        type="secondary"
        square
        selected={$inspectorVisible}
        on:click={workspaceLayout.inspector.toggle}
      >
        <HideSidebar open={$inspectorVisible} size="18px" />

        <svelte:fragment slot="tooltip-content">
          <SlidingWords
            active={$inspectorVisible}
            direction="horizontal"
            reverse
          >
            inspector
          </SlidingWords>
        </svelte:fragment>
      </Button>
    {/if}
  </div>
</header>

<style lang="postcss">
  header {
    @apply bg-gray-100 px-4 w-full;
    @apply justify-between;
    @apply flex flex-none gap-x-2  py-2;
  }

  .wrapper {
    @apply w-fit gap-x-2;
    @apply relative flex items-center;
    @apply font-bold pr-2 self-start pl-1;
    font-size: 16px;
  }
</style>
