<script lang="ts">
  import { WorkspaceContainer } from "../../layout/workspace";
  import YamlWorkspaceBody from "./YamlWorkspaceBody.svelte";
  import YamlWorkspaceHeader from "./YamlWorkspaceHeader.svelte";

  export let filePath: string;
</script>

<WorkspaceContainer inspector={false}>
  <YamlWorkspaceHeader {filePath} slot="header" />
  <YamlWorkspaceBody {filePath} slot="body" />
</WorkspaceContainer>
