<script lang="ts">
  import { page } from "$app/stores";
  import { flip } from "svelte/animate";
  import { slide } from "svelte/transition";
  import { LIST_SLIDE_DURATION as duration } from "../../layout/config";
  import NavigationEntry from "../../layout/navigation/NavigationEntry.svelte";
  import NavigationHeader from "../../layout/navigation/NavigationHeader.svelte";
  import {
    V1TableInfo,
    createRuntimeServiceGetInstance,
  } from "../../runtime-client";
  import { runtime } from "../../runtime-client/runtime-store";
  import TableMenuItems from "./TableMenuItems.svelte";
  import UnsupportedTypesIndicator from "./UnsupportedTypesIndicator.svelte";
  import { makeFullyQualifiedTableName } from "./olap-config";
  import { useTables } from "./selectors";

  let showTables = true;

  $: instance = createRuntimeServiceGetInstance($runtime.instanceId);
  $: connectorInstanceId = $instance.data?.instance?.instanceId;
  $: olapConnector = $instance.data?.instance?.olapConnector;

  $: tables = useTables(
    $runtime.instanceId,
    connectorInstanceId,
    olapConnector,
  );
  $: hasAssets = $tables?.length > 0;

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
      case "pinot":
        return `/connector/pinot/${tableInfo.name}`;
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
        {#if $tables.length > 0}
          {#each $tables as tableInfo (tableInfo)}
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
{/if}
