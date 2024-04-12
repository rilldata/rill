<script lang="ts">
  import type { Extension } from "@codemirror/state";
  import { EditorState } from "@codemirror/state";
  import { EditorView } from "@codemirror/view";
  import { createEventDispatcher, onMount } from "svelte";
  import { bindEditorEventsToDispatcher } from "../../components/editor/dispatch-events";
  import { base } from "../../components/editor/presets/base";

  const dispatch = createEventDispatcher();

  export let blob: string;
  export let latest: string;
  export let extensions: Extension[] = [];

  let editor: EditorView;
  let container: HTMLElement;

  $: latest = blob;

  onMount(() => {
    editor = new EditorView({
      state: EditorState.create({
        doc: blob,
        extensions: [
          // any extensions passed as props
          ...extensions,
          // establish a basic editor
          base(),
          // this will catch certain events and dispatch them to the parent
          bindEditorEventsToDispatcher(dispatch),
        ],
      }),
      parent: container,
    });
  });

  function updateEditorContents(newContent: string) {
    if (editor && !editor.hasFocus) {
      // NOTE: when changing files, we still want to update the editor
      let curContent = editor.state.doc.toString();
      console.log("updateEditorContents", newContent);
      if (newContent != curContent) {
        editor.dispatch({
          changes: {
            from: 0,
            to: curContent.length,
            insert: newContent,
          },
        });
      }
    }
  }

  // reactive statements to dynamically update the editor when inputs change
  $: updateEditorContents(latest);
</script>

<div bind:this={container} class="contents" />
