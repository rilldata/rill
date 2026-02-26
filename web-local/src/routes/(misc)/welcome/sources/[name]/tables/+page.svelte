<script lang="ts">
  import type { LayoutData } from "./$types";
  import ImportTableForm from "@rilldata/web-common/features/sources/import/ImportTableForm.svelte";
  import { ImportTableRunner } from "@rilldata/web-common/features/sources/import/ImportTableRunner.ts";
  import ImportTableStatus from "@rilldata/web-common/features/sources/import/ImportTableStatus.svelte";
  import type { ConnectorTableEntry } from "@rilldata/web-common/features/connectors/explorer/connector-explorer-store.ts";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store.ts";

  export let data: LayoutData;
  $: ({ connectorName, connectorDriver } = data);

  $: ({ instanceId } = $runtime);

  let runner: ImportTableRunner | null = null;
  async function startImport(
    name: string,
    tableEntry: ConnectorTableEntry,
    yaml: string,
  ) {
    runner = new ImportTableRunner(instanceId, name, tableEntry, yaml);
    await runner.run();
  }
</script>

{#if runner}
  <ImportTableStatus {runner} />
{:else}
  <ImportTableForm {connectorName} {connectorDriver} onCreate={startImport} />
{/if}
