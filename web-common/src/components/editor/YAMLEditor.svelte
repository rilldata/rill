<script lang="ts">
  import type { Extension } from "@codemirror/state";
  import { EditorState } from "@codemirror/state";
  import { EditorView } from "@codemirror/view";
  import { createEventDispatcher, onMount } from "svelte";
  import { outsideContentUpdateAnnotation } from "./annotations";
  import { bindEditorEventsToDispatcher } from "./dispatch-events";
  import { basicEditor } from "./languages/base";
  import { yamlEditor } from "./languages/yaml";

  const dispatch = createEventDispatcher();

  export let content: string;
  export let extensions: Extension[] = [];
  export let view: EditorView;

  /** Expose the ability to update the content of the editor. */
  export function optimisticallyUpdateEditorState(newContent: string) {
    if (view)
      view.dispatch({
        changes: {
          from: 0,
          to: view.state.doc.length,
          insert: newContent,
        },
        annotations: outsideContentUpdateAnnotation.of("update-content"),
      });
  }

  let container: HTMLElement;

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
          bindEditorEventsToDispatcher(dispatch),
        ],
      }),
      parent: container,
    });
  });

  // /** Listen for changes to the content. If it doesn't match the editor state,
  //  * update the editor state.
  //  */

  // $: if (
  //   view &&
  //   // content does not match view state.
  //   // debounce the update.
  //   content !== view?.state?.doc?.toString() &&
  //   content?.length
  // ) {
  //   if (!editing)
  //     runtimeUpdateDebouncer(
  //       () => {
  //         view.dispatch({
  //           changes: {
  //             from: 0,
  //             to: view.state.doc.length,
  //             insert: content,
  //           },
  //           annotations: outsideContentUpdateAnnotation.of("runtime-update"),
  //         });
  //       },
  //       // debounce a second only if the doc is not currently empty.
  //       // if it is empty, we should run the update immediately.
  //       view?.state?.doc?.toString()?.length ? 1000 : 0
  //     );
  // } else if (
  //   view &&
  //   // content matches view state.
  //   // let's make sure to clear the debouncer.
  //   content === view?.state?.doc?.toString() &&
  //   content?.length
  // ) {
  //   runtimeUpdateDebouncer.clear();
  // }
</script>

<div class="contents" bind:this={container} />
