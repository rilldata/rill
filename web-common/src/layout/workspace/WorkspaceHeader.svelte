<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import InputWithConfirm from "@rilldata/web-common/components/forms/InputWithConfirm.svelte";
  import File from "@rilldata/web-common/components/icons/File.svelte";
  import HideBottomPane from "@rilldata/web-common/components/icons/HideBottomPane.svelte";
  import HideSidebar from "@rilldata/web-common/components/icons/HideSidebar.svelte";
  import RefreshIcon from "@rilldata/web-common/components/icons/RefreshIcon.svelte";
  import SlidingWords from "@rilldata/web-common/components/tooltip/SlidingWords.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import {
    resourceColorMapping,
    resourceIconMapping,
  } from "@rilldata/web-common/features/entity-management/resource-icon-mapping";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import CodeToggle from "@rilldata/web-common/features/visual-editing/CodeToggle.svelte";
  import WorkspaceBreadcrumbs from "@rilldata/web-common/features/workspaces/WorkspaceBreadcrumbs.svelte";
  import {
    V1ReconcileStatus,
    type V1Resource,
    createRuntimeServiceCreateTrigger,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { Settings } from "lucide-svelte";
  import { navigationOpen } from "../navigation/Navigation.svelte";
  import { workspaces } from "./workspace-stores";

  export let resourceKind: ResourceKind | undefined;
  export let titleInput: string;
  export let editable = true;
  export let showInspectorToggle = true;
  export let showTableToggle = false;
  export let hasUnsavedChanges: boolean;
  export let filePath: string;
  export let codeToggle = false;
  export let resource: V1Resource | undefined = undefined;
  export let onTitleChange: (title: string) => void = () => {};

  let width: number;
  let editing = false;

  $: value = titleInput;
  $: workspaceLayout = workspaces.get(filePath);
  $: inspectorVisible = workspaceLayout.inspector.visible;
  $: tableVisible = workspaceLayout.table.visible;
  $: view = workspaceLayout.view;

  $: ({ instanceId } = $runtime);
  $: isConnector = resourceKind === ResourceKind.Connector;
  $: connectorName = resource?.meta?.name?.name;
  $: isReconciling =
    resource?.meta?.reconcileStatus ===
    V1ReconcileStatus.RECONCILE_STATUS_RUNNING;

  const triggerMutation = createRuntimeServiceCreateTrigger();

  function refreshConnector() {
    if (!isConnector || !connectorName) return;
    $triggerMutation.mutate({
      instanceId,
      data: {
        resources: [{ kind: ResourceKind.Connector, name: connectorName }],
      },
    });
  }
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
          <svelte:component
            this={resourceKind
              ? resourceIconMapping[resourceKind]
              : filePath === "/.env" || filePath === "/rill.yaml"
                ? Settings
                : File}
            size="19px"
            color={resourceKind
              ? resourceColorMapping[resourceKind]
              : "#9CA3AF"}
          />
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

    {#if isConnector}
      <div class="flex items-center gap-x-2">
        <Button
          type="secondary"
          onClick={refreshConnector}
          disabled={$triggerMutation.isPending || isReconciling}
          loading={$triggerMutation.isPending}
          loadingCopy="Refreshing"
        >
          <RefreshIcon size="14px" />
          {isReconciling ? "Refreshing…" : "Refresh"}
        </Button>
      </div>
    {/if}

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
