<script lang="ts">
  import { page } from "$app/stores";
  import { flip } from "svelte/animate";
  import { slide } from "svelte/transition";
  import { LIST_SLIDE_DURATION } from "../../layout/config";
  import NavigationEntry from "../../layout/navigation/NavigationEntry.svelte";
  import NavigationHeader from "../../layout/navigation/NavigationHeader.svelte";
  import { createRuntimeServiceGetInstance } from "../../runtime-client";
  import { runtime } from "../../runtime-client/runtime-store";
  import TableMenuItems from "./TableMenuItems.svelte";
  import { useTableNames } from "./selectors";

  $: instance = createRuntimeServiceGetInstance($runtime.instanceId);
  $: connectorInstanceId = $instance.data?.instance?.instanceId;
  $: olapConnector = $instance.data?.instance?.olapConnector;

  $: tableNames = useTableNames(
    $runtime.instanceId,
    connectorInstanceId,
    olapConnector,
  );
  $: hasAssets = $tableNames && $tableNames.length > 0;

  let showTables = true;
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
      {#if $tableNames.length > 0}
        {#each $tableNames as fullyQualifiedTableName (fullyQualifiedTableName)}
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
