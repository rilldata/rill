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
  import ExternalTableMenuItems from "./ExternalTableMenuItems.svelte";

  $: instance = createRuntimeServiceGetInstance($runtime.instanceId);
  $: connectorInstanceId = $instance.data?.instance?.instanceId;
  $: connectorName = $instance.data?.instance?.olapConnector;
  $: tableNames = createConnectorServiceOLAPListTables(
    {
      instanceId: connectorInstanceId,
      connector: connectorName,
    },
    {
      query: {
        enabled: !!connectorInstanceId && !!connectorName,
        select: (data) => {
          return (
            data?.tables?.map((table) => table.database + "." + table.name) ||
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
  <NavigationHeader bind:show={showTables} toggleText="external tables">
    External tables
  </NavigationHeader>

  {#if showTables}
    <div class="pb-3" transition:slide={{ duration: LIST_SLIDE_DURATION }}>
      {#if $tableNames?.data}
        {#each $tableNames.data as fullyQualifiedTableName (fullyQualifiedTableName)}
          <div
            animate:flip={{ duration: 200 }}
            out:slide|global={{ duration: LIST_SLIDE_DURATION }}
          >
            <NavigationEntry
              name={fullyQualifiedTableName}
              href={`/external-table/${fullyQualifiedTableName}`}
              open={$page.url.pathname ===
                `/external-table/${fullyQualifiedTableName}`}
              expandable={false}
            >
              <!-- on:command-click={() => queryHandler(tableName)} -->
              <!-- <svelte:fragment slot="tooltip-content">
              <SourceTooltip {tableName} connector="" />
            </svelte:fragment> -->

              <svelte:fragment slot="menu-items" let:toggleMenu>
                <ExternalTableMenuItems
                  {fullyQualifiedTableName}
                  {toggleMenu}
                />
              </svelte:fragment>
            </NavigationEntry>
          </div>
        {/each}
      {/if}
    </div>
  {/if}
{/if}
