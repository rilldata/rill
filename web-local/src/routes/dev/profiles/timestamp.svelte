<script lang="ts">
  /**
   * This route will plot all the timestamp detail charts for all the sources
   * you've ingested into Rill Developer. It's a useful way to work on the components
   * in an isolated way, with a number of real-world use-cases.
   */
  import type {
    DerivedTableStore,
    PersistentTableStore,
  } from "$web-local/lib/application-state-stores/table-stores";
  import {
    TimestampDetail,
    TimestampSpark,
  } from "$web-local/lib/components/data-graphic/compositions/timestamp-profile";
  import { TIMESTAMPS } from "$web-local/lib/duckdb-data-types";

  import { getContext } from "svelte";

  const persistentTableStore = getContext(
    "rill:app:persistent-table-store"
  ) as PersistentTableStore;
  const derivedTableStore = getContext(
    "rill:app:derived-table-store"
  ) as DerivedTableStore;

  function convertTimestampPreview(d) {
    return d.map((di) => {
      const pi = { ...di };
      pi.ts = new Date(pi.ts);
      return pi;
    });
  }
</script>

<h1 class="text-xl pb-6">All Available TIMESTAMP columns</h1>
<div class="flex flex-wrap gap-8">
  {#each $persistentTableStore?.entities || [] as table}
    {@const derivedTable = $derivedTableStore.entities.find(
      (t) => t.id == table.id
    )}
    {#if derivedTable && derivedTable?.profile}
      {#each derivedTable.profile as column}
        <!-- {#if TIMESTAMPS.has(column.type) && column?.summary?.rollup}{JSON.stringify(
          column?.summary
        )}{/if} -->
        {#if TIMESTAMPS.has(column.type) && column?.summary?.rollup}
          <div>
            <h2 class="pb-3">
              <div>{column.name}</div>
              <div
                style:grid-template-columns="max-content max-content"
                class="font-normal grid justify-between justify-items-stretch text-gray-600 italic"
              >
                from {table.tableName}
                <TimestampSpark
                  data={convertTimestampPreview(column.summary.rollup.spark)}
                  xAccessor="ts"
                  yAccessor="count"
                  width={98}
                  height={18}
                  top={0}
                  bottom={0}
                  left={0}
                  right={0}
                  leftBuffer={0}
                  rightBuffer={0}
                  area
                  tweenIn
                />
              </div>
            </h2>
            <TimestampDetail
              type={column.type}
              data={column.summary.rollup.results.map((di) => {
                const pi = { ...di };
                pi.ts = new Date(pi.ts);
                return pi;
              })}
              spark={column.summary.rollup.spark.map((di) => {
                const pi = { ...di };
                pi.ts = new Date(pi.ts);
                return pi;
              })}
              xAccessor="ts"
              yAccessor="count"
              mouseover={true}
              height={160}
              width={350}
              rollupGrain={column.summary.rollup.rollupInterval}
              estimatedSmallestTimeGrain={column.summary
                ?.estimatedSmallestTimeGrain}
              interval={column.summary.interval}
            />
          </div>
        {/if}
      {/each}
    {/if}
  {/each}
</div>
