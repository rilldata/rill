<script lang="ts">
  import { DerivedModelEntity } from "$common/data-modeler-state-service/entity-state-service/DerivedModelEntityService";
  import {
    DerivedModelStore,
    PersistentModelStore,
  } from "$lib/application-state-stores/model-stores";
  import Button from "$lib/components/Button.svelte";
  import Tooltip from "$lib/components/tooltip/Tooltip.svelte";
  import TooltipContent from "$lib/components/tooltip/TooltipContent.svelte";
  import { autoCreateMetricsDefinitionForModel } from "$lib/redux-store/source/source-apis";
  import { selectTimestampColumnFromProfileEntity } from "$lib/redux-store/source/source-selectors";
  import { getContext } from "svelte";
  import Explore from "../icons/Explore.svelte";

  export let activeEntityID: string;

  const persistentModelStore = getContext(
    "rill:app:persistent-model-store"
  ) as PersistentModelStore;
  const derivedModelStore = getContext(
    "rill:app:derived-model-store"
  ) as DerivedModelStore;

  let currentDerivedModel: DerivedModelEntity;
  $: currentDerivedModel =
    activeEntityID && $derivedModelStore?.entities
      ? $derivedModelStore.entities.find((q) => q.id === activeEntityID)
      : undefined;

  $: timestampColumns =
    selectTimestampColumnFromProfileEntity(currentDerivedModel);

  const handleCreateMetric = () => {
    // A side effect of the createMetricsDefsApi is we switch active assets to
    // the newly created metrics definition. So, this'll bring us to the
    // MetricsDefinition page. (The logic for this is contained in the
    // not-pictured async thunk.)
    autoCreateMetricsDefinitionForModel(
      $persistentModelStore.entities.find(
        (model) => model.id === activeEntityID
      ).tableName,
      activeEntityID,
      timestampColumns[0].name
    );
  };
</script>

<Tooltip location="bottom" alignment="right" distance={16}>
  <Button
    type="primary"
    disabled={!timestampColumns?.length}
    on:click={handleCreateMetric}
    >Create Dashboard<Explore size="16px" /></Button
  >
  <TooltipContent slot="tooltip-content">
    {#if timestampColumns?.length}
      Auto create metrics based on your model and go to dashboard
    {:else}
      Add a timestamp column to your model in order to create a metric
    {/if}
  </TooltipContent>
</Tooltip>
