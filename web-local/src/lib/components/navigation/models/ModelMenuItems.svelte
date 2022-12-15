<script lang="ts">
  import { goto } from "$app/navigation";
  import {
    useRuntimeServiceDeleteFileAndReconcile,
    useRuntimeServiceGetCatalogEntry,
    useRuntimeServicePutFileAndReconcile,
    V1Model,
    V1ReconcileResponse,
  } from "@rilldata/web-common/runtime-client";
  import { appStore } from "@rilldata/web-local/lib/application-state-stores/app-store";
  import { fileArtifactsStore } from "@rilldata/web-local/lib/application-state-stores/file-artifacts-store";
  import { deleteFileArtifact } from "@rilldata/web-local/lib/svelte-query/actions";
  import { schemaHasTimestampColumn } from "@rilldata/web-local/lib/svelte-query/column-selectors";
  import { invalidateAfterReconcile } from "@rilldata/web-local/lib/svelte-query/invalidation";
  import { useModelNames } from "@rilldata/web-local/lib/svelte-query/models";
  import { EntityType } from "@rilldata/web-local/lib/temp/entity";
  import { getFilePathFromNameAndType } from "@rilldata/web-local/lib/util/entity-mappers";
  import { getName } from "@rilldata/web-local/lib/util/incrementName";
  import { useQueryClient } from "@sveltestack/svelte-query";
  import { createEventDispatcher } from "svelte";
  import { runtimeStore } from "../../../application-state-stores/application-store";
  import {
    addQuickMetricsToDashboardYAML,
    initBlankDashboardYAML,
  } from "../../../application-state-stores/metrics-internal-store";
  import { overlay } from "../../../application-state-stores/overlay-store";
  import { navigationEvent } from "../../../metrics/initMetrics";
  import { BehaviourEventMedium } from "../../../metrics/service/BehaviourEventTypes";
  import {
    EntityTypeToScreenMap,
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "../../../metrics/service/MetricsTypes";
  import { useDashboardNames } from "../../../svelte-query/dashboards";
  import Cancel from "../../icons/Cancel.svelte";
  import EditIcon from "../../icons/EditIcon.svelte";
  import Explore from "../../icons/Explore.svelte";
  import { Divider, MenuItem } from "../../menu";

  export let modelName: string;
  // manually toggle menu to workaround: https://stackoverflow.com/questions/70662482/react-query-mutate-onsuccess-function-not-responding
  export let toggleMenu: () => void;

  const dispatch = createEventDispatcher();

  const queryClient = useQueryClient();

  const deleteModel = useRuntimeServiceDeleteFileAndReconcile();
  const createFileMutation = useRuntimeServicePutFileAndReconcile();

  $: modelNames = useModelNames($runtimeStore.instanceId);
  $: dashboardNames = useDashboardNames($runtimeStore.instanceId);
  $: modelQuery = useRuntimeServiceGetCatalogEntry(
    $runtimeStore.instanceId,
    modelName
  );
  let model: V1Model;
  $: model = $modelQuery.data?.entry?.model;

  const createDashboardFromModel = (modelName: string) => {
    overlay.set({
      title: "Creating a dashboard for " + modelName,
    });
    const newDashboardName = getName(
      `${modelName}_dashboard`,
      $dashboardNames.data
    );
    const blankDashboardYAML = initBlankDashboardYAML(newDashboardName);
    const fullDashboardYAML = addQuickMetricsToDashboardYAML(
      blankDashboardYAML,
      model
    );
    $createFileMutation.mutate(
      {
        data: {
          instanceId: $runtimeStore.instanceId,
          path: getFilePathFromNameAndType(
            newDashboardName,
            EntityType.MetricsDefinition
          ),
          blob: fullDashboardYAML,
          create: true,
          createOnly: true,
          strict: false,
        },
      },
      {
        onSuccess: (resp: V1ReconcileResponse) => {
          fileArtifactsStore.setErrors(resp.affectedPaths, resp.errors);
          goto(`/dashboard/${newDashboardName}`);
          const previousActiveEntity = $appStore?.activeEntity?.type;
          navigationEvent.fireEvent(
            newDashboardName,
            BehaviourEventMedium.Menu,
            MetricsEventSpace.LeftPanel,
            EntityTypeToScreenMap[previousActiveEntity],
            MetricsEventScreenName.Dashboard
          );
          return invalidateAfterReconcile(
            queryClient,
            $runtimeStore.instanceId,
            resp
          );
        },
        onError: (err) => {
          console.error(err);
        },
        onSettled: () => {
          overlay.set(null);
          toggleMenu(); // unmount component
        },
      }
    );
  };

  const handleDeleteModel = async (modelName: string) => {
    await deleteFileArtifact(
      queryClient,
      $runtimeStore.instanceId,
      modelName,
      EntityType.Model,
      $deleteModel,
      $appStore.activeEntity,
      $modelNames.data
    );
    toggleMenu();
  };
</script>

<MenuItem
  disabled={!schemaHasTimestampColumn(model?.schema)}
  icon
  on:select={() => createDashboardFromModel(modelName)}
  propogateSelect={false}
>
  <Explore slot="icon" />
  Autogenerate dashboard
  <svelte:fragment slot="description">
    {#if !schemaHasTimestampColumn(model?.schema)}
      Requires a timestamp column
    {/if}
  </svelte:fragment>
</MenuItem>
<Divider />
<MenuItem
  icon
  on:select={() => {
    dispatch("rename-asset");
  }}
>
  <EditIcon slot="icon" />
  Rename...
</MenuItem>
<MenuItem
  icon
  on:select={() => handleDeleteModel(modelName)}
  propogateSelect={false}
>
  <Cancel slot="icon" />
  Delete
</MenuItem>
