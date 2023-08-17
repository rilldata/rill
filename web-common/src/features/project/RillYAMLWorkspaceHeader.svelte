<script lang="ts">
  import {
    createRuntimeServiceGetFile,
    createRuntimeServicePutFileAndReconcile,
    createRuntimeServiceRefreshAndReconcile,
  } from "@rilldata/web-common/runtime-client";
  import { appQueryStatusStore } from "@rilldata/web-common/runtime-client/application-store";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { WorkspaceHeader } from "../../layout/workspace";
  import { runtime } from "../../runtime-client/runtime-store";

  const queryClient = useQueryClient();

  const refreshSourceMutation = createRuntimeServiceRefreshAndReconcile();
  const createSource = createRuntimeServicePutFileAndReconcile();

  $: runtimeInstanceId = $runtime.instanceId;
  $: file = createRuntimeServiceGetFile(runtimeInstanceId, "rill.yaml");

  // function onRevertChanges() {
  //   sourceStore.set({ clientYAML: $file.data?.blob || "" });
  // }

  // const onSaveAndRefreshClick = async (tableName: string) => {
  //   overlay.set({ title: `Importing ${tableName}.yaml` });
  //   await saveAndRefresh(queryClient, tableName, $sourceStore.clientYAML);
  //   overlay.set(null);
  // };

  // const onRefreshClick = async (tableName: string) => {
  //   try {
  //     await refreshSource(
  //       connector,
  //       tableName,
  //       runtimeInstanceId,
  //       $refreshSourceMutation,
  //       $createSource,
  //       queryClient,
  //       source?.connector === "s3" ||
  //         source?.connector === "gcs" ||
  //         source?.connector === "https"
  //         ? source?.properties?.path
  //         : sourceName
  //     );
  //     // invalidate the "refreshed_on" time
  //     const queryKey = getRuntimeServiceGetCatalogEntryQueryKey(
  //       runtimeInstanceId,
  //       tableName
  //     );
  //     await queryClient.refetchQueries(queryKey);
  //   } catch (err) {
  //     // no-op
  //   }
  //   overlay.set(null);
  // };

  // const sourceStore = useSourceStore(sourceName);

  // $: isSourceUnsavedQuery = useIsSourceUnsaved(
  //   $runtime.instanceId,
  //   sourceName,
  //   $sourceStore.clientYAML
  // );
  // $: isSourceUnsaved = $isSourceUnsavedQuery.data;

  function isHeaderWidthSmall(width: number) {
    return width < 800;
  }
</script>

<div class="grid items-center" style:grid-template-columns="auto max-content">
  <WorkspaceHeader
    titleInput="rill.yaml"
    onChangeCallback={undefined}
    appRunning={$appQueryStatusStore}
    editable={false}
    showInspectorToggle={false}
    let:width={headerWidth}
  >
    <svelte:fragment slot="cta">
      <!-- <PanelCTA side="right">
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
                  transition:slideRight={{ duration: 250 }}
                >
                  Save and
                </div>{/if}
              {#if !isSourceUnsaved}R{:else}r{/if}efresh
            </div>
          </ResponsiveButtonText>
        </Button>
      </PanelCTA> -->
    </svelte:fragment>
  </WorkspaceHeader>
</div>
