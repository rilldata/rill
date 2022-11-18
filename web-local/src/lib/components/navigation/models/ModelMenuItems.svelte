<script lang="ts">
  import { goto } from "$app/navigation";
  import {
    getRuntimeServiceListFilesQueryKey,
    useRuntimeServiceDeleteFileAndMigrate,
  } from "@rilldata/web-common/runtime-client";
  import type { DerivedModelEntity } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/DerivedModelEntityService";
  import { EntityType } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";
  import { BehaviourEventMedium } from "@rilldata/web-local/common/metrics-service/BehaviourEventTypes";
  import {
    EntityTypeToScreenMap,
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-local/common/metrics-service/MetricsTypes";
  import { getNextEntityId } from "@rilldata/web-local/common/utils/getNextEntityId";
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
  import { createEventDispatcher, getContext } from "svelte";
  import { runtimeStore } from "../../../application-state-stores/application-store";
  import { navigationEvent } from "../../../metrics/initMetrics";
  import { queryClient } from "../../../svelte-query/globalQueryClient";
  import Cancel from "../../icons/Cancel.svelte";
  import EditIcon from "../../icons/EditIcon.svelte";
  import Explore from "../../icons/Explore.svelte";
  import { Divider, MenuItem } from "../../menu";

  export let modelName: string;
  // manually toggle menu to workaround: https://stackoverflow.com/questions/70662482/react-query-mutate-onsuccess-function-not-responding
  export let toggleMenu: () => void;

  const dispatch = createEventDispatcher();

  const deleteModel = useRuntimeServiceDeleteFileAndMigrate();

  const persistentModelStore = getContext(
    "rill:app:persistent-model-store"
  ) as PersistentModelStore;
  const derivedModelStore = getContext(
    "rill:app:derived-model-store"
  ) as DerivedModelStore;
  const applicationStore = getContext("rill:app:store") as ApplicationStore;

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

  const handleDeleteModel = (modelName: string) => {
    $deleteModel.mutate(
      {
        data: {
          repoId: $runtimeStore.repoId,
          instanceId: $runtimeStore.instanceId,
          path: `models/${modelName}.sql`,
        },
      },
      {
        onSuccess: () => {
          if (
            $applicationStore.activeEntity.type === EntityType.Model &&
            $applicationStore.activeEntity.id === persistentModel.id
          ) {
            const nextModelId = getNextEntityId(
              $persistentModelStore.entities,
              persistentModel.id
            );
            const nextModelName = $persistentModelStore.entities.find(
              (source) => source.id === nextModelId
            ).tableName;
            if (nextModelName) {
              goto(`/model/${nextModelName}`);
            } else {
              goto("/");
            }
          }

          queryClient.invalidateQueries(
            getRuntimeServiceListFilesQueryKey($runtimeStore.repoId)
          );
        },
        onSettled: () => {
          // onSettled gets triggered *after* both onSuccess and onError
          toggleMenu();
        },
      }
    );
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
  propogateSelect={false}
  on:select={() => handleDeleteModel(modelName)}
>
  <Cancel slot="icon" />
  delete
</MenuItem>
