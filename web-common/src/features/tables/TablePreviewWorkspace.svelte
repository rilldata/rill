<script lang="ts">
  import { ConnectedPreviewTable } from "../../components/preview-table";
  import WorkspaceContainer from "../../layout/workspace/WorkspaceContainer.svelte";
  import TableWorkspaceHeader from "./TableWorkspaceHeader.svelte";
  import { makeFullyQualifiedTableName } from "./selectors";

  export let connector: string;
  export let database: string = "";
  export let databaseSchema: string;
  export let table: string;

  $: fullyQualifiedTableName = makeFullyQualifiedTableName(
    database,
    databaseSchema,
    table,
  );
</script>

<WorkspaceContainer inspector={false}>
  <TableWorkspaceHeader {fullyQualifiedTableName} slot="header" />
  <ConnectedPreviewTable
    {connector}
    {database}
    {databaseSchema}
    {table}
    loading={false}
    slot="body"
  />
</WorkspaceContainer>
