<script lang="ts">
  import { ConnectedPreviewTable } from "../../components/preview-table";
  import WorkspaceContainer from "../../layout/workspace/WorkspaceContainer.svelte";
  import TableWorkspaceHeader from "./TableWorkspaceHeader.svelte";

  export let connector: string;
  export let database: string = "";
  export let databaseSchema: string;
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
  <ConnectedPreviewTable
    {connector}
    {database}
    {databaseSchema}
    {table}
    slot="body"
  />
</WorkspaceContainer>
