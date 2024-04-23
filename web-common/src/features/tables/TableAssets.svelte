<script lang="ts">
  import { slide } from "svelte/transition";
  import CaretDownIcon from "../../components/icons/CaretDownIcon.svelte";
  import Resizer from "../../layout/Resizer.svelte";
  import { LIST_SLIDE_DURATION as duration } from "../../layout/config";
  import { createRuntimeServiceGetInstance } from "../../runtime-client";
  import { runtime } from "../../runtime-client/runtime-store";
  import TableAsset from "./TableAsset.svelte";
  import { useTables } from "./selectors";

  export let startingHeight: number;

  const MIN_HEIGHT = 43; // The height of the "Tables" header

  let showTables = true;
  let sectionHeight = startingHeight;

  $: instance = createRuntimeServiceGetInstance($runtime.instanceId);
  $: connectorInstanceId = $instance.data?.instance?.instanceId;
  $: olapConnector = $instance.data?.instance?.olapConnector;

  $: tablesQuery = useTables(connectorInstanceId, olapConnector);
  $: typedTables = $tablesQuery.data?.tables as
    | {
        name: string;
        database: string;
        databaseSchema: string;
        hasUnsupportedDataTypes: boolean;
      }[]
    | undefined;
</script>

{#if connectorInstanceId && olapConnector}
  <section
    class="flex flex-col border-t border-t-gray-200"
    style:min-height="{MIN_HEIGHT}px"
    style:height="{sectionHeight}px"
  >
    <Resizer
      bind:dimension={sectionHeight}
      direction="NS"
      side="top"
      min={10}
      basis={showTables ? startingHeight : MIN_HEIGHT}
      max={2000}
      absolute={false}
    />
    <button
      class="flex justify-between items-center w-full pl-2 pr-3.5 pt-2 pb-2 text-gray-500"
      on:click={() => {
        showTables = !showTables;
      }}
    >
      <h3 class="font-semibold text-[10px] uppercase">Tables</h3>
      <CaretDownIcon
        className="transform transition-transform {showTables
          ? 'rotate-0'
          : '-rotate-180'}"
      />
    </button>
    <div class="flex flex-col overflow-y-auto">
      {#if showTables}
        <ol transition:slide={{ duration }}>
          {#if typedTables && typedTables.length > 0}
            {#each typedTables as tableInfo (tableInfo)}
              <TableAsset
                {connectorInstanceId}
                connector={olapConnector}
                database={tableInfo.database}
                databaseSchema={tableInfo.databaseSchema}
                table={tableInfo.name}
                hasUnsupportedDataTypes={tableInfo.hasUnsupportedDataTypes}
              />
            {/each}
          {/if}
        </ol>
      {/if}
    </div>
  </section>
{/if}
