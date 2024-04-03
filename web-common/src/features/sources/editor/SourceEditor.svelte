<script lang="ts">
  import type { EditorView } from "@codemirror/view";
  import YAMLEditor from "@rilldata/web-common/components/editor/YAMLEditor.svelte";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import { checkSourceImported } from "@rilldata/web-common/features/sources/source-imported-utils";
  import { setLineStatuses } from "../../../components/editor/line-status";
  import { overlay } from "../../../layout/overlay-store";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { mapParseErrorsToLines } from "../../metrics-views/errors";
  import { saveAndRefresh } from "../saveAndRefresh";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";

  export let filePath: string;
  export let yaml: string;
  export let latest: string;
  export let isSourceUnsaved: boolean;

  let editor: YAMLEditor;
  let view: EditorView;

  $: latest = yaml;
  $: fileArtifact = fileArtifacts.getFileArtifact(filePath);
  $: allErrors = fileArtifact.getAllErrors(queryClient, $runtime.instanceId);

  function handleUpdate(e: CustomEvent<{ content: string }>) {
    latest = e.detail.content;

    // Clear line errors (it's confusing when they're outdated)
    setLineStatuses([], view);
  }

  /**
   * Handle errors.
   */
  $: {
    const lineBasedReconciliationErrors = mapParseErrorsToLines(
      $allErrors,
      yaml,
    );
    if (view) setLineStatuses(lineBasedReconciliationErrors, view);
  }

  async function handleModSave(event: KeyboardEvent) {
    // Check if a Modifier Key + S is pressed
    if (!(event.metaKey || event.ctrlKey) || event.key !== "s") return;

    // Prevent default behaviour
    event.preventDefault();

    // Save the source, if it's unsaved
    if (!isSourceUnsaved) return;
    overlay.set({ title: `Importing ${filePath}` });
    await saveAndRefresh(filePath, latest);
    checkSourceImported(queryClient, filePath);
    overlay.set(null);
  }
</script>

<svelte:window on:keydown={handleModSave} />

<div class="editor flex flex-col border border-gray-200 rounded h-full">
  <div class="grow flex bg-white overflow-y-auto rounded">
    <YAMLEditor
      bind:this={editor}
      bind:view
      content={latest}
      on:update={handleUpdate}
    />
  </div>
</div>
