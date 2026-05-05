<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import InputWithConfirm from "@rilldata/web-common/components/forms/InputWithConfirm.svelte";
  import HideBottomPane from "@rilldata/web-common/components/icons/HideBottomPane.svelte";
  import HideSidebar from "@rilldata/web-common/components/icons/HideSidebar.svelte";
  import SlidingWords from "@rilldata/web-common/components/tooltip/SlidingWords.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { getIconComponent } from "@rilldata/web-common/features/entity-management/resource-icon-mapping";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import CodeToggle from "@rilldata/web-common/features/visual-editing/CodeToggle.svelte";
  import WorkspaceBreadcrumbs from "@rilldata/web-common/features/workspaces/WorkspaceBreadcrumbs.svelte";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import { navigationOpen } from "../navigation/Navigation.svelte";
  import { workspaces } from "./workspace-stores";
  import ConnectorRefreshButton from "@rilldata/web-common/features/connectors/ConnectorRefreshButton.svelte";
  import ConnectorAddModelButton from "@rilldata/web-common/features/connectors/ConnectorAddModelButton.svelte";
  import type { Snippet } from "svelte";

  let {
    resourceKind,
    resource = undefined,
    titleInput,
    editable = true,
    nonEditableMessage,
    showInspectorToggle = true,
    showTableToggle = false,
    hasUnsavedChanges,
    filePath,
    codeToggle = false,
    onTitleChange,
    workspaceControls,
    cta,
  }: {
    resourceKind: ResourceKind | undefined;
    resource?: V1Resource | undefined;
    titleInput: string;
    editable?: boolean;
    nonEditableMessage?: Snippet<[]>;
    showInspectorToggle?: boolean;
    showTableToggle?: boolean;
    hasUnsavedChanges: boolean;
    filePath: string;
    codeToggle?: boolean;
    onTitleChange?: (title: string) => void;
    workspaceControls?: Snippet<[number]>;
    cta?: Snippet<[number]>;
  } = $props();

  let width: number = $state(0);
  let editing = $state(false);

  let value = $derived(titleInput);
  let workspaceLayout = $derived(workspaces.get(filePath));
  let inspectorVisible = $derived(workspaceLayout.inspector.visible);
  let tableVisible = $derived(workspaceLayout.table.visible);
  let view = $derived(workspaceLayout.view);

  // Check if it's a connector by resourceKind or by file path.
  // File path fallback is needed when reconcile fails and resourceKind is unavailable.
  let isConnector = $derived(
    resourceKind === ResourceKind.Connector ||
      (filePath && filePath.startsWith("/connectors/")),
  );

  let IconComponent = $derived(getIconComponent(resourceKind, filePath));
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
      {#if codeToggle && resourceKind}
        <CodeToggle bind:selectedView={$view} {resourceKind} />
      {:else}
        <span class="flex-none">
          <IconComponent size="19px" />
        </span>
      {/if}

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
      {#if !editable && nonEditableMessage}
        {@render nonEditableMessage()}
      {/if}

      {#if isConnector}
        <ConnectorRefreshButton {resource} {hasUnsavedChanges} />
        <ConnectorAddModelButton {resource} {hasUnsavedChanges} />
      {/if}

      {@render workspaceControls?.(width)}
      <div class="flex-none">
        {@render cta?.(width)}
      </div>

      {#if showTableToggle}
        <Tooltip distance={8}>
          <Button
            type="secondary"
            square
            selected={$tableVisible}
            label="Toggle table visibility"
            onClick={workspaceLayout.table.toggle}
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
            label="Toggle inspector visibility"
            onClick={workspaceLayout.inspector.toggle}
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
