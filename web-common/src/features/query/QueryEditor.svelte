<script lang="ts">
  import { autocompletion } from "@codemirror/autocomplete";
  import {
    keywordCompletionSource,
    schemaCompletionSource,
    sql,
  } from "@codemirror/lang-sql";
  import { Compartment, EditorState } from "@codemirror/state";
  import { EditorView, keymap } from "@codemirror/view";
  import { base as baseExtensions } from "../../components/editor/presets/base";
  import { DuckDBSQL } from "../../components/editor/presets/duckDBDialect";
  import { createEventDispatcher, onMount } from "svelte";

  export let initialValue = "";

  const dispatch = createEventDispatcher<{
    run: void;
    change: string;
  }>();

  let parent: HTMLElement;
  let editor: EditorView | null = null;

  const autocompleteCompartment = new Compartment();

  function makeAutocompleteConfig() {
    return autocompletion({
      override: [keywordCompletionSource(DuckDBSQL)],
      icons: false,
    });
  }

  // Cmd/Ctrl+Enter to run query
  const runKeymap = keymap.of([
    {
      key: "Mod-Enter",
      run: () => {
        dispatch("run");
        return true;
      },
    },
  ]);

  onMount(() => {
    editor = new EditorView({
      state: EditorState.create({
        doc: initialValue,
        extensions: [
          baseExtensions(),
          sql({ dialect: DuckDBSQL }),
          autocompleteCompartment.of(makeAutocompleteConfig()),
          runKeymap,
          EditorView.updateListener.of((update) => {
            if (update.docChanged) {
              dispatch("change", update.state.doc.toString());
            }
          }),
        ],
      }),
      parent,
    });

    return () => {
      editor?.destroy();
    };
  });

  export function getContent(): string {
    return editor?.state.doc.toString() ?? "";
  }

  export function focus() {
    editor?.focus();
  }
</script>

<div
  bind:this={parent}
  class="size-full overflow-hidden"
  role="textbox"
  aria-label="SQL query editor"
  tabindex="0"
/>

<style lang="postcss">
  :global(.cm-editor) {
    padding-top: 2px;
  }
</style>
