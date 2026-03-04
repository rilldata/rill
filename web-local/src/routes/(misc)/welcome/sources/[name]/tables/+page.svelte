<script lang="ts">
  import { goto } from "$app/navigation";
  import type { PageData } from "./$types";
  import ImportTableForm from "@rilldata/web-common/features/add-data/import/ImportTableForm.svelte";
  import type { ConnectorTableEntry } from "@rilldata/web-common/features/connectors/explorer/connector-explorer-store.ts";

  export let data: PageData;
  $: ({ connectorName, connectorDriver } = data);

  function startImport(
    name: string,
    tableEntry: ConnectorTableEntry,
    yaml: string,
  ) {
    const url = new URL(window.location.href);
    url.pathname = `/welcome/sources/${connectorName}/tables/import`;
    url.searchParams.set("name", name);
    url.searchParams.set("database", tableEntry.database);
    url.searchParams.set("schema", tableEntry.schema);
    url.searchParams.set("table", tableEntry.table);
    sessionStorage.setItem("yaml", yaml);
    // ImportTableForm does form submit, it can interfere with the navigation.
    setTimeout(() => {
      void goto(url.toString());
    }, 100);
  }
</script>

{#if connectorDriver}
  <ImportTableForm {connectorName} {connectorDriver} onCreate={startImport} />
{/if}
