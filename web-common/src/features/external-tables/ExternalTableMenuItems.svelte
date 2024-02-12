<script lang="ts">
  import { goto } from "$app/navigation";
  import Explore from "@rilldata/web-common/components/icons/Explore.svelte";
  import Model from "@rilldata/web-common/components/icons/Model.svelte";
  import { MenuItem } from "@rilldata/web-common/components/menu";
  import { useDashboardFileNames } from "@rilldata/web-common/features/dashboards/selectors";
  import { getFilePathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { getFileHasErrors } from "@rilldata/web-common/features/entity-management/resources-store";
  import { useModelFileNames } from "@rilldata/web-common/features/models/selectors";
  import { useSource } from "@rilldata/web-common/features/sources/selectors";
  import { appScreen } from "@rilldata/web-common/layout/app-store";
  import { overlay } from "@rilldata/web-common/layout/overlay-store";
  import { waitUntil } from "@rilldata/web-common/lib/waitUtils";
  import { behaviourEvent } from "@rilldata/web-common/metrics/initMetrics";
  import { BehaviourEventMedium } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
  import {
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-common/metrics/service/MetricsTypes";
  import type { V1SourceV2 } from "@rilldata/web-common/runtime-client";
  import { V1ReconcileStatus } from "@rilldata/web-common/runtime-client";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { runtime } from "../../runtime-client/runtime-store";
  import { getName } from "../entity-management/name-utils";
  import { EntityType } from "../entity-management/types";
  import { useCreateDashboardFromSource } from "../sources/createDashboard";
  import { createModelFromSource } from "../sources/createModel";

  export let tableName: string;
  // manually toggle menu to workaround: https://stackoverflow.com/questions/70662482/react-query-mutate-onsuccess-function-not-responding
  export let toggleMenu: () => void;
  $: filePath = getFilePathFromNameAndType(tableName, EntityType.Table);

  const queryClient = useQueryClient();

  $: runtimeInstanceId = $runtime.instanceId;

  $: sourceQuery = useSource(runtimeInstanceId, tableName);
  let source: V1SourceV2 | undefined;
  $: source = $sourceQuery.data?.source;
  $: embedded = false; // TODO: remove embedded support
  $: path = source?.spec?.properties?.path;
  $: sourceHasError = getFileHasErrors(
    queryClient,
    runtimeInstanceId,
    filePath,
  );
  $: sourceIsIdle =
    $sourceQuery.data?.meta?.reconcileStatus ===
    V1ReconcileStatus.RECONCILE_STATUS_IDLE;
  $: disableCreateDashboard = $sourceHasError || !sourceIsIdle;

  $: modelNames = useModelFileNames($runtime.instanceId);
  $: dashboardNames = useDashboardFileNames($runtime.instanceId);

  const createDashboardFromSourceMutation = useCreateDashboardFromSource();

  const handleCreateModel = async () => {
    try {
      const previousActiveEntity = $appScreen?.type;
      const newModelName = await createModelFromSource(
        runtimeInstanceId,
        $modelNames.data ?? [],
        tableName,
        embedded ? `"${path}"` : tableName,
      );

      behaviourEvent.fireNavigationEvent(
        newModelName,
        BehaviourEventMedium.Menu,
        MetricsEventSpace.LeftPanel,
        previousActiveEntity,
        MetricsEventScreenName.Model,
      );
    } catch (err) {
      console.error(err);
    }
  };

  const handleCreateDashboardFromSource = async (sourceName: string) => {
    overlay.set({
      title: "Creating a dashboard for " + sourceName,
    });
    const newModelName = getName(`${sourceName}_model`, $modelNames.data ?? []);
    const newDashboardName = getName(
      `${sourceName}_dashboard`,
      $dashboardNames.data ?? [],
    );

    await waitUntil(() => !!$sourceQuery.data);
    if (!$sourceQuery.data) {
      // this should never happen because of above `waitUntil`,
      // but adding this guard provides type narrowing below
      return;
    }

    $createDashboardFromSourceMutation.mutate(
      {
        data: {
          instanceId: $runtime.instanceId,
          sourceResource: $sourceQuery.data,
          newModelName,
          newDashboardName,
        },
      },
      {
        onSuccess: async () => {
          goto(`/dashboard/${newDashboardName}`);
          const previousActiveEntity = $appScreen?.type;
          behaviourEvent.fireNavigationEvent(
            newDashboardName,
            BehaviourEventMedium.Menu,
            MetricsEventSpace.LeftPanel,
            previousActiveEntity,
            MetricsEventScreenName.Dashboard,
          );
        },
        onSettled: () => {
          overlay.set(null);
          toggleMenu(); // unmount component
        },
      },
    );
  };
</script>

<!-- TODO: exclude if Druid -->
<MenuItem icon on:select={() => handleCreateModel()}>
  <Model slot="icon" />
  Create new model
</MenuItem>

<MenuItem
  disabled={disableCreateDashboard}
  icon
  on:select={() => handleCreateDashboardFromSource(tableName)}
  propogateSelect={false}
>
  <Explore slot="icon" />
  Autogenerate dashboard
  <svelte:fragment slot="description">
    {#if $sourceHasError}
      Source has errors
    {:else if !sourceIsIdle}
      Source is being ingested
    {/if}
  </svelte:fragment>
</MenuItem>
