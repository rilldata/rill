<script lang="ts">
  import { goto } from "$app/navigation";
  import {
    getRuntimeServiceListFilesQueryKey,
    useRuntimeServiceDeleteFileAndReconcile,
    useRuntimeServiceGetCatalogEntry,
    useRuntimeServicePutFileAndReconcile,
    V1Model,
  } from "@rilldata/web-common/runtime-client";
  import { EntityType } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";
  import { appStore } from "@rilldata/web-local/lib/application-state-stores/app-store";
  import { fileArtifactsStore } from "@rilldata/web-local/lib/application-state-stores/file-artifacts-store";
  import { schemaHasTimestampColumn } from "@rilldata/web-local/lib/svelte-query/column-selectors";
  import { useQueryClient } from "@sveltestack/svelte-query";
  import { createEventDispatcher } from "svelte";
  import { getName } from "../../../../common/utils/incrementName";
  import { runtimeStore } from "../../../application-state-stores/application-store";
  import { generateMeasuresAndDimension } from "../../../application-state-stores/metrics-internal-store";
  import { overlay } from "../../../application-state-stores/overlay-store";
  import { navigationEvent } from "../../../metrics/initMetrics";
  import { BehaviourEventMedium } from "../../../metrics/service/BehaviourEventTypes";
  import {
    EntityTypeToScreenMap,
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "../../../metrics/service/MetricsTypes";
  import { deleteFileArtifact } from "../../../svelte-query/actions";
  import { useDashboardNames } from "../../../svelte-query/dashboards";
  import { useModelNames } from "../../../svelte-query/models";
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
    const generatedYAML = generateMeasuresAndDimension(model, {
      display_name: `${newDashboardName} dashboard`,
      description: `A dashboard generated for ${modelName}`,
    });
    $createFileMutation.mutate(
      {
        data: {
          instanceId: $runtimeStore.instanceId,
          path: `dashboards/${newDashboardName}.yaml`,
          blob: generatedYAML,
          create: true,
          createOnly: true,
          strict: false,
        },
      },
      {
        onSuccess: (resp) => {
          fileArtifactsStore.setErrors(resp.affectedPaths, resp.errors);
          goto(`/dashboard/${newDashboardName}`);
          queryClient.invalidateQueries(
            getRuntimeServiceListFilesQueryKey($runtimeStore.instanceId)
          );
          const previousActiveEntity = $appStore?.activeEntity?.type;
          navigationEvent.fireEvent(
            newDashboardName,
            BehaviourEventMedium.Menu,
            MetricsEventSpace.LeftPanel,
            EntityTypeToScreenMap[previousActiveEntity],
            MetricsEventScreenName.Dashboard
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
  autogenerate dashboard
  <svelte:fragment slot="description">
    {#if !schemaHasTimestampColumn(model?.schema)}
      requires a timestamp column
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
  rename...
</MenuItem>
<MenuItem
  icon
  on:select={() => handleDeleteModel(modelName)}
  propogateSelect={false}
>
  <Cancel slot="icon" />
  delete
</MenuItem>
