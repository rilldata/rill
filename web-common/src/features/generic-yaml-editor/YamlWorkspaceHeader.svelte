<script lang="ts">
  import { WorkspaceHeader } from "../../layout/workspace";

  export let fileName: string;
</script>

<div class="grid items-center" style:grid-template-columns="auto max-content">
  <WorkspaceHeader
    titleInput={fileName}
    editable={false}
    showInspectorToggle={false}
  />
</div>
