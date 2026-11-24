<script lang="ts">
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import { get } from "svelte/store";
  import { untrack } from "svelte";

  let { fileArtifact }: { fileArtifact: FileArtifact } = $props();

  // Like the code editor, read the content once when mounting.
  const content = untrack(() => get(fileArtifact.editorContent));
</script>

<pre data-testid="workspace">{content}</pre>
