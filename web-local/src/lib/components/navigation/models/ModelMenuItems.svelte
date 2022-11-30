<script lang="ts">
  import { goto } from "$app/navigation";
  import {
    getRuntimeServiceListFilesQueryKey,
    useRuntimeServiceDeleteFileAndReconcile,
    useRuntimeServicePutFileAndReconcile,
  } from "@rilldata/web-common/runtime-client";
  import { EntityType } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";
  import type { ApplicationStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import type {
    DerivedModelStore,
    PersistentModelStore,
  } from "@rilldata/web-local/lib/application-state-stores/model-stores";
  import { derivedProfileEntityHasTimestampColumn } from "@rilldata/web-local/lib/redux-store/source/source-selectors";
  import { deleteFileArtifact } from "@rilldata/web-local/lib/svelte-query/actions";
  import { useModelNames } from "@rilldata/web-local/lib/svelte-query/models";
  import { createEventDispatcher, getContext } from "svelte";
  import { BehaviourEventMedium } from "../../../../common/metrics-service/BehaviourEventTypes";
  import {
    EntityTypeToScreenMap,
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "../../../../common/metrics-service/MetricsTypes";
  import { getName } from "../../../../common/utils/incrementName";
  import { runtimeStore } from "../../../application-state-stores/application-store";
  import { metricsTemplate } from "../../../application-state-stores/metrics-internal-store";
  import { navigationEvent } from "../../../metrics/initMetrics";
  import { useDashboardNames } from "../../../svelte-query/dashboards";
  import { queryClient } from "../../../svelte-query/globalQueryClient";
  import Cancel from "../../icons/Cancel.svelte";
  import EditIcon from "../../icons/EditIcon.svelte";
  import Explore from "../../icons/Explore.svelte";
  import { Divider, MenuItem } from "../../menu";

  export let modelName: string;
  // manually toggle menu to workaround: https://stackoverflow.com/questions/70662482/react-query-mutate-onsuccess-function-not-responding
  export let toggleMenu: () => void;

  const rillAppStore = getContext("rill:app:store") as ApplicationStore;

  const dispatch = createEventDispatcher();

  const deleteModel = useRuntimeServiceDeleteFileAndReconcile();
  const createFileMutation = useRuntimeServicePutFileAndReconcile();

  const persistentModelStore = getContext(
    "rill:app:persistent-model-store"
  ) as PersistentModelStore;
  const derivedModelStore = getContext(
    "rill:app:derived-model-store"
  ) as DerivedModelStore;
  const applicationStore = getContext("rill:app:store") as ApplicationStore;

  $: modelNames = useModelNames($runtimeStore.instanceId);
  $: dashboardNames = useDashboardNames($runtimeStore.instanceId);

  $: persistentModel = $persistentModelStore?.entities?.find(
    (model) => model.tableName === modelName
  );

  $: derivedModel = $derivedModelStore?.entities?.find(
    (model) => model.id === persistentModel?.id
  );

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
      $runtimeStore.instanceId,
      modelName,
      EntityType.Model,
      $deleteModel,
      $applicationStore.activeEntity,
      $modelNames.data
    );
    // onSettled gets triggered *after* both onSuccess and onError
    toggleMenu();
  };
</script>

<MenuItem
  disabled={!derivedProfileEntityHasTimestampColumn(derivedModel)}
  icon
  on:select={() => createDashboardFromModel(modelName)}
>
  <Explore slot="icon" />
  autogenerate dashboard
  <svelte:fragment slot="description">
    {#if !derivedProfileEntityHasTimestampColumn(derivedModel)}
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
