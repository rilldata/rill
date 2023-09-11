<script lang="ts">
  import { goto } from "$app/navigation";
  import {
    Button,
    IconSpaceFixer,
  } from "@rilldata/web-common/components/button";
  import RefreshIcon from "@rilldata/web-common/components/icons/RefreshIcon.svelte";
  import PanelCTA from "@rilldata/web-common/components/panel/PanelCTA.svelte";
  import ResponsiveButtonText from "@rilldata/web-common/components/panel/ResponsiveButtonText.svelte";
  import { saveFile } from "@rilldata/web-common/features/entity-management/file-actions";
  import { validateAndRenameEntity } from "@rilldata/web-common/features/entity-management/rename-entity";
  import { useSource } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { overlay } from "@rilldata/web-common/layout/overlay-store";
  import { slideRight } from "@rilldata/web-common/lib/transitions";
  import {
    createRuntimeServiceGetFile,
    createRuntimeServicePutFile,
    createRuntimeServiceRefreshAndReconcile,
    createRuntimeServiceRenameFile,
    getRuntimeServiceGetCatalogEntryQueryKey,
    V1SourceV2,
  } from "@rilldata/web-common/runtime-client";
  import { appQueryStatusStore } from "@rilldata/web-common/runtime-client/application-store";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { fade } from "svelte/transition";
  import EnterIcon from "../../../components/icons/EnterIcon.svelte";
  import UndoIcon from "../../../components/icons/UndoIcon.svelte";
  import { WorkspaceHeader } from "../../../layout/workspace";
  import { behaviourEvent } from "../../../metrics/initMetrics";
  import { BehaviourEventMedium } from "../../../metrics/service/BehaviourEventTypes";
  import {
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "../../../metrics/service/MetricsTypes";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { getFilePathFromNameAndType } from "../../entity-management/entity-mappers";
  import {
    fileArtifactsStore,
    getFileArtifactReconciliationErrors,
  } from "../../entity-management/file-artifacts-store";
  import { useAllNames } from "../../entity-management/selectors";
  import { createModelFromSourceV2 } from "../createModel";
  import { refreshSource } from "../refreshSource";
  import { useIsSourceUnsaved } from "../selectors";
  import { useSourceStore } from "../sources-store";

  export let sourceName: string;

  const queryClient = useQueryClient();

  const refreshSourceMutation = createRuntimeServiceRefreshAndReconcile();
  const renameSource = createRuntimeServiceRenameFile();
  const saveSource = createRuntimeServicePutFile();

  $: runtimeInstanceId = $runtime.instanceId;

  $: sourceQuery = useSource(runtimeInstanceId, sourceName);
  let source: V1SourceV2;
  $: source = $sourceQuery.data;

  $: file = createRuntimeServiceGetFile(
    runtimeInstanceId,
    getFilePathFromNameAndType(sourceName, EntityType.Table)
  );

  let connector: string;
  $: connector = source?.spec?.sourceConnector;

  $: allNamesQuery = useAllNames(runtimeInstanceId);

  const onChangeCallback = async (e) => {
    if (
      !(await validateAndRenameEntity(
        runtimeInstanceId,
        e.target.value,
        sourceName,
        $allNamesQuery.data,
        EntityType.Table,
        renameSource
      ))
    ) {
      e.target.value = sourceName; // resets the input
    }
  };

  function onRevertChanges() {
    sourceStore.set({ clientYAML: $file.data?.blob || "" });
  }

  const onSaveAndRefreshClick = async (tableName: string) => {
    overlay.set({ title: `Importing ${tableName}.yaml` });
    await saveFile(
      runtimeInstanceId,
      tableName,
      EntityType.Table,
      $sourceStore.clientYAML,
      saveSource
    );
    // TODO: emit telemetry
    //       should it emit only when a source is modified from UI?
    //       or perhaps change screen and others based on where it is emitted from and always fire?
    overlay.set(null);
  };

  const onRefreshClick = async (tableName: string) => {
    try {
      // TODO: what is the replacement for refresh in new reconcile?
      await refreshSource(
        connector,
        tableName,
        runtimeInstanceId,
        $refreshSourceMutation,
        $saveSource,
        queryClient,
        connector === "s3" || connector === "gcs" || connector === "https"
          ? source?.spec?.properties?.path
          : sourceName
      );
      // invalidate the "refreshed_on" time
      const queryKey = getRuntimeServiceGetCatalogEntryQueryKey(
        runtimeInstanceId,
        tableName
      );
      await queryClient.refetchQueries(queryKey);
    } catch (err) {
      // no-op
    }
    overlay.set(null);
  };

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
    $sourceStore.clientYAML
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
      MetricsEventScreenName.Model
    );
  };

  let hasReconciliationErrors: boolean;
  $: {
    const reconciliationErrors = getFileArtifactReconciliationErrors(
      $fileArtifactsStore,
      `${sourceName}.yaml`
    );
    hasReconciliationErrors = reconciliationErrors?.length > 0;
  }

  function isHeaderWidthSmall(width: number) {
    return width < 800;
  }
</script>

<div class="grid items-center" style:grid-template-columns="auto max-content">
  <WorkspaceHeader
    {...{ titleInput: sourceName, onChangeCallback }}
    appRunning={$appQueryStatusStore}
    let:width={headerWidth}
  >
    <svelte:fragment slot="workspace-controls">
      {#if $refreshSourceMutation.isLoading}
        Refreshing...
      {:else}
        <div class="flex items-center pr-2 gap-x-2">
          {#if $sourceQuery.isSuccess && $sourceQuery.data?.state?.refreshedOn}
            <div
              class="ui-copy-muted"
              style:font-size="11px"
              transition:fade|local={{ duration: 200 }}
            >
              Imported on {formatRefreshedOn(
                $sourceQuery.data?.state?.refreshedOn
              )}
            </div>
          {/if}
        </div>
      {/if}
    </svelte:fragment>
    <svelte:fragment slot="cta">
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
        <Button
          label={isSourceUnsaved ? "Save and refresh" : "Refresh"}
          on:click={() =>
            isSourceUnsaved
              ? onSaveAndRefreshClick(sourceName)
              : onRefreshClick(sourceName)}
          type={isSourceUnsaved ? "primary" : "secondary"}
        >
          <IconSpaceFixer pullLeft pullRight={isHeaderWidthSmall(headerWidth)}>
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
        </Button>
        <Button
          disabled={isSourceUnsaved || hasReconciliationErrors}
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
