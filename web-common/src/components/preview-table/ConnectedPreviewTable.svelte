<script lang="ts">
  import PreviewTable from "@rilldata/web-common/components/preview-table/PreviewTable.svelte";
  import ReconcilingSpinner from "@rilldata/web-common/features/entity-management/ReconcilingSpinner.svelte";
  import {
    createQueryServiceTableColumns,
    createQueryServiceTableRows,
  } from "@rilldata/web-common/runtime-client";

  import { runtime } from "../../runtime-client/runtime-store";

  export let objectName: string | undefined;
  export let limit = 150;
  export let loading = false;

  $: profileColumnsQuery =
    objectName === undefined
      ? undefined
      : createQueryServiceTableColumns($runtime?.instanceId, objectName, {});

  $: profileColumns =
    profileColumnsQuery === undefined
      ? undefined
      : $profileColumnsQuery?.data?.profileColumns ?? profileColumns; // Retain old profileColumns

  $: tableQuery =
    objectName === undefined
      ? undefined
      : createQueryServiceTableRows($runtime?.instanceId, objectName, {
          limit,
        });

  $: rows = $tableQuery?.data?.data ?? rows; // Retain old rows
</script>

{#if loading}
  <ReconcilingSpinner />
{:else if rows && profileColumns}
  <PreviewTable {rows} columns={profileColumns} />
{/if}
