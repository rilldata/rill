<script lang="ts">
  import Explore from "$lib/components/icons/Explore.svelte";
  import Metrics from "$lib/components/icons/Metrics.svelte";
  import Model from "$lib/components/icons/Model.svelte";
  import Source from "$lib/components/icons/Source.svelte";
  import { getContext } from "svelte";
  import {
    DerivedTableStore,
    PersistentTableStore,
  } from "$lib/application-state-stores/table-stores";
  import { selectSourcesWithTimestampColumns } from "$lib/redux-store/source/source-selectors";
  import type { PersistentTableEntity } from "$common/data-modeler-state-service/entity-state-service/PersistentTableEntityService";
  import Button from "$lib/components/Button.svelte";
  import { quickStartSource } from "$lib/redux-store/source/source-apis";
  import { PersistentModelStore } from "$lib/application-state-stores/model-stores";

  const steps = [
    {
      heading: "Import your data source",
      description:
        "Add to your sources by clicking on the + icon, or by dragging a csv or parquet file to this window.",
      icon: Source,
    },
    {
      heading: "Model your sources into one big table",
      description:
        "Build intuition about your sources and use SQL to model them into an analytics-ready resource.",
      icon: Model,
    },
    {
      heading: "Define your metrics and dimensions",
      description:
        "Define aggregate metrics and break out dimensions for your modeled data.",
      icon: Metrics,
    },
    {
      heading: "Explore your metrics dashboard",
      description:
        "Interactively explore line charts and leaderboards to uncover insights.",
      icon: Explore,
    },
  ];

  const persistentTableStore = getContext(
    "rill:app:persistent-table-store"
  ) as PersistentTableStore;
  const derivedTableStore = getContext(
    "rill:app:derived-table-store"
  ) as DerivedTableStore;
  let sourcesWithTimestampColumns: Array<PersistentTableEntity>;
  $: sourcesWithTimestampColumns = selectSourcesWithTimestampColumns(
    $persistentTableStore.entities,
    $derivedTableStore.entities
  );

  const persistentModelStore = getContext(
    "rill:app:persistent-model-store"
  ) as PersistentModelStore;

  const quickStartMetrics = async (id: string, tableName: string) => {
    await quickStartSource(
      $persistentModelStore.entities,
      $derivedTableStore.entities,
      id,
      tableName
    );
  };
</script>

<div class="mt-10 p-2 place-content-center">
  <div class="text-center">
    <div class="font-bold">Getting started</div>
    <p>Building data intuition at every step of analysis</p>
  </div>
  <div class="p-5 pt-2">
    {#each steps as step (step.heading)}
      <div
        class="flex items-center p-6 mt-3 bg-gray-50 rounded-lg border border-gray-200"
      >
        <div>
          <svelte:component this={step.icon} color="grey" size="3em" />
        </div>
        <div class="ml-5">
          <h5 class="font-bold">{step.heading}</h5>
          <p class="italic">{step.description}</p>
          {#if step.heading.includes("Explore") && sourcesWithTimestampColumns?.length}
            {#each sourcesWithTimestampColumns as persistentSource (persistentSource.id)}
              <p class="p-1">
                <Button
                  type="secondary"
                  on:click={() =>
                    quickStartMetrics(
                      persistentSource.id,
                      persistentSource.tableName
                    )}
                >
                  Quick start for {persistentSource.tableName}
                </Button>
              </p>
            {/each}
          {/if}
        </div>
      </div>
    {/each}
  </div>
</div>
