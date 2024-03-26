<script lang="ts">
  import { page } from "$app/stores";
  import { flip } from "svelte/animate";
  import { writable } from "svelte/store";
  import { slide } from "svelte/transition";
  import { LIST_SLIDE_DURATION as duration } from "../../layout/config";
  import NavigationEntry from "../../layout/navigation/NavigationEntry.svelte";
  import NavigationHeader from "../../layout/navigation/NavigationHeader.svelte";
  import { debounce } from "../../lib/create-debouncer";
  import { createRuntimeServiceGetInstance } from "../../runtime-client";
  import { runtime } from "../../runtime-client/runtime-store";
  import TableMenuItems from "./TableMenuItems.svelte";
  import { useTableNames } from "./selectors";

  let showTables = true;

  // Debounce table names to prevent flickering
  const debouncedTableNames = writable<string[]>([]);
  const setDebouncedTableNames = debounce(
    (tableNames: string[]) => debouncedTableNames.set(tableNames),
    200,
  );

  $: if ($tableNames) {
    setDebouncedTableNames($tableNames);
  }

  $: hasAssets = $debouncedTableNames.length > 0;

  $: instance = createRuntimeServiceGetInstance($runtime.instanceId);
  $: connectorInstanceId = $instance.data?.instance?.instanceId;
  $: olapConnector = $instance.data?.instance?.olapConnector;

  $: tableNames = useTableNames(
    $runtime.instanceId,
    connectorInstanceId,
    olapConnector,
  );
</script>

{#if hasAssets}
  <div class="h-fit flex flex-col">
    <NavigationHeader bind:show={showTables}>Tables</NavigationHeader>

    {#if showTables}
      <ol transition:slide={{ duration }}>
        {#if $debouncedTableNames.length > 0}
          {#each $debouncedTableNames as fullyQualifiedTableName (fullyQualifiedTableName)}
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
                <TableMenuItems slot="menu-items" {fullyQualifiedTableName} />
              </NavigationEntry>
            </li>
          {/each}
        {/if}
      </ol>
    {/if}
  </div>
{/if}
