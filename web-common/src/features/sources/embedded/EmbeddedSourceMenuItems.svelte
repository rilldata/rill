<script lang="ts">
  import Model from "@rilldata/web-common/components/icons/Model.svelte";
  import RefreshIcon from "@rilldata/web-common/components/icons/RefreshIcon.svelte";
  import { MenuItem } from "@rilldata/web-common/components/menu";
  import { overlay } from "@rilldata/web-common/layout/overlay-store";
  import {
    createRuntimeServicePutFileAndReconcile,
    createRuntimeServiceRefreshAndReconcile,
    getRuntimeServiceGetCatalogEntryQueryKey,
  } from "@rilldata/web-common/runtime-client";
  import { appStore } from "@rilldata/web-local/lib/application-state-stores/app-store";
  import { behaviourEvent } from "@rilldata/web-local/lib/metrics/initMetrics";
  import { BehaviourEventMedium } from "@rilldata/web-local/lib/metrics/service/BehaviourEventTypes";
  import {
    EntityTypeToScreenMap,
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-local/lib/metrics/service/MetricsTypes";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { useModelNames } from "../../models/selectors";
  import { createModelFromSource } from "../createModel";
  import { refreshAndReconcile } from "../refreshSource";

  export let uri: string;
  export let cachedSourceName: string;
  export let connector: string;

  const queryClient = useQueryClient();

  $: runtimeInstanceId = $runtime.instanceId;
  $: modelNames = useModelNames($runtime.instanceId);

  const refreshSourceMutation = createRuntimeServiceRefreshAndReconcile();
  const createFileMutation = createRuntimeServicePutFileAndReconcile();

  const handleCreateModel = async () => {
    try {
      const previousActiveEntity = $appStore.activeEntity?.type;
      const newModelName = await createModelFromSource(
        queryClient,
        runtimeInstanceId,
        $modelNames.data,
        cachedSourceName,
        `"${uri}"`,
        $createFileMutation
      );

      behaviourEvent.fireNavigationEvent(
        newModelName,
        BehaviourEventMedium.Menu,
        MetricsEventSpace.LeftPanel,
        EntityTypeToScreenMap[previousActiveEntity],
        MetricsEventScreenName.Model
      );
    } catch (err) {
      console.error(err);
    }
  };

  const handleRefreshSource = async () => {
    if (!connector) {
      // if parse failed or there is no catalog entry, we cannot refresh source
      // TODO: show the import source modal with fixed tableName
      return;
    }

    try {
      await refreshAndReconcile(
        cachedSourceName,
        runtimeInstanceId,
        $refreshSourceMutation,
        queryClient,
        uri,
        uri
      );

      // invalidate the data preview (async)
      // TODO: use new runtime approach
      // Old approach: dataModelerService.dispatch("collectTableInfo", [currentSource?.id]);

      // invalidate the "refreshed_on" time
      const queryKey = getRuntimeServiceGetCatalogEntryQueryKey(
        runtimeInstanceId,
        cachedSourceName
      );
      await queryClient.refetchQueries(queryKey);
    } catch (err) {
      // no-op
    }
    overlay.set(null);
  };
</script>

<MenuItem icon on:select={handleCreateModel}>
  <Model slot="icon" />
  Create new model
</MenuItem>

<MenuItem icon on:select={handleRefreshSource}>
  <svelte:fragment slot="icon">
    <RefreshIcon />
  </svelte:fragment>
  Refresh source data
</MenuItem>
