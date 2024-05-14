<script lang="ts">
  import type { EditorView } from "@codemirror/view";
  import type { V1ParseError } from "@rilldata/web-common/runtime-client";
  import { createEventDispatcher } from "svelte";
  import { setLineStatuses } from "../../../components/editor/line-status";
  import { mapParseErrorsToLines } from "../../metrics-views/errors";
  import Editor from "../../editor/Editor.svelte";
  import { FileExtensionToEditorExtension } from "../../editor/getExtensionsForFile";

  const dispatch = createEventDispatcher();

  export let blob: string;
  export let localContent: string | null;
  export let hasUnsavedChanges: boolean;
  export let allErrors: V1ParseError[];
  export let filePath: string;

  let editor: EditorView;

  function handleUpdate(e: CustomEvent<{ content: string }>) {
    localContent = e.detail.content;

    // Clear line errors (it's confusing when they're outdated)
    setLineStatuses([], editor);
  }

  //  Handle errors
  $: if (editor)
    setLineStatuses(mapParseErrorsToLines(allErrors, blob), editor);

  function handleModSave(event: KeyboardEvent) {
    // Check if a Modifier Key + S is pressed
    if (!(event.metaKey || event.ctrlKey) || event.key !== "s") return;

    event.preventDefault();

    if (!hasUnsavedChanges) return;
    dispatch("save");
  }
</script>

<svelte:window on:keydown={handleModSave} />

<div class="editor flex flex-col border border-gray-200 rounded h-full">
  <div class="grow flex bg-white overflow-y-auto rounded">
    <Editor
      extensions={FileExtensionToEditorExtension[".yaml"]}
      remoteContent={blob}
      {hasUnsavedChanges}
      bind:localContent
      bind:editor
      autoSave={false}
      disableAutoSave={true}
      on:save={handleUpdate}
      key={filePath}
    />
  </div>
</div>
