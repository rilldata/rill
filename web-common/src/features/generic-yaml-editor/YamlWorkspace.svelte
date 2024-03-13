<script lang="ts">
  import { WorkspaceContainer } from "../../layout/workspace";
  import YamlWorkspaceBody from "./YamlWorkspaceBody.svelte";
  import YamlWorkspaceHeader from "./YamlWorkspaceHeader.svelte";

  export let fileName: string;
</script>

<WorkspaceContainer inspector={false}>
  <YamlWorkspaceHeader slot="header" {fileName} />
  <YamlWorkspaceBody slot="body" {fileName} />
</WorkspaceContainer>
