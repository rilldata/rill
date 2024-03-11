<script lang="ts">
  import { page } from "$app/stores";
  import { flip } from "svelte/animate";
  import { slide } from "svelte/transition";
  import { LIST_SLIDE_DURATION } from "../../layout/config";
  import NavigationEntry from "../../layout/navigation/NavigationEntry.svelte";
  import NavigationHeader from "../../layout/navigation/NavigationHeader.svelte";
  import {
    createConnectorServiceOLAPListTables,
    createRuntimeServiceGetInstance,
  } from "../../runtime-client";
  import { runtime } from "../../runtime-client/runtime-store";
  import {
    ResourceKind,
    useFilteredResourceNames,
  } from "../entity-management/resource-selectors";
  import TableMenuItems from "./TableMenuItems.svelte";

  $: instance = createRuntimeServiceGetInstance($runtime.instanceId);
  $: connectorInstanceId = $instance.data?.instance?.instanceId;
  $: olapConnector = $instance.data?.instance?.olapConnector;

  // Get managed table names
  $: sourceNamesQuery = useFilteredResourceNames(
    $runtime.instanceId,
    ResourceKind.Source,
  );
  $: modelNamesQuery = useFilteredResourceNames(
    $runtime.instanceId,
    ResourceKind.Model,
  );
  $: sourceNames = $sourceNamesQuery.data;
  $: modelNames = $modelNamesQuery.data;

  $: tableNames = createConnectorServiceOLAPListTables(
    {
      instanceId: connectorInstanceId,
      connector: olapConnector,
    },
    {
      query: {
        enabled:
          !!connectorInstanceId &&
          !!olapConnector &&
          !!sourceNames &&
          !!modelNames,
        select: (data) => {
          // If sourceNames or modelNames are not available, return an empty array
          if (!sourceNames || !modelNames) {
            return [];
          }

          // Filter out managed tables (sources and models)
          const filteredTables = data?.tables?.filter(
            (table) =>
              !(sourceNames as string[]).includes(table.name as string) &&
              !(modelNames as string[]).includes(table.name as string),
          );

          // Return the fully qualified table names
          return (
            filteredTables?.map((table) => table.database + "." + table.name) ||
            []
          );
        },
      },
    },
  );

  let showTables = true;

  $: hasAssets = $tableNames.data && $tableNames.data.length > 0;
</script>

{#if hasAssets}
  <NavigationHeader bind:show={showTables} toggleText="tables">
    Tables
  </NavigationHeader>

  {#if showTables}
    <div
      class="pb-3 max-h-96 overflow-auto"
      transition:slide={{ duration: LIST_SLIDE_DURATION }}
    >
      {#if $tableNames?.data}
        {#each $tableNames.data as fullyQualifiedTableName (fullyQualifiedTableName)}
          <div
            animate:flip={{ duration: 200 }}
            out:slide|global={{ duration: LIST_SLIDE_DURATION }}
          >
            <NavigationEntry
              name={fullyQualifiedTableName}
              href={`/table/${fullyQualifiedTableName}`}
              open={$page.url.pathname === `/table/${fullyQualifiedTableName}`}
              expandable={false}
            >
              <svelte:fragment slot="menu-items" let:toggleMenu>
                <TableMenuItems {fullyQualifiedTableName} {toggleMenu} />
              </svelte:fragment>
            </NavigationEntry>
          </div>
        {/each}
      {/if}
    </div>
  {/if}
{/if}
