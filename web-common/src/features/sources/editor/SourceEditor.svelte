<script lang="ts">
  import type { EditorView } from "@codemirror/view";
  import YAMLEditor from "@rilldata/web-common/components/editor/YAMLEditor.svelte";
  import { fileArtifactsStore } from "@rilldata/web-common/features/entity-management/file-artifacts-store";
  import { checkSourceImported } from "@rilldata/web-common/features/sources/source-imported-utils";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { setLineStatuses } from "../../../components/editor/line-status";
  import { overlay } from "../../../layout/overlay-store";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { mapParseErrorsToLines } from "../../metrics-views/errors";
  import { saveAndRefresh } from "../saveAndRefresh";
  import { useIsSourceUnsaved } from "../selectors";
  import { useSourceStore } from "../sources-store";

  export let filePath: string;
  export let yaml: string;

  let editor: YAMLEditor;
  let view: EditorView;

  const queryClient = useQueryClient();
  const sourceStore = useSourceStore(filePath);

  $: isSourceUnsavedQuery = useIsSourceUnsaved(
    $runtime.instanceId,
    filePath,
    $sourceStore.clientYAML,
  );
  $: isSourceUnsaved = $isSourceUnsavedQuery.data;

  function handleUpdate(e: CustomEvent<{ content: string }>) {
    // Update the client-side store
    sourceStore.set({ clientYAML: e.detail.content });

    // Clear line errors (it's confusing when they're outdated)
    setLineStatuses([], view);
  }

  $: allErrors = fileArtifactsStore.getAllErrorsForFile(
    queryClient,
    $runtime.instanceId,
    filePath,
  );

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
    await saveAndRefresh(filePath, $sourceStore.clientYAML);
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
      content={$sourceStore.clientYAML}
      on:update={handleUpdate}
    />
  </div>
</div>
