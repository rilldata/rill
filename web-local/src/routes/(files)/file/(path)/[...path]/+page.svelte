<script lang="ts">
  import Editor from "@rilldata/web-common/features/models/workspace/Editor.svelte";
  import WorkspaceContainer from "@rilldata/web-common/layout/workspace/WorkspaceContainer.svelte";
  import WorkspaceHeader from "@rilldata/web-common/layout/workspace/WorkspaceHeader.svelte";

  export let data;

  let latest = data.file.blob ?? "";

  $: blob = data.file.blob ?? "";
</script>

<WorkspaceContainer inspector={false}>
  <WorkspaceHeader
    titleInput={data.file.name}
    slot="header"
    showInspectorToggle={false}
  />
  <Editor {blob} slot="body" hasUnsavedChanges={false} bind:latest />
</WorkspaceContainer>
