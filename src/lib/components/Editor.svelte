<script lang="ts">
  import {onMount, createEventDispatcher, getContext} from "svelte";
  import {
    keymap,
    highlightSpecialChars,
    drawSelection,
    highlightActiveLine,
    dropCursor,
    EditorView,
    Decoration,
    DecorationSet,
  } from "@codemirror/view";
  import {
    EditorState,
    StateEffect,
    StateField,
    Prec,
    Text
  } from "@codemirror/state";
  import { history, historyKeymap } from "@codemirror/history";
  import { foldGutter, foldKeymap } from "@codemirror/fold";
  import { indentOnInput } from "@codemirror/language";
  import { lineNumbers, highlightActiveLineGutter } from "@codemirror/gutter";
  import {
    defaultKeymap,
    insertNewline,
    indentWithTab,
  } from "@codemirror/commands";
  import { bracketMatching } from "@codemirror/matchbrackets";
  import {
    closeBrackets,
    closeBracketsKeymap,
  } from "@codemirror/closebrackets";
  import { searchKeymap, highlightSelectionMatches } from "@codemirror/search";
  import {
    acceptCompletion,
    autocompletion,
    completionKeymap,
  } from "@codemirror/autocomplete";
  import { commentKeymap } from "@codemirror/comment";
  import { rectangularSelection } from "@codemirror/rectangular-selection";
  import { defaultHighlightStyle } from "@codemirror/highlight";
  import { lintKeymap } from "@codemirror/lint";
  import { sql } from "@codemirror/lang-sql";
  import {PersistentModelStore} from "$lib/application-state-stores/model-stores";
  import {ApplicationStore} from "$lib/application-state-stores/application-store";
  import {
    PersistentModelEntity
  } from "$common/data-modeler-state-service/entity-state-service/PersistentModelEntityService";

  const store = getContext("rill:app:store") as ApplicationStore;
  const persistentModelStore = getContext(
    "rill:app:persistent-model-store"
  ) as PersistentModelStore;
  $: activeEntityID = $store?.activeEntity?.id;
  let currentModel: PersistentModelEntity;
  $: currentModel =
    activeEntityID && $persistentModelStore?.entities
      ? $persistentModelStore.entities.find((q) => q.id === activeEntityID)
      : undefined;

  const dispatch = createEventDispatcher();
  export let content;
  export let editorHeight = 0;
  export let selections = [];

  let componentContainer;

  $: editorHeight = componentContainer?.offsetHeight || 0;

  let oldContent = content;

  let editor;
  let editorContainer;
  let editorContainerComponent;

  const addUnderline = StateEffect.define<{ from: number; to: number }>();

  const underlineField = StateField.define<DecorationSet>({
    create() {
      return Decoration.none;
    },
    update(underlines, tr) {
      underlines = underlines.map(tr.changes);
      underlines = underlines.update({
        filter: () => false,
      });

      for (let e of tr.effects)
        if (e.is(addUnderline)) {
          underlines = underlines.update({
            add: [underlineMark.range(e.value.from, e.value.to)],
          });
        }
      return underlines;
    },
    provide: (f) => EditorView.decorations.from(f),
  });

  const underlineMark = Decoration.mark({ class: "cm-underline" });

  const underlineTheme = EditorView.baseTheme({
    ".cm-underline": {
      backgroundColor: "rgb(254 240 138)",
    },
  });

  const rillTheme = EditorView.theme({
    "&.cm-editor": {
      "&.cm-focused": {
        outline: "none",
      },
    },
    "&.cm-focused .cm-selectionBackground, .cm-selectionBackground, .cm-content ::selection":
      { backgroundColor: "rgb(65 99 255 / 25%)" },
    ".cm-selectionMatch": { backgroundColor: "rgb(189 233 255)" },
    ".cm-breakpoint-gutter .cm-gutterElement": {
      color: "red",
      paddingLeft: "24px",
      paddingRight: "24px",
      cursor: "default",
    },
  });

  function underlineSelection(view: EditorView, selections) {
    const effects = selections
      .map(({ start, end }) => ({ from: start, to: end }))
      .map(({ from, to }) => addUnderline.of({ from, to }));

    if (!view.state.field(underlineField, false))
      effects.push(
        StateEffect.appendConfig.of([underlineField, underlineTheme])
      );
    view.dispatch({ effects });
    return true;
  }

  $: if (editor) {
    underlineSelection(editor, selections || []);
  }

  $: if (currentModel && currentModel.query && editor) {
    const currentEditorValue = editor.state.doc.toString();
    if (currentEditorValue !== currentModel.query) {
      // send the changes to the editor if the query is different in the model
      editor.dispatch({
        changes: {
          from: 0,
          to: editor.state.doc.length,
          insert: currentModel.query
        }
      });
    }
  }

  let cursorLocation = 0;

  onMount(() => {
    editor = new EditorView({
      state: EditorState.create({
        doc: oldContent,
        extensions: [
          lineNumbers(),
          highlightActiveLineGutter(),
          highlightSpecialChars(),
          history(),
          foldGutter(),
          drawSelection(),
          dropCursor(),
          EditorState.allowMultipleSelections.of(true),
          indentOnInput(),
          defaultHighlightStyle.fallback,
          bracketMatching(),
          closeBrackets(),
          autocompletion(),
          rectangularSelection(),
          highlightActiveLine(),
          highlightSelectionMatches(),
          keymap.of([
            ...closeBracketsKeymap,
            ...defaultKeymap,
            ...searchKeymap,
            ...historyKeymap,
            ...foldKeymap,
            ...commentKeymap,
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
          sql(),
          keymap.of([indentWithTab]),
          rillTheme,
          EditorView.updateListener.of((v) => {
            const candidateLocation = v.state.selection.ranges[0].head;
            if (candidateLocation !== cursorLocation) {
              cursorLocation = candidateLocation;
              dispatch("cursor-location", {
                location: cursorLocation,
                content: v.state.doc.toString(),
              });
            }
            if (v.focusChanged) {
              if (v.view.hasFocus) {
                dispatch("receive-focus");
              } else {
                dispatch("release-focus");
              }
            }
            if (v.docChanged) {
              dispatch("write", {
                content: v.state.doc.toString(),
              });
            }
          }),
        ],
      }),
      parent: editorContainerComponent,
    });
    const obs = new ResizeObserver(() => {
      editorHeight = componentContainer?.offsetHeight;
    });
    obs.observe(componentContainer);
  });
</script>

<div bind:this={componentContainer} class="h-full">
  <div class="editor-container border h-full" bind:this={editorContainer}>
    <div bind:this={editorContainerComponent} />
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
