<script lang="ts">
  import type { DerivedModelEntity } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/DerivedModelEntityService";
  import { EntityType } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";
  import { getNextEntityId } from "@rilldata/web-local/common/utils/getNextEntityId";
  import type { ApplicationStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import type {
    DerivedModelStore,
    PersistentModelStore,
  } from "@rilldata/web-local/lib/application-state-stores/model-stores";
  import { deleteModelApi } from "@rilldata/web-local/lib/redux-store/model/model-apis";
  import { autoCreateMetricsDefinitionForModel } from "@rilldata/web-local/lib/redux-store/source/source-apis";
  import {
    derivedProfileEntityHasTimestampColumn,
    selectTimestampColumnFromProfileEntity,
  } from "@rilldata/web-local/lib/redux-store/source/source-selectors";
  import { createEventDispatcher, getContext } from "svelte";
  import Cancel from "../../icons/Cancel.svelte";
  import EditIcon from "../../icons/EditIcon.svelte";
  import Explore from "../../icons/Explore.svelte";

  import { navigationEvent } from "../../../metrics/initMetrics";

  import { goto } from "$app/navigation";
  import { BehaviourEventMedium } from "@rilldata/web-local/common/metrics-service/BehaviourEventTypes";
  import {
    EntityTypeToScreenMap,
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-local/common/metrics-service/MetricsTypes";
  import { Divider, MenuItem } from "../../menu";

  export let modelID;

  const dispatch = createEventDispatcher();

  const persistentModelStore = getContext(
    "rill:app:persistent-model-store"
  ) as PersistentModelStore;
  const derivedModelStore = getContext(
    "rill:app:derived-model-store"
  ) as DerivedModelStore;
  const applicationStore = getContext("rill:app:store") as ApplicationStore;

  $: persistentModel = $persistentModelStore?.entities?.find(
    (model) => model.id === modelID
  );

  $: derivedModel = $derivedModelStore?.entities?.find(
    (model) => model.id === modelID
  );

  /** functionality for bootstrapping a dashboard */
  const bootstrapDashboard = (derivedModel: DerivedModelEntity) => {
    const previousActiveEntity = $applicationStore?.activeEntity?.type;

    autoCreateMetricsDefinitionForModel(
      persistentModel.tableName,
      modelID,
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

  /** delete model */
  const deleteModel = (id: string) => {
    if (
      $applicationStore.activeEntity.type === EntityType.Model &&
      $applicationStore.activeEntity.id === id
    ) {
      const nextModelId = getNextEntityId($persistentModelStore.entities, id);

      if (nextModelId) {
        goto(`/model/${nextModelId}`);
      } else {
        goto("/");
      }
    }

    deleteModelApi(id);
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
<MenuItem icon on:select={() => deleteModel(derivedModel.id)}>
  <Cancel slot="icon" />
  delete
</MenuItem>
