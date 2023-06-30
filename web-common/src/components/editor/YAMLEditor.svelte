<script lang="ts">
  import type { Extension } from "@codemirror/state";
  import { EditorState } from "@codemirror/state";
  import { EditorView } from "@codemirror/view";
  import { createEventDispatcher, onMount } from "svelte";
  import { bindEditorEventsToDispatcher } from "./dispatch-events";
  import { base } from "./presets/base";
  import { yaml } from "./presets/yaml";

  const dispatch = createEventDispatcher();

  export let content: string;
  export let extensions: Extension[] = [];
  export let view: EditorView = undefined;

  let container: HTMLElement;

  onMount(() => {
    view = new EditorView({
      state: EditorState.create({
        doc: content,
        extensions: [
          // any extensions passed as props
          ...extensions,
          // establish a basic editor
          base(),
          // establish the yaml editor, which currently only has
          // syntax highlighting
          yaml(),
          // this will catch certain events and dispatch them to the parent
          bindEditorEventsToDispatcher(dispatch),
        ],
      }),
      parent: container,
    });
  });
</script>

<div class="contents" bind:this={container} />
