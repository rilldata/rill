<script lang="ts">
  import { goto } from "$app/navigation";
  import {
    Button,
    IconSpaceFixer,
  } from "@rilldata/web-common/components/button";
  import RefreshIcon from "@rilldata/web-common/components/icons/RefreshIcon.svelte";
  import { notifications } from "@rilldata/web-common/components/notifications";
  import PanelCTA from "@rilldata/web-common/components/panel/PanelCTA.svelte";
  import ResponsiveButtonText from "@rilldata/web-common/components/panel/ResponsiveButtonText.svelte";
  import { fileArtifactsStore } from "@rilldata/web-common/features/entity-management/file-artifacts-store";
  import {
    resourceIsLoading,
    useAllNames,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { checkSourceImported } from "@rilldata/web-common/features/sources/source-imported-utils";
  import { overlay } from "@rilldata/web-common/layout/overlay-store";
  import { slideRight } from "@rilldata/web-common/lib/transitions";
  import {
    V1SourceV2,
    createRuntimeServiceGetFile,
    createRuntimeServiceRefreshAndReconcile,
  } from "@rilldata/web-common/runtime-client";
  import { useQueryClient } from "@tanstack/svelte-query";
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
  import { renameFileArtifact } from "../../entity-management/actions";
  import {
    getFilePathFromNameAndType,
    getRouteFromName,
  } from "../../entity-management/entity-mappers";
  import {
    INVALID_NAME_MESSAGE,
    VALID_NAME_PATTERN,
    isDuplicateName,
  } from "../../entity-management/name-utils";
  import { createModelFromSourceV2 } from "../createModel";
  import {
    refreshSource,
    replaceSourceWithUploadedFile,
  } from "../refreshSource";
  import { saveAndRefresh } from "../saveAndRefresh";
  import {
    useIsLocalFileConnector,
    useIsSourceUnsaved,
    useSource,
  } from "../selectors";
  import { useSourceStore } from "../sources-store";

  export let sourceName: string;
  $: filePath = getFilePathFromNameAndType(sourceName, EntityType.Table);

  const queryClient = useQueryClient();

  const refreshSourceMutation = createRuntimeServiceRefreshAndReconcile();

  $: runtimeInstanceId = $runtime.instanceId;
  $: sourceQuery = useSource(runtimeInstanceId, sourceName);
  $: file = createRuntimeServiceGetFile(runtimeInstanceId, filePath);

  let source: V1SourceV2 | undefined;
  $: source = $sourceQuery.data?.source;
  $: sourceIsReconciling = resourceIsLoading($sourceQuery.data);

  let connector: string | undefined;
  $: connector = source?.state?.connector;

  $: allNamesQuery = useAllNames(runtimeInstanceId);

  const onChangeCallback = async (e) => {
    if (!e.target.value.match(VALID_NAME_PATTERN)) {
      notifications.send({
        message: INVALID_NAME_MESSAGE,
      });
      e.target.value = sourceName; // resets the input
      return;
    }
    if (
      isDuplicateName(e.target.value, sourceName, $allNamesQuery?.data ?? [])
    ) {
      notifications.send({
        message: `Name ${e.target.value} is already in use`,
      });
      e.target.value = sourceName; // resets the input
      return;
    }

    try {
      const toName = e.target.value;
      const entityType = EntityType.Table;
      await renameFileArtifact(
        runtimeInstanceId,
        sourceName,
        toName,
        entityType,
      );
      goto(getRouteFromName(toName, entityType), {
        replaceState: true,
      });
    } catch (err) {
      console.error(err.response.data.message);
    }
  };

  function onRevertChanges() {
    sourceStore.set({ clientYAML: $file.data?.blob || "" });
  }

  const onSaveAndRefreshClick = async (tableName: string) => {
    overlay.set({ title: `Importing ${tableName}.yaml` });
    await saveAndRefresh(tableName, $sourceStore.clientYAML);
    checkSourceImported(queryClient, sourceName, filePath);
    overlay.set(null);
  };

  const onRefreshClick = async (tableName: string) => {
    // no-op if connector is undefined
    if (connector === undefined) return;

    try {
      await refreshSource(connector, tableName, runtimeInstanceId);
    } catch (err) {
      // no-op
    }
  };

  $: isLocalFileConnectorQuery = useIsLocalFileConnector(
    $runtime.instanceId,
    sourceName,
  );
  $: isLocalFileConnector = $isLocalFileConnectorQuery.data;

  async function onReplaceSource(sourceName: string) {
    await replaceSourceWithUploadedFile(runtimeInstanceId, sourceName);
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

  const sourceStore = useSourceStore(sourceName);

  $: isSourceUnsavedQuery = useIsSourceUnsaved(
    $runtime.instanceId,
    sourceName,
    $sourceStore.clientYAML,
  );
  $: isSourceUnsaved = $isSourceUnsavedQuery.data;

  const handleCreateModelFromSource = async () => {
    const modelName = await createModelFromSourceV2(queryClient, sourceName);
    goto(`/model/${modelName}`);
    behaviourEvent.fireNavigationEvent(
      modelName,
      BehaviourEventMedium.Button,
      MetricsEventSpace.RightPanel,
      MetricsEventScreenName.Source,
      MetricsEventScreenName.Model,
    );
  };

  $: hasErrors = fileArtifactsStore.getFileHasErrors(
    queryClient,
    $runtime.instanceId,
    filePath,
  );

  function isHeaderWidthSmall(width: number) {
    return width < 800;
  }
</script>

<div class="grid items-center" style:grid-template-columns="auto max-content">
  <WorkspaceHeader {...{ titleInput: sourceName, onChangeCallback }}>
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
                ? onSaveAndRefreshClick(sourceName)
                : isLocalFileConnector
                  ? toggleFloatingElement()
                  : onRefreshClick(sourceName)}
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
                    transition:slideRight|global={{ duration: 250 }}
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
                onRefreshClick(sourceName);
              }}
            >
              Refresh source
            </MenuItem>
            <MenuItem
              on:select={() => {
                toggleFloatingElement();
                onReplaceSource(sourceName);
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
