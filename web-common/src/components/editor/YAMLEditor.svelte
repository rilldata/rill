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
    rectangularSelection,
  } from "@codemirror/view";
  import { editorTheme } from "./theme";

  import { createEventDispatcher, onMount } from "svelte";

  import * as yamlMode from "@codemirror/legacy-modes/mode/yaml";

  export let content;
  export let plugins = [];
  /**
   * @param {string} content
   * @param {string} key
   * @param {string} value
   */
  export let stateFieldUpdaters = [];

  let latestContent = content;

  const yaml = StreamLanguage.define(yamlMode.yaml);

  let container: HTMLDivElement;

  const dispatch = createEventDispatcher();

  let view: EditorView;

  export let hasFocus = false;

  onMount(() => {
    view = new EditorView({
      state: EditorState.create({
        doc: latestContent,
        extensions: [
          ...plugins,

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
          editorTheme(),
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
            const state = v.state;
            hasFocus = v.view.hasFocus;
            const cursor = state.selection.main.head;
            const line = state.doc.lineAt(cursor);
            // dispatch current cursor location
            dispatch("cursor", {
              line: line.number,
              column: cursor - line.from,
            });
            if (v.focusChanged && v.view.hasFocus) {
              dispatch("receive-focus");
            }
            if (v.docChanged) {
              dispatch("update", { content: state.doc.toString() });
              stateFieldUpdaters.forEach((updater) => {
                updater(view);
              });
            }
          }),
        ],
      }),
      parent: container,
    });
  });

  /** Run all the state field updaters once view is ready */
  $: if (view) {
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
