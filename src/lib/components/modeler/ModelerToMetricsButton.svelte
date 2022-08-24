<script lang="ts">
  import type { DerivedModelEntity } from "$common/data-modeler-state-service/entity-state-service/DerivedModelEntityService";
  import type {
    DerivedModelStore,
    PersistentModelStore,
  } from "$lib/application-state-stores/model-stores";
  import { Button } from "$lib/components/button";
  import Tooltip from "$lib/components/tooltip/Tooltip.svelte";
  import TooltipContent from "$lib/components/tooltip/TooltipContent.svelte";
  import { autoCreateMetricsDefinitionForModel } from "$lib/redux-store/source/source-apis";
  import { selectTimestampColumnFromProfileEntity } from "$lib/redux-store/source/source-selectors";
  import { getContext } from "svelte";
  import Explore from "../icons/Explore.svelte";
  import ResponsiveButtonText from "../panel/ResponsiveButtonText.svelte";

  export let activeEntityID: string;
  export let hasError = false;
  export let width = undefined;

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
  >
    <ResponsiveButtonText {width}>Create Dashboard</ResponsiveButtonText>
    <Explore size="16px" /></Button
  >
  <TooltipContent slot="tooltip-content">
    {#if hasError}
      Fix the errors in your model to autogenerate dashboard
    {:else if timestampColumns?.length}
      Generate a dashboard based on your model
    {:else}
      Add a timestamp column to your model in order to generate a dashboard
    {/if}
  </TooltipContent>
</Tooltip>
