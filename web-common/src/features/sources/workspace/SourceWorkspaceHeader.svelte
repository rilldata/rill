<script lang="ts">
  import {
    Button,
    IconSpaceFixer,
  } from "@rilldata/web-common/components/button";
  import RefreshIcon from "@rilldata/web-common/components/icons/RefreshIcon.svelte";
  import PanelCTA from "@rilldata/web-common/components/panel/PanelCTA.svelte";
  import ResponsiveButtonText from "@rilldata/web-common/components/panel/ResponsiveButtonText.svelte";
  import { createFileSaver } from "@rilldata/web-common/features/entity-management/file-actions";
  import { createEntityRefresher } from "@rilldata/web-common/features/entity-management/refresh-entity";
  import { createFileValidatorAndRenamer } from "@rilldata/web-common/features/entity-management/rename-entity";
  import {
    useAllEntityNames,
    useSource,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { overlay } from "@rilldata/web-common/layout/overlay-store";
  import { slideRight } from "@rilldata/web-common/lib/transitions";
  import { getRightPanelParams } from "@rilldata/web-common/metrics/service/metrics-helpers";
  import {
    createRuntimeServiceGetFile,
    createRuntimeServiceRefreshAndReconcile,
    V1SourceV2,
  } from "@rilldata/web-common/runtime-client";
  import { appQueryStatusStore } from "@rilldata/web-common/runtime-client/application-store";
  import { fade } from "svelte/transition";
  import { createModelFromSourceCreator } from "web-common/src/features/sources/createModelFromSource";
  import EnterIcon from "../../../components/icons/EnterIcon.svelte";
  import UndoIcon from "../../../components/icons/UndoIcon.svelte";
  import { WorkspaceHeader } from "../../../layout/workspace";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { getFilePathFromNameAndType } from "../../entity-management/entity-mappers";
  import {
    fileArtifactsStore,
    getFileArtifactReconciliationErrors,
  } from "../../entity-management/file-artifacts-store";
  import { useIsSourceUnsaved } from "../selectors";
  import { useSourceStore } from "../sources-store";

  export let sourceName: string;

  const refreshSourceMutation = createRuntimeServiceRefreshAndReconcile();
  const fileSaver = createFileSaver();
  const sourceRefresher = createEntityRefresher();

  $: allNamesQuery = useAllEntityNames(runtimeInstanceId);

  $: fileValidatorAndRenamer = createFileValidatorAndRenamer(allNamesQuery);
  $: modelFromSourceCreator = createModelFromSourceCreator(
    allNamesQuery,
    getRightPanelParams()
  );

  $: runtimeInstanceId = $runtime.instanceId;

  $: sourceQuery = useSource(runtimeInstanceId, sourceName);
  let source: V1SourceV2;
  $: source = $sourceQuery.data?.source;

  $: file = createRuntimeServiceGetFile(
    runtimeInstanceId,
    getFilePathFromNameAndType(sourceName, EntityType.Table)
  );

  const onChangeCallback = async (e) => {
    if (
      !(await fileValidatorAndRenamer(
        sourceName,
        e.target.value,
        EntityType.Table
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
    await fileSaver(
      getFilePathFromNameAndType(tableName, EntityType.Table),
      $sourceStore.clientYAML
    );
    // TODO: emit telemetry
    //       should it emit only when a source is modified from UI?
    //       or perhaps change screen and others based on where it is emitted from and always fire?
    overlay.set(null);
  };

  const onRefreshClick = async () => {
    try {
      await sourceRefresher($sourceQuery.data);
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
    await modelFromSourceCreator(
      undefined, // TODO
      sourceName,
      "/models/"
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
          {#if $sourceQuery.isSuccess && $sourceQuery.data?.meta?.reconcileOn}
            <div
              class="ui-copy-muted"
              style:font-size="11px"
              transition:fade|local={{ duration: 200 }}
            >
              Imported on {formatRefreshedOn(
                $sourceQuery.data?.meta?.reconcileOn
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
              : onRefreshClick()}
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
