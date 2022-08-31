<script lang="ts">
  import type { DerivedTableEntity } from "$common/data-modeler-state-service/entity-state-service/DerivedTableEntityService";
  import type { PersistentTableEntity } from "$common/data-modeler-state-service/entity-state-service/PersistentTableEntityService";
  import { Debounce } from "$common/utils/Debounce";
  import type {
    DerivedTableStore,
    PersistentTableStore,
  } from "$lib/application-state-stores/table-stores";
  import {
    acceptCompletion,
    autocompletion,
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
    keywordCompletionSource,
    schemaCompletionSource,
    sql,
    SQLDialect,
  } from "@codemirror/lang-sql";
  import {
    bracketMatching,
    defaultHighlightStyle,
    indentOnInput,
    syntaxHighlighting,
  } from "@codemirror/language";
  import { lintKeymap } from "@codemirror/lint";
  import { highlightSelectionMatches, searchKeymap } from "@codemirror/search";
  import {
    Compartment,
    EditorState,
    Prec,
    StateEffect,
    StateField,
  } from "@codemirror/state";
  import {
    Decoration,
    DecorationSet,
    drawSelection,
    dropCursor,
    EditorView,
    highlightActiveLine,
    highlightActiveLineGutter,
    highlightSpecialChars,
    keymap,
    lineNumbers,
    rectangularSelection,
  } from "@codemirror/view";
  import { createEventDispatcher, getContext, onMount } from "svelte";
  import { createResizeListenerActionFactory } from "./actions/create-resize-listener-factory";

  const dispatch = createEventDispatcher();
  export let content: string;
  export let editorHeight = 0;
  export let selections: any[] = [];

  const QUERY_UPDATE_DEBOUNCE_TIMEOUT = 0; // disables debouncing
  // const QUERY_SYNC_DEBOUNCE_TIMEOUT = 1000;

  const { observedNode, listenToNodeResize } =
    createResizeListenerActionFactory();

  $: editorHeight = $observedNode?.offsetHeight || 0;

  let latestContent = content;
  const debounce = new Debounce();

  let editor: EditorView;
  let editorContainer;
  let editorContainerComponent;

  // DESIGN

  const highlightBackground = "#f3f9ff";

  // TODO: These hardcoded colors ain't good. Try to move this to app.css and use Tailwind
  // colors. Might have to navigated CodeMirror generated classes.
  const rillTheme = EditorView.theme({
    "&.cm-editor": {
      "&.cm-focused": {
        outline: "none",
      },
    },
    "&.cm-focused .cm-selectionBackground, .cm-selectionBackground, .cm-content ::selection":
      { backgroundColor: "rgb(65 99 255 / 25%)" },
    ".cm-selectionMatch": { backgroundColor: "rgb(189 233 255)" },
    ".cm-activeLine": { backgroundColor: highlightBackground },
    ".cm-activeLineGutter": {
      backgroundColor: highlightBackground,
    },
    ".cm-lineNumbers .cm-gutterElement": {
      paddingLeft: "5px",
      paddingRight: "10px",
      minWidth: "32px",
    },
    ".cm-breakpoint-gutter .cm-gutterElement": {
      color: "red",
      paddingLeft: "24px",
      paddingRight: "24px",
      cursor: "default",
    },
    ".cm-tooltip": {
      border: "none",
      borderRadius: "0.25rem",
      backgroundColor: "rgb(243 249 255)",
      color: "black",
    },
    ".cm-tooltip-autocomplete": {
      "& > ul > li[aria-selected]": {
        border: "none",
        borderRadius: "0.25rem",
        backgroundColor: "rgb(15 119 204 / .25)",
        color: "black",
      },
    },
    ".cm-completionLabel": {
      fontSize: "13px",
      fontFamily: "MD IO",
    },
    ".cm-completionMatchedText": {
      textDecoration: "none",
      color: "rgb(15 119 204)",
    },
    ".cm-underline": {
      backgroundColor: "rgb(254 240 138)",
    },
  });

  // AUTOCOMPLETE

  let autocompleteCompartment = new Compartment();

  const persistentTableStore = getContext(
    "rill:app:persistent-table-store"
  ) as PersistentTableStore;
  const derivedTableStore = getContext(
    "rill:app:derived-table-store"
  ) as DerivedTableStore;

  let schema: { [table: string]: string[] };
  $: {
    schema = $persistentTableStore.entities.reduce(
      (acc, persistentTable: PersistentTableEntity) => {
        const derivedTable: DerivedTableEntity =
          $derivedTableStore.entities.find(
            (derivedTable) => persistentTable.id === derivedTable.id
          );
        // defensive check since persistentTableStore updates incrementally and can
        // have transition states where tables are defined but none of their attributes are
        if (derivedTable?.profile) {
          const columnNames = derivedTable?.profile.map((col) => col.name);
          acc[persistentTable.tableName] = columnNames;
        }
        return acc;
      },
      {}
    );
  }

  const DuckDBSQL: SQLDialect = SQLDialect.define({
    keywords:
      "select from where group by all having order limit sample unnest with window qualify values filter exclude replace like ilike glob",
  });

  function makeAutocompleteConfig(schema: { [table: string]: string[] }) {
    return autocompletion({
      override: [
        keywordCompletionSource(DuckDBSQL),
        schemaCompletionSource({ schema }),
      ],
      icons: false,
    });
  }

  // UNDERLINES

  const addUnderline = StateEffect.define<{ from: number; to: number }>();
  const underlineMark = Decoration.mark({ class: "cm-underline" });
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

  onMount(() => {
    editor = new EditorView({
      state: EditorState.create({
        doc: latestContent,
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
          autocompleteCompartment.of(makeAutocompleteConfig(schema)), // a compartment makes the config dynamic
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
          sql({ dialect: DuckDBSQL }),
          keymap.of([indentWithTab]),
          rillTheme,
          EditorView.updateListener.of((v) => {
            if (v.focusChanged && v.view.hasFocus) {
              dispatch("receive-focus");
            }
            if (v.docChanged) {
              latestContent = v.state.doc.toString();
              debounce.debounce(
                "write",
                () => {
                  dispatch("write", {
                    content: latestContent,
                  });
                },
                QUERY_UPDATE_DEBOUNCE_TIMEOUT
              );
            }
          }),
        ],
      }),
      parent: editorContainerComponent,
    });
  });

  // REACTIVE FUNCTIONS

  // function updateEditorContents(newContent: string) {
  //   if (editor) {
  //     let curContent = editor.state.doc.toString();
  //     if (newContent != curContent) {
  //       latestContent = newContent;
  //       debounce.debounce(
  //         "update",
  //         () => {
  //           editor.dispatch({
  //             changes: {
  //               from: 0,
  //               to: latestContent.length,
  //               insert: latestContent,
  //             },
  //           });
  //         },
  //         QUERY_SYNC_DEBOUNCE_TIMEOUT
  //       );
  //     }
  //   }
  // }

  function updateAutocompleteSources(schema: { [table: string]: string[] }) {
    if (editor) {
      editor.dispatch({
        effects: autocompleteCompartment.reconfigure(
          makeAutocompleteConfig(schema)
        ),
      });
    }
  }

  function underlineSelection(selections: any) {
    if (editor) {
      const effects = selections
        .map(({ start, end }) => ({ from: start, to: end }))
        .map(({ from, to }) => addUnderline.of({ from, to }));

      if (!editor.state.field(underlineField, false))
        effects.push(StateEffect.appendConfig.of([underlineField]));
      editor.dispatch({ effects });
      return true;
    }
  }

  // reactive statements to dynamically update the editor when inputs change
  // $: updateEditorContents(content);
  $: updateAutocompleteSources(schema);
  $: underlineSelection(selections || []);
</script>

<div use:listenToNodeResize class="h-full">
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
