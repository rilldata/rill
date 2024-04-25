<script lang="ts">
  import Tooltip from "../../components/tooltip/Tooltip.svelte";
  import TooltipContent from "../../components/tooltip/TooltipContent.svelte";
  import { createQueryServiceTableColumns } from "../../runtime-client";
  import { runtime } from "../../runtime-client/runtime-store";

  export let connector: string;
  export let database: string = ""; // The backend interprets an empty string as the default database
  export let databaseSchema: string = ""; // The backend interprets an empty string as the default schema
  export let table: string;

  $: columnsQuery = createQueryServiceTableColumns(
    $runtime?.instanceId,
    table,
    {
      connector,
      database,
      databaseSchema,
    },
  );
</script>

<ul class="schema-list">
  {#if $columnsQuery.isError}
    <div>
      Error loading schema: {$columnsQuery.error?.response.data.message}
    </div>
  {:else if $columnsQuery.data && $columnsQuery.data.profileColumns}
    {#each $columnsQuery.data.profileColumns as column}
      <li>
        <Tooltip distance={4}>
          <span class="font-mono truncate">{column.name}</span>
          <TooltipContent slot="tooltip-content">
            {column.name}
          </TooltipContent>
        </Tooltip>
        <span class="uppercase text-gray-700">{column.type}</span>
      </li>
    {/each}
  {/if}
</ul>

<style lang="postcss">
  .schema-list {
    @apply pl-[30px] pr-4 py-1.5;
    @apply flex flex-col gap-y-0.5;
  }

  .schema-list li {
    @apply flex justify-between gap-x-2;
  }
</style>
