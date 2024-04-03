<script lang="ts">
  import { goto } from "$app/navigation";
  import {
    Button,
    IconSpaceFixer,
  } from "@rilldata/web-common/components/button";
  import RefreshIcon from "@rilldata/web-common/components/icons/RefreshIcon.svelte";
  import PanelCTA from "@rilldata/web-common/components/panel/PanelCTA.svelte";
  import ResponsiveButtonText from "@rilldata/web-common/components/panel/ResponsiveButtonText.svelte";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import { resourceIsLoading } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { handleEntityRename } from "@rilldata/web-common/features/entity-management/ui-actions";
  import { slideRight } from "@rilldata/web-common/lib/transitions";
  import { createRuntimeServiceRefreshAndReconcile } from "@rilldata/web-common/runtime-client";
  import { fade } from "svelte/transition";
  import { WithTogglableFloatingElement } from "../../../components/floating-element";
  import CaretDownIcon from "../../../components/icons/CaretDownIcon.svelte";
  import EnterIcon from "../../../components/icons/EnterIcon.svelte";
  import UndoIcon from "../../../components/icons/UndoIcon.svelte";
  import { Menu, MenuItem } from "../../../components/menu";
  import { WorkspaceHeader } from "../../../layout/workspace";
  import { behaviourEvent } from "../../../metrics/initMetrics";
  import { BehaviourEventMedium } from "../../../metrics/service/BehaviourEventTypes";
  import {
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "../../../metrics/service/MetricsTypes";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { createModelFromSourceV2 } from "../createModel";
  import {
    refreshSource,
    replaceSourceWithUploadedFile,
  } from "../refreshSource";
  import { useIsLocalFileConnector } from "../selectors";
  import { createEventDispatcher } from "svelte";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";

  const refreshSourceMutation = createRuntimeServiceRefreshAndReconcile();

  const dispatch = createEventDispatcher();

  export let filePath: string;
  export let isSourceUnsaved: boolean;
  export let sourceName: string;

  $: fileArtifact = fileArtifacts.getFileArtifact(filePath);

  $: runtimeInstanceId = $runtime.instanceId;
  $: sourceQuery = fileArtifact.getResource(queryClient, runtimeInstanceId);

  $: source = $sourceQuery.data?.source;
  $: sourceIsReconciling = resourceIsLoading($sourceQuery.data);

  $: connector = source?.state?.connector;

  $: hasErrors = fileArtifact.getHasErrors(queryClient, $runtime.instanceId);

  $: isLocalFileConnectorQuery = useIsLocalFileConnector(
    $runtime.instanceId,
    filePath,
  );
  $: isLocalFileConnector = $isLocalFileConnectorQuery.data;

  function onChangeCallback(
    e: Event & {
      currentTarget: EventTarget & HTMLInputElement;
    },
  ) {
    return handleEntityRename(
      queryClient,
      runtimeInstanceId,
      e,
      filePath,
      EntityType.Table,
    );
  }

  async function onReplaceSource() {
    await replaceSourceWithUploadedFile(runtimeInstanceId, filePath);
  }

  function formatRefreshedOn(refreshedOn: string) {
    const date = new Date(refreshedOn);
    return date.toLocaleString(undefined, {
      month: "short",
      day: "numeric",
      year: "numeric",
      hour: "numeric",
      minute: "numeric",
    });
  }

  const handleCreateModelFromSource = async () => {
    const modelName = await createModelFromSourceV2(
      queryClient,
      source?.state?.table ?? "",
    );
    await goto(`/model/${modelName}`);
    await behaviourEvent.fireNavigationEvent(
      modelName,
      BehaviourEventMedium.Button,
      MetricsEventSpace.RightPanel,
      MetricsEventScreenName.Source,
      MetricsEventScreenName.Model,
    );
  };

  function onRevertChanges() {
    dispatch("revert");
  }

  function onSaveAndRefreshClick() {
    dispatch("save");
  }

  const onRefreshClick = async () => {
    // no-op if connector is undefined
    if (connector === undefined) return;

    try {
      await refreshSource(
        connector,
        filePath,
        $sourceQuery.data?.meta?.name?.name ?? "",
        runtimeInstanceId,
      );
    } catch (err) {
      // no-op
    }
  };

  function isHeaderWidthSmall(width: number) {
    return width < 800;
  }
</script>

<div class="grid items-center" style:grid-template-columns="auto max-content">
  <WorkspaceHeader
    {...{ titleInput: sourceName, onChangeCallback }}
    {isSourceUnsaved}
  >
    <svelte:fragment slot="workspace-controls">
      {#if $refreshSourceMutation.isLoading}
        Refreshing...
      {:else}
        <div class="flex items-center pr-2 gap-x-2">
          {#if $sourceQuery && source?.state?.refreshedOn}
            <div
              class="ml-2 ui-copy-muted line-clamp-2"
              style:font-size="11px"
              transition:fade={{ duration: 200 }}
            >
              Ingested on {formatRefreshedOn(source?.state?.refreshedOn)}
            </div>
          {/if}
        </div>
      {/if}
    </svelte:fragment>
    <svelte:fragment let:width={headerWidth} slot="cta">
      <PanelCTA side="right">
        <Button
          disabled={!isSourceUnsaved}
          on:click={() => onRevertChanges()}
          type="secondary"
        >
          <IconSpaceFixer pullLeft pullRight={isHeaderWidthSmall(headerWidth)}>
            <UndoIcon size="14px" />
          </IconSpaceFixer>
          <ResponsiveButtonText collapse={isHeaderWidthSmall(headerWidth)}>
            Revert changes
          </ResponsiveButtonText>
        </Button>
        <WithTogglableFloatingElement
          alignment="end"
          distance={8}
          let:toggleFloatingElement
          location="bottom"
        >
          <Button
            disabled={sourceIsReconciling}
            label={isSourceUnsaved ? "Save and refresh" : "Refresh"}
            on:click={() =>
              isSourceUnsaved
                ? onSaveAndRefreshClick()
                : isLocalFileConnector
                  ? toggleFloatingElement()
                  : onRefreshClick()}
            type={isSourceUnsaved ? "primary" : "secondary"}
          >
            <IconSpaceFixer
              pullLeft
              pullRight={isHeaderWidthSmall(headerWidth)}
            >
              <RefreshIcon size="14px" />
            </IconSpaceFixer>
            <ResponsiveButtonText collapse={isHeaderWidthSmall(headerWidth)}>
              <div class="flex">
                {#if isSourceUnsaved}<div
                    class="pr-1"
                    transition:slideRight={{ duration: 250 }}
                  >
                    Save and
                  </div>{/if}
                {#if !isSourceUnsaved}R{:else}r{/if}efresh
              </div>
            </ResponsiveButtonText>
            {#if !isSourceUnsaved && isLocalFileConnector}
              <CaretDownIcon size="14px" />
            {/if}
          </Button>
          <Menu
            dark
            let:toggleFloatingElement
            on:click-outside={toggleFloatingElement}
            on:escape={toggleFloatingElement}
            slot="floating-element"
          >
            <MenuItem
              on:select={() => {
                toggleFloatingElement();
                onRefreshClick();
              }}
            >
              Refresh source
            </MenuItem>
            <MenuItem
              on:select={() => {
                toggleFloatingElement();
                onReplaceSource();
              }}
            >
              Replace source with uploaded file
            </MenuItem>
          </Menu>
        </WithTogglableFloatingElement>
        <Button
          disabled={isSourceUnsaved || $hasErrors}
          on:click={handleCreateModelFromSource}
        >
          <ResponsiveButtonText collapse={isHeaderWidthSmall(headerWidth)}>
            Create model
          </ResponsiveButtonText>
          <IconSpaceFixer pullLeft pullRight={isHeaderWidthSmall(headerWidth)}>
            <EnterIcon size="14px" />
          </IconSpaceFixer>
        </Button>
      </PanelCTA>
    </svelte:fragment>
  </WorkspaceHeader>
</div>
