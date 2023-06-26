<script lang="ts">
  import type { Extension } from "@codemirror/state";
  import { EditorState } from "@codemirror/state";
  import { EditorView } from "@codemirror/view";
  import { createEventDispatcher, onMount } from "svelte";
  import { bindEditorEventsToDispatcher } from "./dispatch-events";
  import { basicEditor } from "./languages/base";
  import { yamlEditor } from "./languages/yaml";

  const dispatch = createEventDispatcher();

  export let content: string;
  export let extensions: Extension[] = [];
  export let view: EditorView;

  let container: HTMLElement;
  export let stateFieldUpdaters: ((arg0: EditorView) => void)[] = [];

  onMount(() => {
    view = new EditorView({
      state: EditorState.create({
        doc: content,
        extensions: [
          // any extensions passed as props
          ...extensions,
          // establish a basic editor
          basicEditor(),
          // establish the yaml editor, which currently only has
          // syntax highlighting
          yamlEditor(),
          // this will catch certain events and dispatch them to the parent
          bindEditorEventsToDispatcher(dispatch, stateFieldUpdaters),
        ],
      }),
      parent: container,
    });
  });

  /** Run all the state field updaters once view is ready.
   * We should find a way to remove this code block.
   */
  $: if (view) {
    if (view.updateState !== 2)
      stateFieldUpdaters.forEach((updater) => {
        updater(view);
      });
  }

  /** Listen for changes to the content. If it doesn't match the editor state,
   * update the editor state.
   */
  $: if (view && content !== view?.state?.doc?.toString() && content?.length) {
    view.dispatch({
      changes: {
        from: 0,
        to: view.state.doc.length,
        insert: content,
      },
    });
  }
</script>

<div class="contents" bind:this={container} />
