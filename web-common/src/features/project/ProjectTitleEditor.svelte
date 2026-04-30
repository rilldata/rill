<script lang="ts">
  import InputWithConfirm from "@rilldata/web-common/components/forms/InputWithConfirm.svelte";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import { useProjectTitle } from "@rilldata/web-common/features/project/selectors";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { get } from "svelte/store";
  import { parseDocument } from "yaml";

  const runtimeClient = useRuntimeClient();

  $: projectTitleQuery = useProjectTitle(runtimeClient);
  $: projectTitle = $projectTitleQuery?.data ?? "Untitled Rill Project";
  $: ({ unsavedFiles } = fileArtifacts);
  $: ({ size: unsavedFileCount } = $unsavedFiles);

  async function submitTitleChange(editedTitle: string) {
    const artifact = fileArtifacts.getFileArtifact("/rill.yaml");
    let content = get(artifact.editorContent);
    if (!content) {
      await artifact.fetchContent();
      content = get(artifact.remoteContent);
      if (!content) return;
    }
    const parsed = parseDocument(content);
    parsed.set("display_name", editedTitle);
    artifact.updateEditorContent(parsed.toString(), true);
    await artifact.saveLocalContent();
  }
</script>

<InputWithConfirm
  size="md"
  bumpDown
  type="Project"
  textClass="font-medium"
  value={projectTitle}
  onConfirm={submitTitleChange}
  showIndicator={unsavedFileCount > 0}
/>
