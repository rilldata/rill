<script lang="ts">
  import {
    useRuntimeServiceDeleteFileAndReconcile,
    useRuntimeServiceGetCatalogEntry,
    V1Model,
  } from "@rilldata/web-common/runtime-client";
  import { appStore } from "@rilldata/web-local/lib/application-state-stores/app-store";
  import { schemaHasTimestampColumn } from "@rilldata/web-local/lib/svelte-query/column-selectors";
  import { deleteFileArtifact } from "@rilldata/web-local/lib/svelte-query/actions";
  import { useModelNames } from "@rilldata/web-local/lib/svelte-query/models";
  import { EntityType } from "@rilldata/web-local/lib/temp/entity";
  import { useQueryClient } from "@sveltestack/svelte-query";
  import { createEventDispatcher } from "svelte";
  import { runtimeStore } from "../../../application-state-stores/application-store";
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

  $: modelNames = useModelNames($runtimeStore.instanceId);
  $: modelQuery = useRuntimeServiceGetCatalogEntry(
    $runtimeStore.instanceId,
    modelName
  );
  let model: V1Model;
  $: model = $modelQuery.data?.entry?.model;

  // const metricMigrate = useRuntimeServicePutFileAndReconcile();

  /** functionality for bootstrapping a dashboard */
  const bootstrapDashboard = (_: string) => {
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
  on:select={() => bootstrapDashboard(modelName)}
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
