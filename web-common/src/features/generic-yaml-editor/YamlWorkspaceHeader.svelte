<script lang="ts">
  import { WorkspaceHeader } from "../../layout/workspace";

  export let filePath: string;
</script>

<div class="grid items-center" style:grid-template-columns="auto max-content">
  <WorkspaceHeader
    editable={false}
    showInspectorToggle={false}
    titleInput={filePath}
  />
</div>
