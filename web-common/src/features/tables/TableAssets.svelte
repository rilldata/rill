<script lang="ts">
  import { page } from "$app/stores";
  import { flip } from "svelte/animate";
  import { writable } from "svelte/store";
  import { slide } from "svelte/transition";
  import WarningIcon from "../../components/icons/WarningIcon.svelte";
  import Tooltip from "../../components/tooltip/Tooltip.svelte";
  import TooltipContent from "../../components/tooltip/TooltipContent.svelte";
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
  import { useTables } from "./selectors";

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
</script>

{#if hasAssets}
  <div class="h-fit flex flex-col">
    <NavigationHeader bind:show={showTables}>Tables</NavigationHeader>

    {#if showTables}
      <ol transition:slide={{ duration }}>
        {#if $debouncedTables.length > 0}
          {#each $debouncedTables as table (table)}
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
                  {#if table.hasUnsupportedDataTypes}
                    <Tooltip distance={8}>
                      <WarningIcon />
                      <TooltipContent slot="tooltip-content">
                        This table contains unsupported data types.<br />
                        The affected columns will not be available for querying.
                      </TooltipContent>
                    </Tooltip>
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
