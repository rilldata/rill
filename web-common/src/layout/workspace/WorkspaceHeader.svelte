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
  import WorkspaceBreadcrumbs from "@rilldata/web-common/features/workspaces/WorkspaceBreadcrumbs.svelte";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";

  export let resourceKind: ResourceKind | undefined;
  export let titleInput: string;
  export let editable = true;
  export let showInspectorToggle = true;
  export let showTableToggle = false;
  export let hasUnsavedChanges: boolean;
  export let filePath: string;
  export let resource: V1Resource | undefined = undefined;
  export let onTitleChange: (title: string) => void = () => {};

  let width: number;
  let editing: boolean;

  $: value = titleInput;
  $: workspaceLayout = workspaces.get(filePath);
  $: inspectorVisible = workspaceLayout.inspector.visible;
  $: tableVisible = workspaceLayout.table.visible;
</script>

<header bind:clientWidth={width}>
  <div
    class="slide pl-3.5 h-7 flex items-center"
    class:!pl-10={!$navigationOpen}
  >
    <WorkspaceBreadcrumbs {resource} {filePath} />
  </div>

  <div class="second-level-wrapper">
    <div class="flex gap-x-1 items-center w-full" class:truncate={!editing}>
      <span class="flex-none">
        <svelte:component
          this={resourceKind
            ? resourceIconMapping[resourceKind]
            : filePath === "/.env" || filePath === "/rill.yaml"
              ? Settings
              : File}
          size="19px"
          color={resourceKind ? resourceColorMapping[resourceKind] : "#9CA3AF"}
        />
      </span>

      <InputWithConfirm
        bind:editing
        size="md"
        {editable}
        id="model-title-input"
        textClass="text-xl font-semibold"
        {value}
        onConfirm={onTitleChange}
        showIndicator={hasUnsavedChanges}
      />
    </div>

    <div class="flex items-center gap-x-2 w-fit flex-none">
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
  </div>
</header>

<style lang="postcss">
  .second-level-wrapper {
    @apply px-4 py-2 w-full h-7;
    @apply flex justify-between gap-x-2;
    @apply items-center;
  }

  header {
    @apply flex flex-col py-2 gap-y-2;
  }
</style>
