<script lang="ts">
  import { WorkspaceContainer } from "../../../layout/workspace";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { useIsModelingSupportedForCurrentOlapDriver } from "../../tables/selectors";
  import MetricsEditor from "./editor/MetricsEditor.svelte";
  import MetricsInspector from "./inspector/MetricsInspector.svelte";
  import MetricsWorkspaceHeader from "./MetricsWorkspaceHeader.svelte";

  export let filePath: string;

  $: isModelingSupportedForCurrentOlapDriver =
    useIsModelingSupportedForCurrentOlapDriver($runtime.instanceId);
  $: showInspector = $isModelingSupportedForCurrentOlapDriver.data;
</script>

<WorkspaceContainer inspector={showInspector}>
  <MetricsWorkspaceHeader
    {filePath}
    showInspectorToggle={showInspector}
    slot="header"
  />
  <MetricsEditor {filePath} slot="body" />
  <MetricsInspector {filePath} slot="inspector" />
</WorkspaceContainer>
