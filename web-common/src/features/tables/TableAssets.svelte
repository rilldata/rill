<script lang="ts">
  import { page } from "$app/stores";
  import { flip } from "svelte/animate";
  import { writable } from "svelte/store";
  import { slide } from "svelte/transition";
  import { LIST_SLIDE_DURATION as duration } from "../../layout/config";
  import NavigationEntry from "../../layout/navigation/NavigationEntry.svelte";
  import NavigationHeader from "../../layout/navigation/NavigationHeader.svelte";
  import { debounce } from "../../lib/create-debouncer";
  import {
    V1TableInfo,
    createRuntimeServiceGetInstance,
  } from "../../runtime-client";
  import { runtime } from "../../runtime-client/runtime-store";
  import TableMenuItems from "./TableMenuItems.svelte";
  import UnsupportedTypesIndicator from "./UnsupportedTypesIndicator.svelte";
  import { makeFullyQualifiedTableName, useTables } from "./selectors";

  let showTables = true;

  // Debounce to prevent flickering
  const debouncedTables = writable<V1TableInfo[]>([]);
  const setDebouncedTables = debounce(
    (tables: V1TableInfo[]) => debouncedTables.set(tables),
    200,
  );

  $: if ($tables) {
    setDebouncedTables($tables);
  }

  $: hasAssets = $debouncedTables.length > 0;

  $: instance = createRuntimeServiceGetInstance($runtime.instanceId);
  $: connectorInstanceId = $instance.data?.instance?.instanceId;
  $: olapConnector = $instance.data?.instance?.olapConnector;

  $: tables = useTables(
    $runtime.instanceId,
    connectorInstanceId,
    olapConnector,
  );

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

{#if connectorInstanceId && olapConnector && hasAssets}
  <div class="h-fit flex flex-col">
    <NavigationHeader bind:show={showTables}>Tables</NavigationHeader>

    {#if showTables}
      <ol transition:slide={{ duration }}>
        {#if $debouncedTables.length > 0}
          {#each $debouncedTables as tableInfo (tableInfo)}
            {@const fullyQualifiedTableName = makeFullyQualifiedTableName(
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
                <TableMenuItems slot="menu-items" {fullyQualifiedTableName} />
              </NavigationEntry>
            </li>
          {/each}
        {/if}
      </ol>
    {/if}
  </div>
{/if}
