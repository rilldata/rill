<script lang="ts">
  import { EditorState } from "@codemirror/state";
  import { EditorView } from "@codemirror/view";

  import { createEventDispatcher, onMount } from "svelte";

  import { bindEditorEventsToDispatcher } from "./dispatch-events";
  import { basicEditor } from "./languages/base";
  import { yamlEditor } from "./languages/yaml";

  export let content: string;
  export let extensions = [];

  let container: HTMLElement;
  /**
   * @param {string} content
   * @param {string} key
   * @param {string} value
   */
  export let stateFieldUpdaters = [];

  let latestContent = content;

  const dispatch = createEventDispatcher();

  let view: EditorView;

  onMount(() => {
    view = new EditorView({
      state: EditorState.create({
        doc: latestContent,
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

  /** Run all the state field updaters once view is ready */
  $: if (view) {
    /** view.updateState doesn't appear to be in the EditorView type, even though
     * it clearly exists in this view object.
     */
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
