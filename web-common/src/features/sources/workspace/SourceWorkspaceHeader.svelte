<script lang="ts">
  import { goto } from "$app/navigation";
  import { Button, IconButton } from "@rilldata/web-common/components/button";
  import Explore from "@rilldata/web-common/components/icons/Explore.svelte";
  import Import from "@rilldata/web-common/components/icons/Import.svelte";
  import Model from "@rilldata/web-common/components/icons/Model.svelte";
  import RefreshIcon from "@rilldata/web-common/components/icons/RefreshIcon.svelte";
  import Source from "@rilldata/web-common/components/icons/Source.svelte";
  import { notifications } from "@rilldata/web-common/components/notifications";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { EntityType } from "@rilldata/web-common/lib/entity";
  import {
    getRuntimeServiceGetCatalogEntryQueryKey,
    useRuntimeServiceGetCatalogEntry,
    useRuntimeServicePutFileAndReconcile,
    useRuntimeServiceRefreshAndReconcile,
    useRuntimeServiceRenameFileAndReconcile,
    V1ReconcileResponse,
    V1Source,
  } from "@rilldata/web-common/runtime-client";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { fileArtifactsStore } from "@rilldata/web-local/lib/application-state-stores/file-artifacts-store";
  import { overlay } from "@rilldata/web-local/lib/application-state-stores/overlay-store";
  import PanelCTA from "@rilldata/web-local/lib/components/panel/PanelCTA.svelte";
  import ResponsiveButtonText from "@rilldata/web-local/lib/components/panel/ResponsiveButtonText.svelte";
  import WorkspaceHeader from "@rilldata/web-local/lib/components/workspace/core/WorkspaceHeader.svelte";
  import { navigationEvent } from "@rilldata/web-local/lib/metrics/initMetrics";
  import { BehaviourEventMedium } from "@rilldata/web-local/lib/metrics/service/BehaviourEventTypes";
  import {
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-local/lib/metrics/service/MetricsTypes";
  import {
    isDuplicateName,
    renameFileArtifact,
    useAllNames,
    useCreateDashboardFromSource,
  } from "@rilldata/web-local/lib/svelte-query/actions";
  import { selectTimestampColumnFromSchema } from "@rilldata/web-local/lib/svelte-query/column-selectors";
  import { useDashboardNames } from "@rilldata/web-local/lib/svelte-query/dashboards";
  import { invalidateAfterReconcile } from "@rilldata/web-local/lib/svelte-query/invalidation";
  import { getRouteFromName } from "@rilldata/web-local/lib/util/entity-mappers";
  import { getName } from "@rilldata/web-local/lib/util/incrementName";
  import { useQueryClient } from "@sveltestack/svelte-query";
  import { fade } from "svelte/transition";
  import { useModelNames } from "../../models/selectors";
  import { createModelFromSource } from "../createModel";
  import { refreshSource } from "../refreshSource";

  export let sourceName: string;
  export let path: string;
  export let embedded = false;

  const queryClient = useQueryClient();

  const renameSource = useRuntimeServiceRenameFileAndReconcile();

  $: runtimeInstanceId = $runtimeStore.instanceId;
  const refreshSourceMutation = useRuntimeServiceRefreshAndReconcile();
  const createSource = useRuntimeServicePutFileAndReconcile();

  $: getSource = useRuntimeServiceGetCatalogEntry(
    runtimeInstanceId,
    sourceName
  );

  let source: V1Source;
  $: source = $getSource?.data?.entry?.source;

  $: modelNames = useModelNames(runtimeInstanceId);
  $: dashboardNames = useDashboardNames(runtimeInstanceId);
  const createModelMutation = useRuntimeServicePutFileAndReconcile();
  const createDashboardFromSourceMutation = useCreateDashboardFromSource();

  $: timestampColumns = selectTimestampColumnFromSchema(source?.schema);

  $: connector = $getSource.data?.entry?.source.connector as string;

  $: allNamesQuery = useAllNames(runtimeInstanceId);

  const handleCreateModelFromSource = async () => {
    const modelName = await createModelFromSource(
      queryClient,
      runtimeInstanceId,
      $modelNames.data,
      sourceName,
      embedded ? `"${path}"` : sourceName,
      $createModelMutation
    );
    navigationEvent.fireEvent(
      modelName,
      BehaviourEventMedium.Button,
      MetricsEventSpace.RightPanel,
      MetricsEventScreenName.Source,
      MetricsEventScreenName.Model
    );
  };

  const handleCreateDashboardFromSource = (sourceName: string) => {
    overlay.set({
      title: "Creating a dashboard for " + sourceName,
    });
    const newModelName = getName(`${sourceName}_model`, $modelNames.data);
    const newDashboardName = getName(
      `${sourceName}_dashboard`,
      $dashboardNames.data
    );
    $createDashboardFromSourceMutation.mutate(
      {
        data: {
          instanceId: $runtimeStore.instanceId,
          sourceName,
          newModelName,
          newDashboardName,
        },
      },
      {
        onSuccess: async (resp: V1ReconcileResponse) => {
          fileArtifactsStore.setErrors(resp.affectedPaths, resp.errors);
          goto(`/dashboard/${newDashboardName}`);
          navigationEvent.fireEvent(
            newDashboardName,
            BehaviourEventMedium.Button,
            MetricsEventSpace.RightPanel,
            MetricsEventScreenName.Source,
            MetricsEventScreenName.Dashboard
          );
          return invalidateAfterReconcile(queryClient, runtimeInstanceId, resp);
        },
        onSettled: () => {
          overlay.set(null);
        },
      }
    );
  };

  const onChangeCallback = async (e) => {
    if (!e.target.value.match(/^[a-zA-Z_][a-zA-Z0-9_]*$/)) {
      notifications.send({
        message:
          "Source name must start with a letter or underscore and contain only letters, numbers, and underscores",
      });
      e.target.value = sourceName; // resets the input
      return;
    }
    if (isDuplicateName(e.target.value, $allNamesQuery.data)) {
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

  const onRefreshClick = async (tableName: string) => {
    try {
      await refreshSource(
        connector,
        tableName,
        runtimeInstanceId,
        $refreshSourceMutation,
        $createSource,
        queryClient
      );
      // invalidate the data preview (async)
      // TODO: use new runtime approach
      // Old approach: dataModelerService.dispatch("collectTableInfo", [currentSource?.id]);

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
</script>

<div class="grid items-center" style:grid-template-columns="auto max-content">
  <WorkspaceHeader
    {...{ titleInput: embedded ? path : sourceName, onChangeCallback }}
    showStatus={false}
    let:width
    editable={!embedded}
  >
    <svelte:fragment slot="icon">
      <Source />
    </svelte:fragment>
    <svelte:fragment slot="workspace-controls">
      {#if $refreshSourceMutation.isLoading}
        Refreshing...
      {:else}
        <div class="flex items-center pr-2 gap-x-2">
          {#if $getSource.isSuccess && $getSource.data?.entry?.refreshedOn}
            <div
              class="ui-copy-muted"
              style:font-size="11px"
              transition:fade|local={{ duration: 200 }}
            >
              Imported on {formatRefreshedOn(
                $getSource.data?.entry?.refreshedOn
              )}
            </div>
          {/if}
          {#if connector === "file"}
            <Tooltip location="bottom" distance={8}>
              <div style="transformY(-1px)">
                <IconButton on:click={() => onRefreshClick(sourceName)}>
                  <Import size="15px" />
                </IconButton>
              </div>
              <TooltipContent slot="tooltip-content">
                Import local file to refresh source
              </TooltipContent>
            </Tooltip>
          {:else}
            <Tooltip location="bottom" distance={8}>
              <IconButton on:click={() => onRefreshClick(sourceName)}>
                <RefreshIcon size="15px" />
              </IconButton>
              <TooltipContent slot="tooltip-content">
                Refresh the source data
              </TooltipContent>
            </Tooltip>
          {/if}
        </div>
      {/if}
    </svelte:fragment>
    <svelte:fragment slot="cta">
      <PanelCTA side="right">
        <Tooltip location="left" distance={16}>
          <Button type="secondary" on:click={handleCreateModelFromSource}>
            <ResponsiveButtonText collapse={width < 800}>
              Create Model
            </ResponsiveButtonText>
            <Model size="16px" />
          </Button>
          <TooltipContent slot="tooltip-content">
            Model this source with SQL
          </TooltipContent>
        </Tooltip>
        {#if !embedded}
          <Tooltip location="bottom" alignment="right" distance={16}>
            <Button
              type="primary"
              disabled={!timestampColumns?.length}
              on:click={() => handleCreateDashboardFromSource(sourceName)}
            >
              <ResponsiveButtonText collapse={width < 800}>
                Create Dashboard
              </ResponsiveButtonText>

              <Explore size="16px" />
            </Button>
            <TooltipContent slot="tooltip-content">
              {#if timestampColumns?.length}
                Create a dashboard for this source
              {:else}
                This data source does not have a TIMESTAMP column
              {/if}
            </TooltipContent>
          </Tooltip>
        {/if}
      </PanelCTA>
    </svelte:fragment>
  </WorkspaceHeader>
</div>
