<script lang="ts">
  import { autocompletion } from "@codemirror/autocomplete";
  import { keywordCompletionSource, sql } from "@codemirror/lang-sql";
  import { Compartment, EditorState } from "@codemirror/state";
  import { EditorView, keymap } from "@codemirror/view";
  import { base as baseExtensions } from "../../components/editor/presets/base";
  import { DuckDBSQL } from "../../components/editor/presets/duckDBDialect";
  import { createEventDispatcher, onMount } from "svelte";

  export let initialValue = "";

  const dispatch = createEventDispatcher<{
    run: { selectedText?: string };
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

  // Cmd/Ctrl+Enter to run query (selected text or full content)
  const runKeymap = keymap.of([
    {
      key: "Mod-Enter",
      run: (view) => {
        const sel = view.state.selection.main;
        const hasSelection = sel.from !== sel.to;
        const selectedText = hasSelection
          ? view.state.sliceDoc(sel.from, sel.to)
          : undefined;
        dispatch("run", { selectedText });
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

  export function getSelectedText(): string | undefined {
    if (!editor) return undefined;
    const sel = editor.state.selection.main;
    if (sel.from === sel.to) return undefined;
    return editor.state.sliceDoc(sel.from, sel.to);
  }

  export function setContent(text: string) {
    if (!editor) return;
    editor.dispatch({
      changes: { from: 0, to: editor.state.doc.length, insert: text },
    });
  }

  export function insertAtCursor(text: string) {
    if (!editor) return;
    const pos = editor.state.selection.main.head;
    editor.dispatch({
      changes: { from: pos, insert: text },
      selection: { anchor: pos + text.length },
    });
    editor.focus();
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
  div :global(.cm-editor) {
    padding-top: 2px;
  }
</style>
