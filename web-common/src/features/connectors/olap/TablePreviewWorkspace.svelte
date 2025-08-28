<script lang="ts">
  import { ConnectedPreviewTable } from "../../../components/preview-table";
  import WorkspaceContainer from "../../../layout/workspace/WorkspaceContainer.svelte";
  import TableWorkspaceHeader from "../explorer/TableWorkspaceHeader.svelte";

  export let connector: string;
  export let database: string = "";
  export let databaseSchema: string = "";
  export let table: string;
</script>

<WorkspaceContainer inspector={false}>
  <TableWorkspaceHeader
    {connector}
    {database}
    {databaseSchema}
    {table}
    slot="header"
  />
  <div class="border size-full rounded-[2px]" slot="body">
    <ConnectedPreviewTable {connector} {database} {databaseSchema} {table} />
  </div>
</WorkspaceContainer>
