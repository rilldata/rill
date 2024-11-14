<script lang="ts">
  import HideSidebar from "@rilldata/web-common/components/icons/HideSidebar.svelte";
  import SlidingWords from "@rilldata/web-common/components/tooltip/SlidingWords.svelte";
  import { workspaces } from "./workspace-stores";
  import { navigationOpen } from "../navigation/Navigation.svelte";
  import HideBottomPane from "@rilldata/web-common/components/icons/HideBottomPane.svelte";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import {
    resourceColorMapping,
    resourceIconMapping,
  } from "@rilldata/web-common/features/entity-management/resource-icon-mapping";
  import InputWithConfirm from "@rilldata/web-common/components/forms/InputWithConfirm.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import type { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { Settings } from "lucide-svelte";
  import File from "@rilldata/web-common/components/icons/File.svelte";

  export let resourceKind: ResourceKind | undefined;
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

<header class="slide" bind:clientWidth={width} class:!pl-10={!$navigationOpen}>
  <div class="flex gap-x-0 items-center">
    <svelte:component
      this={resourceKind
        ? resourceIconMapping[resourceKind]
        : filePath === "/.env" || filePath === "/rill.yaml"
          ? Settings
          : File}
      size="19px"
      color={resourceKind ? resourceColorMapping[resourceKind] : "#9CA3AF"}
    />

    <InputWithConfirm
      size="md"
      {editable}
      id="model-title-input"
      textClass="text-xl font-semibold"
      {value}
      onConfirm={onTitleChange}
      showIndicator={hasUnsavedChanges}
    />
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
      <Tooltip distance={8}>
        <Button
          type="secondary"
          square
          selected={$inspectorVisible}
          on:click={workspaceLayout.inspector.toggle}
        >
          <HideSidebar open={$inspectorVisible} size="18px" />
        </Button>

        <TooltipContent slot="tooltip-content">
          <SlidingWords
            active={$inspectorVisible}
            direction="horizontal"
            reverse
          >
            inspector
          </SlidingWords>
        </TooltipContent>
      </Tooltip>
    {/if}
  </div>
</header>

<style lang="postcss">
  header {
    @apply px-4 w-full;
    @apply justify-between;
    @apply flex flex-none py-2;
  }
</style>
