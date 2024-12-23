<script lang="ts">
  import Tooltip from "../../../components/tooltip/Tooltip.svelte";
  import TooltipContent from "../../../components/tooltip/TooltipContent.svelte";
  import { createQueryServiceTableColumns } from "../../../runtime-client";
  import { runtime } from "../../../runtime-client/runtime-store";

  export let connector: string;
  export let database: string = ""; // The backend interprets an empty string as the default database
  export let databaseSchema: string = ""; // The backend interprets an empty string as the default schema
  export let table: string;

  $: ({ instanceId } = $runtime);

  $: columnsQuery = createQueryServiceTableColumns(instanceId, table, {
    connector,
    database,
    databaseSchema,
  });
  $: ({ data, error, isError } = $columnsQuery);

  function prettyPrintType(type: string) {
    // If the type starts with "CODE_", remove it
    return type.replace(/^CODE_/, "");
  }
</script>

<ul class="table-schema-list">
  {#if isError}
    <div>
      Error loading schema: {error?.response.data?.message}
    </div>
  {:else if data && data.profileColumns}
    {#each data.profileColumns as column (column)}
      <li class="table-schema-entry {database ? 'pl-[78px]' : 'pl-[60px]'}">
        <Tooltip distance={4}>
          <span class="font-mono truncate">{column.name}</span>
          <TooltipContent slot="tooltip-content">
            {column.name}
          </TooltipContent>
        </Tooltip>
        <span class="uppercase text-gray-700">
          {prettyPrintType(column.type ?? "")}
        </span>
      </li>
    {/each}
  {/if}
</ul>

<style lang="postcss">
  .table-schema-list {
    @apply pr-4 py-1.5; /* padding-left is set dynamically above */
    @apply flex flex-col gap-y-0.5;
  }

  .table-schema-entry {
    @apply flex justify-between gap-x-2;
  }
</style>
