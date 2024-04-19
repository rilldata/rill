<script lang="ts">
  import { page } from "$app/stores";
  import { flip } from "svelte/animate";
  import { slide } from "svelte/transition";
  import CaretDownIcon from "../../components/icons/CaretDownIcon.svelte";
  import Resizer from "../../layout/Resizer.svelte";
  import { LIST_SLIDE_DURATION as duration } from "../../layout/config";
  import NavigationEntry from "../../layout/navigation/NavigationEntry.svelte";
  import {
    V1TableInfo,
    createRuntimeServiceGetInstance,
  } from "../../runtime-client";
  import { runtime } from "../../runtime-client/runtime-store";
  import TableMenuItems from "./TableMenuItems.svelte";
  import UnsupportedTypesIndicator from "./UnsupportedTypesIndicator.svelte";
  import { makeFullyQualifiedTableName } from "./olap-config";
  import { useTables } from "./selectors";

  export let startingHeight: number;

  const MIN_HEIGHT = 43; // The height of the "Tables" header

  let showTables = true;
  let sectionHeight = startingHeight;

  $: instance = createRuntimeServiceGetInstance($runtime.instanceId);
  $: connectorInstanceId = $instance.data?.instance?.instanceId;
  $: olapConnector = $instance.data?.instance?.olapConnector;

  $: tablesQuery = useTables(connectorInstanceId, olapConnector);
  $: tables = $tablesQuery.data?.tables;

  function getTableRouteForOLAPConnector(
    olapConnector: string,
    tableInfo: V1TableInfo,
  ): string {
    switch (olapConnector) {
      case "clickhouse":
        return `/connector/clickhouse/${tableInfo.databaseSchema}/${tableInfo.name}`;
      case "druid":
        return `/connector/druid/${tableInfo.databaseSchema}/${tableInfo.name}`;
      case "duckdb":
        return `/connector/duckdb/${tableInfo.database}/${tableInfo.databaseSchema}/${tableInfo.name}`;
      default:
        throw new Error(`Unsupported OLAP connector: ${olapConnector}`);
    }
  }
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
      basis={MIN_HEIGHT}
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
    <div class="h-fit flex flex-col overflow-y-auto">
      {#if showTables}
        <ol transition:slide={{ duration }}>
          {#if tables && tables.length > 0}
            {#each tables as tableInfo (tableInfo)}
              {@const fullyQualifiedTableName = makeFullyQualifiedTableName(
                olapConnector,
                tableInfo.database ?? "",
                tableInfo.databaseSchema ?? "",
                tableInfo.name ?? "",
              )}
              {@const tableRoute = getTableRouteForOLAPConnector(
                olapConnector,
                tableInfo,
              )}
              <li
                animate:flip={{ duration }}
                aria-label={fullyQualifiedTableName}
              >
                <NavigationEntry
                  name={fullyQualifiedTableName}
                  href={tableRoute}
                  open={$page.url.pathname === tableRoute}
                >
                  <svelte:fragment slot="icon">
                    {#if tableInfo.hasUnsupportedDataTypes}
                      <UnsupportedTypesIndicator
                        instanceId={connectorInstanceId}
                        connector={olapConnector}
                        {tableInfo}
                      />
                    {/if}
                  </svelte:fragment>
                  <TableMenuItems
                    slot="menu-items"
                    connector={olapConnector}
                    database={tableInfo.database}
                    databaseSchema={tableInfo.databaseSchema ?? ""}
                    table={tableInfo.name ?? ""}
                  />
                </NavigationEntry>
              </li>
            {/each}
          {/if}
        </ol>
      {/if}
    </div>
  </section>
{/if}
