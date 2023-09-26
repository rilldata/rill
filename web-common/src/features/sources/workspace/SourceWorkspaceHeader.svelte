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
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { overlay } from "@rilldata/web-common/layout/overlay-store";
  import { slideRight } from "@rilldata/web-common/lib/transitions";
  import {
    createRuntimeServiceGetCatalogEntry,
    createRuntimeServiceGetFile,
    createRuntimeServicePutFileAndReconcile,
    createRuntimeServiceRefreshAndReconcile,
    createRuntimeServiceRenameFileAndReconcile,
    getRuntimeServiceGetCatalogEntryQueryKey,
    V1CatalogEntry,
    V1Source,
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
  import { renameFileArtifact } from "../../entity-management/actions";
  import {
    getFilePathFromNameAndType,
    getRouteFromName,
  } from "../../entity-management/entity-mappers";
  import {
    fileArtifactsStore,
    getFileArtifactReconciliationErrors,
  } from "../../entity-management/file-artifacts-store";
  import { isDuplicateName } from "../../entity-management/name-utils";
  import { useAllNames } from "../../entity-management/selectors";
  import { createModelFromSourceV2 } from "../createModel";
  import { refreshSource } from "../refreshSource";
  import { saveAndRefresh } from "../saveAndRefresh";
  import { useIsSourceUnsaved } from "../selectors";
  import { useSourceStore } from "../sources-store";

  export let sourceName: string;

  const queryClient = useQueryClient();

  const renameSource = createRuntimeServiceRenameFileAndReconcile();
  const refreshSourceMutation = createRuntimeServiceRefreshAndReconcile();
  const createSource = createRuntimeServicePutFileAndReconcile();

  $: runtimeInstanceId = $runtime.instanceId;
  $: getSource = createRuntimeServiceGetCatalogEntry(
    runtimeInstanceId,
    sourceName
  );
  $: file = createRuntimeServiceGetFile(
    runtimeInstanceId,
    getFilePathFromNameAndType(sourceName, EntityType.Table)
  );

  let entry: V1CatalogEntry;
  let source: V1Source;
  $: entry = $getSource?.data?.entry;
  $: source = entry?.source;

  let connector: string;
  $: connector = $getSource.data?.entry?.source?.connector as string;

  $: allNamesQuery = useAllNames(runtimeInstanceId);

  const onChangeCallback = async (e) => {
    if (!e.target.value.match(/^[a-zA-Z_][a-zA-Z0-9_]*$/)) {
      notifications.send({
        message:
          "Source name must start with a letter or underscore and contain only letters, numbers, and underscores",
      });
      e.target.value = sourceName; // resets the input
      return;
    }
    if (isDuplicateName(e.target.value, sourceName, $allNamesQuery.data)) {
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
        queryClient,
        runtimeInstanceId,
        sourceName,
        toName,
        entityType,
        $renameSource
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
    await saveAndRefresh(queryClient, tableName, $sourceStore.clientYAML);
    overlay.set(null);
  };

  const onRefreshClick = async (tableName: string) => {
    try {
      await refreshSource(
        connector,
        tableName,
        runtimeInstanceId,
        $refreshSourceMutation,
        $createSource,
        queryClient,
        source?.connector === "s3" ||
          source?.connector === "gcs" ||
          source?.connector === "https"
          ? source?.properties?.path
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
          {#if $getSource.isSuccess && $getSource.data?.entry?.refreshedOn}
            <div
              class="ui-copy-muted"
              style:font-size="11px"
              transition:fade={{ duration: 200 }}
            >
              Imported on {formatRefreshedOn(
                $getSource.data?.entry?.refreshedOn
              )}
            </div>
          {/if}
        </div>
      {/if}
    </svelte:fragment>
    <svelte:fragment slot="cta">
      <PanelCTA side="right">
        <Button
          on:click={() => onRevertChanges()}
          type="secondary"
          disabled={!isSourceUnsaved}
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
                  transition:slideRight|global={{ duration: 250 }}
                >
                  Save and
                </div>{/if}
              {#if !isSourceUnsaved}R{:else}r{/if}efresh
            </div>
          </ResponsiveButtonText>
        </Button>
        <Button
          on:click={handleCreateModelFromSource}
          disabled={isSourceUnsaved || hasReconciliationErrors}
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
