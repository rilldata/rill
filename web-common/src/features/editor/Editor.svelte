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
  import { debounce } from "@rilldata/web-common/lib/create-debouncer";
  import { createEventDispatcher, onMount } from "svelte";

  export let content: string;
  export let focusOnMount = false;

  const dispatch = createEventDispatcher();

  let editor: EditorView;
  let editorContainerComponent: HTMLDivElement;
  let latestContent = content;
  const writeDebounce = debounce(() => {
    dispatch("write", {
      content: latestContent,
    });
  }, 500);

  onMount(() => {
    editor = new EditorView({
      state: EditorState.create({
        doc: content,
        extensions: [
          lineNumbers(),
          highlightActiveLineGutter(),
          highlightSpecialChars(),
          history(),
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
            ]),
          ),
          Prec.highest(
            keymap.of([
              {
                key: "Tab",
                run: acceptCompletion,
              },
            ]),
          ),
          keymap.of([indentWithTab]),
          EditorView.updateListener.of((v) => {
            if (v.focusChanged && v.view.hasFocus) {
              dispatch("receive-focus");
            }
            if (v.docChanged) {
              latestContent = v.state.doc.toString();
              writeDebounce();
            }
          }),
        ],
      }),
      parent: editorContainerComponent,
    });
    if (focusOnMount) editor.focus();
  });

  // REACTIVE FUNCTIONS
  function updateEditorContents(newContent: string) {
    if (editor && !editor.hasFocus) {
      let curContent = editor.state.doc.toString();
      if (newContent != curContent) {
        // TODO: should we debounce this?
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
  $: updateEditorContents(content);
</script>

<div class="h-full w-full overflow-x-auto">
  <div class="editor-container h-full w-full overflow-x-auto">
    <div
      bind:this={editorContainerComponent}
      class="w-full overflow-x-auto h-full"
      on:click={() => {
        /** give the editor focus no matter where we click */
        if (!editor.hasFocus) editor.focus();
      }}
      on:keydown={() => {
        /** no op for now */
      }}
      role="textbox"
      tabindex="0"
    />
  </div>
</div>

<style>
  .editor-container {
    padding: 0.5rem;
    background-color: white;
    border-radius: 0.25rem;
    /* min-height: 400px; */
    min-height: 100%;
    display: grid;
    align-items: stretch;
  }
</style>
