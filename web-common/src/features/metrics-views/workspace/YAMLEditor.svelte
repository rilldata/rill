<script lang="ts">
  import {
    acceptCompletion,
    closeBrackets,
    closeBracketsKeymap,
    completionKeymap,
  } from "@codemirror/autocomplete";
  import {
    defaultKeymap,
    history,
    historyKeymap,
    indentWithTab,
    insertNewline,
  } from "@codemirror/commands";
  import {
    StreamLanguage,
    bracketMatching,
    defaultHighlightStyle,
    indentOnInput,
    syntaxHighlighting,
  } from "@codemirror/language";
  import { lintKeymap } from "@codemirror/lint";
  import { highlightSelectionMatches, searchKeymap } from "@codemirror/search";
  import { EditorState, Prec } from "@codemirror/state";
  import {
    EditorView,
    drawSelection,
    dropCursor,
    highlightActiveLine,
    highlightActiveLineGutter,
    highlightSpecialChars,
    keymap,
    lineNumbers,
    rectangularSelection,
  } from "@codemirror/view";

  import { createEventDispatcher, onMount } from "svelte";

  import * as yamlMode from "@codemirror/legacy-modes/mode/yaml";

  export let content;
  export let plugins = [];
  export let updaters = [];

  let latestContent = content;

  const yaml = StreamLanguage.define(yamlMode.yaml);

  let container: HTMLDivElement;

  const dispatch = createEventDispatcher();

  let editor: EditorView;

  onMount(() => {
    editor = new EditorView({
      state: EditorState.create({
        doc: latestContent,
        extensions: [
          lineNumbers(),
          highlightActiveLineGutter(),
          highlightSpecialChars(),
          history(),
          yaml,

          drawSelection(),
          dropCursor(),
          EditorState.allowMultipleSelections.of(true),
          indentOnInput(),
          syntaxHighlighting(defaultHighlightStyle, { fallback: true }),
          bracketMatching(),
          closeBrackets(),
          rectangularSelection(),
          highlightActiveLine(),
          highlightSelectionMatches(),
          ...plugins,
          keymap.of([
            ...closeBracketsKeymap,
            ...defaultKeymap,
            ...searchKeymap,
            ...historyKeymap,
            ...completionKeymap,
            ...lintKeymap,
            indentWithTab,
          ]),
          Prec.high(
            keymap.of([
              {
                key: "Enter",
                run: insertNewline,
              },
            ])
          ),
          Prec.highest(
            keymap.of([
              {
                key: "Tab",
                run: acceptCompletion,
              },
            ])
          ),
          keymap.of([indentWithTab]),
          EditorView.updateListener.of((v) => {
            if (v.focusChanged && v.view.hasFocus) {
              dispatch("receive-focus");
            }
            if (v.docChanged) {
              dispatch("update", { content: v.state.doc.toString() });
            }
          }),
        ],
      }),
      parent: container,
    });
  });

  $: if (editor) {
    updaters.forEach((updater) => {
      updater(editor);
    });
  }

  /** Listen for changes to the content. If it doesn't match the editor state,
   * update the editor state.
   */
  $: if (
    editor &&
    content !== editor?.state?.doc?.toString() &&
    content?.length
  ) {
    editor.dispatch({
      changes: {
        from: 0,
        to: editor.state.doc.length,
        insert: content,
      },
    });
  }
</script>

<div bind:this={container} />
