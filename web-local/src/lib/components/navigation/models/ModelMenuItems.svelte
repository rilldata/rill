<script lang="ts">
  import { useRuntimeServiceDeleteFileAndReconcile } from "@rilldata/web-common/runtime-client";
  import type { DerivedModelEntity } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/DerivedModelEntityService";
  import { EntityType } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";
  import { BehaviourEventMedium } from "@rilldata/web-local/common/metrics-service/BehaviourEventTypes";
  import {
    EntityTypeToScreenMap,
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-local/common/metrics-service/MetricsTypes";
  import type { ApplicationStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import type {
    DerivedModelStore,
    PersistentModelStore,
  } from "@rilldata/web-local/lib/application-state-stores/model-stores";
  import { autoCreateMetricsDefinitionForModel } from "@rilldata/web-local/lib/redux-store/source/source-apis";
  import {
    derivedProfileEntityHasTimestampColumn,
    selectTimestampColumnFromProfileEntity,
  } from "@rilldata/web-local/lib/redux-store/source/source-selectors";
  import { deleteEntity } from "@rilldata/web-local/lib/svelte-query/actions";
  import { useModelNames } from "@rilldata/web-local/lib/svelte-query/models";
  import { createEventDispatcher, getContext } from "svelte";
  import { runtimeStore } from "../../../application-state-stores/application-store";
  import { navigationEvent } from "../../../metrics/initMetrics";
  import Cancel from "../../icons/Cancel.svelte";
  import EditIcon from "../../icons/EditIcon.svelte";
  import Explore from "../../icons/Explore.svelte";
  import { Divider, MenuItem } from "../../menu";

  export let modelName: string;
  // manually toggle menu to workaround: https://stackoverflow.com/questions/70662482/react-query-mutate-onsuccess-function-not-responding
  export let toggleMenu: () => void;

  const dispatch = createEventDispatcher();

  const deleteModel = useRuntimeServiceDeleteFileAndReconcile();

  const persistentModelStore = getContext(
    "rill:app:persistent-model-store"
  ) as PersistentModelStore;
  const derivedModelStore = getContext(
    "rill:app:derived-model-store"
  ) as DerivedModelStore;
  const applicationStore = getContext("rill:app:store") as ApplicationStore;

  $: modelNames = useModelNames($runtimeStore.instanceId);

  $: persistentModel = $persistentModelStore?.entities?.find(
    (model) => model.tableName === modelName
  );

  $: derivedModel = $derivedModelStore?.entities?.find(
    (model) => model.id === persistentModel?.id
  );

  /** functionality for bootstrapping a dashboard */
  const bootstrapDashboard = (derivedModel: DerivedModelEntity) => {
    const previousActiveEntity = $applicationStore?.activeEntity?.type;

    autoCreateMetricsDefinitionForModel(
      modelName,
      persistentModel?.id,
      selectTimestampColumnFromProfileEntity(derivedModel)[0].name
    ).then((createdMetricsId) => {
      navigationEvent.fireEvent(
        createdMetricsId,
        BehaviourEventMedium.Menu,
        MetricsEventSpace.LeftPanel,
        EntityTypeToScreenMap[previousActiveEntity],
        MetricsEventScreenName.Dashboard
      );
    });
  };

  const handleDeleteModel = async (modelName: string) => {
    await deleteEntity(
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
  on:select={() => bootstrapDashboard(derivedModel)}
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
