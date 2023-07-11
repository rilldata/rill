<script lang="ts">
  import { WorkspaceContainer } from "../../../layout/workspace";
  import MetricsWorkspaceHeader from "./MetricsWorkspaceHeader.svelte";
  import MetricsEditor from "./editor/MetricsEditor.svelte";
  import MetricsInspector from "./inspector/MetricsInspector.svelte";

  export let metricsDefName: string;
</script>

<WorkspaceContainer inspector={true} assetID={`${metricsDefName}-config`}>
  <MetricsWorkspaceHeader slot="header" {metricsDefName} />
  <MetricsEditor slot="body" {metricsDefName} />
  <MetricsInspector slot="inspector" {metricsDefName} />
</WorkspaceContainer>
