<script lang="ts">
  import type { EditorView } from "@codemirror/view";
  import YAMLEditor from "@rilldata/web-common/components/editor/YAMLEditor.svelte";
  import type { V1ParseError } from "@rilldata/web-common/runtime-client";
  import { setLineStatuses } from "../../../components/editor/line-status";
  import { mapParseErrorsToLines } from "../../metrics-views/errors";
  import { FileArtifact } from "../../entity-management/file-artifacts";

  export let blob: string;

  export let allErrors: V1ParseError[];
  export let filePath: string;
  export let fileArtifact: FileArtifact;

  let view: EditorView;

  $: ({ saveLocalContent } = fileArtifact);

  function handleUpdate(e: CustomEvent<{ content: string }>) {
    // Clear line errors (it's confusing when they're outdated)
    setLineStatuses([], view);
  }

  //  Handle errors
  $: if (view) setLineStatuses(mapParseErrorsToLines(allErrors, blob), view);

  function handleModSave(event: KeyboardEvent) {
    // Check if a Modifier Key + S is pressed
    if (!(event.metaKey || event.ctrlKey) || event.key !== "s") return;

    event.preventDefault();

    saveLocalContent().catch(console.error);
  }
</script>

<svelte:window on:keydown={handleModSave} />

<div class="editor flex flex-col border border-gray-200 rounded h-full">
  <div class="grow flex bg-white overflow-y-auto rounded">
    <YAMLEditor content={blob} bind:view {fileArtifact} key={filePath} />
  </div>
</div>
