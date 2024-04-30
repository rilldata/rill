<script lang="ts">
  import YAMLEditor from "@rilldata/web-common/components/editor/YAMLEditor.svelte";
  import { setLineStatuses } from "../../../components/editor/line-status";
  import { mapParseErrorsToLines } from "../../metrics-views/errors";
  import { createEventDispatcher } from "svelte";
  import type { EditorView } from "@codemirror/view";
  import type { V1ParseError } from "@rilldata/web-common/runtime-client";

  const dispatch = createEventDispatcher();

  export let blob: string;
  export let latest: string;
  export let hasUnsavedChanges: boolean;
  export let allErrors: V1ParseError[];
  export let filePath: string;

  let view: EditorView;

  $: latest = blob;

  function handleUpdate(e: CustomEvent<{ content: string }>) {
    latest = e.detail.content;

    // Clear line errors (it's confusing when they're outdated)
    setLineStatuses([], view);
  }

  //  Handle errors
  $: if (view) setLineStatuses(mapParseErrorsToLines(allErrors, blob), view);

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
    <YAMLEditor
      content={latest}
      bind:view
      on:update={handleUpdate}
      key={filePath}
    />
  </div>
</div>
