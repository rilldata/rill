<script lang="ts">
  import { page } from "$app/stores";
  import { flip } from "svelte/animate";
  import { slide } from "svelte/transition";
  import { LIST_SLIDE_DURATION as duration } from "../../layout/config";
  import NavigationEntry from "../../layout/navigation/NavigationEntry.svelte";
  import NavigationHeader from "../../layout/navigation/NavigationHeader.svelte";
  import { createRuntimeServiceGetInstance } from "../../runtime-client";
  import { runtime } from "../../runtime-client/runtime-store";
  import TableMenuItems from "./TableMenuItems.svelte";
  import UnsupportedTypesIndicator from "./UnsupportedTypesIndicator.svelte";
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
</script>

{#if hasAssets}
  <div class="h-fit flex flex-col">
    <NavigationHeader bind:show={showTables}>Tables</NavigationHeader>

    {#if showTables}
      <ol transition:slide={{ duration }}>
        {#if $tables.length > 0}
          {#each $tables as table (table)}
            {@const fullyQualifiedTableName = table.database + "." + table.name}
            <li
              animate:flip={{ duration }}
              aria-label={fullyQualifiedTableName}
            >
              <NavigationEntry
                name={fullyQualifiedTableName}
                href={`/table/${fullyQualifiedTableName}`}
                open={$page.url.pathname ===
                  `/table/${fullyQualifiedTableName}`}
              >
                <svelte:fragment slot="icon">
                  {#if connectorInstanceId && olapConnector && table.name && table.hasUnsupportedDataTypes}
                    <UnsupportedTypesIndicator
                      instanceId={connectorInstanceId}
                      connector={olapConnector}
                      tableName={table.name}
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
