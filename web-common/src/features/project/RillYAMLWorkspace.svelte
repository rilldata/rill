<script>
  import { WorkspaceContainer } from "../../layout/workspace";
  import RillYamlWorkspaceBody from "./RillYAMLWorkspaceBody.svelte";
  import RillYamlWorkspaceHeader from "./RillYAMLWorkspaceHeader.svelte";
</script>

<WorkspaceContainer assetID="rill_yaml" inspector={false}>
  <RillYamlWorkspaceHeader slot="header" />
  <RillYamlWorkspaceBody slot="body" />
</WorkspaceContainer>
