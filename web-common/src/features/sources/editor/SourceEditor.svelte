<script lang="ts">
  import type { EditorView } from "@codemirror/view";
  import YAMLEditor from "@rilldata/web-common/components/editor/YAMLEditor.svelte";
  import { getFilePathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { getAllErrorsForFile } from "@rilldata/web-common/features/entity-management/resources-store";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { setLineStatuses } from "../../../components/editor/line-status";
  import { mapReconciliationErrorsToLines } from "../../metrics-views/errors";
  import { useSourceStore } from "../sources-store";

  export let sourceName: string;
  export let yaml: string;
  $: filePath = getFilePathFromNameAndType(sourceName, EntityType.Table);

  const queryClient = useQueryClient();

  let editor: YAMLEditor;
  let view: EditorView;

  const sourceStore = useSourceStore(sourceName);

  function handleUpdate(e: CustomEvent<{ content: string }>) {
    // Update the client-side store
    sourceStore.set({ clientYAML: e.detail.content });

    // Clear line errors (it's confusing when they're outdated)
    setLineStatuses([], view);
  }

  $: allErrors = getAllErrorsForFile(
    queryClient,
    $runtime.instanceId,
    filePath
  );

  /**
   * Handle errors.
   */
  $: {
    const lineBasedReconciliationErrors = mapReconciliationErrorsToLines(
      $allErrors,
      yaml
    );
    if (view) setLineStatuses(lineBasedReconciliationErrors, view);
  }
</script>

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
