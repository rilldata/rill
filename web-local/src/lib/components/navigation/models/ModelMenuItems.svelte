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
  import { schemaHasTimestampColumn } from "@rilldata/web-local/lib/svelte-query/column-selectors";
  import { useQueryClient } from "@sveltestack/svelte-query";
  import { createEventDispatcher, getContext } from "svelte";
  import { getName } from "../../../../common/utils/incrementName";
  import { runtimeStore } from "../../../application-state-stores/application-store";
  import { metricsTemplate } from "../../../application-state-stores/metrics-internal-store";
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

  const rillAppStore = getContext("rill:app:store") as ApplicationStore;

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

  // const metricMigrate = useRuntimeServicePutFileAndReconcile();

  /** functionality for bootstrapping a dashboard */
  const createDashboardFromModel = (_: string) => {
    // create dashboard from model
    const newDashboardName = getName(
      `${modelName}_dashboard`,
      $dashboardNames.data
    );
    $createFileMutation.mutate(
      {
        data: {
          instanceId: $runtimeStore.instanceId,
          path: `dashboards/${newDashboardName}.yaml`,
          blob: metricsTemplate, // TODO: compile a real yaml file
          create: true,
          createOnly: true,
          strict: false,
        },
      },
      {
        onSuccess: () => {
          goto(`/dashboard/${newDashboardName}`);
          queryClient.invalidateQueries(
            getRuntimeServiceListFilesQueryKey($runtimeStore.instanceId)
          );
          const previousActiveEntity = $rillAppStore?.activeEntity?.type;
          navigationEvent.fireEvent(
            newDashboardName, // TODO: we're hashing these to get an unique ID for telemetry, right?
            BehaviourEventMedium.Menu,
            MetricsEventSpace.LeftPanel,
            EntityTypeToScreenMap[previousActiveEntity],
            MetricsEventScreenName.Dashboard
          );
        },
        onError: (err) => {
          console.error(err);
        },
      }
    );
    // const previousActiveEntity = $applicationStore?.activeEntity?.type;
    // const metricsLabel = `${modelName}_dashboard`;
    // const generatedYAML = generateMeasuresAndDimension(model, {
    //   display_name: metricsLabel,
    // });
    // await $metricMigrate.mutateAsync({
    //   data: {
    //     instanceId: $runtimeStore.instanceId,
    //     path: `dashboards/${metricsLabel}.yaml`,
    //     blob: generatedYAML,
    //     create: false,
    //   },
    // });
    // navigationEvent.fireEvent(
    //     createdMetricsId,
    //     BehaviourEventMedium.Menu,
    //     MetricsEventSpace.LeftPanel,
    //     EntityTypeToScreenMap[previousActiveEntity],
    //     MetricsEventScreenName.Dashboard
    //   );
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
    // onSettled gets triggered *after* both onSuccess and onError
    toggleMenu();
  };
</script>

<MenuItem
  disabled={!schemaHasTimestampColumn(model?.schema)}
  icon
  on:select={() => createDashboardFromModel(modelName)}
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
