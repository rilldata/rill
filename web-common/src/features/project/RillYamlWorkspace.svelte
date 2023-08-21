<script>
  import { WorkspaceContainer } from "../../layout/workspace";
  import RillYamlWorkspaceBody from "./RillYamlWorkspaceBody.svelte";
  import RillYamlWorkspaceHeader from "./RillYamlWorkspaceHeader.svelte";
</script>

<WorkspaceContainer assetID="rill_yaml" inspector={false}>
  <RillYamlWorkspaceHeader slot="header" />
  <RillYamlWorkspaceBody slot="body" />
</WorkspaceContainer>
