<script lang="ts">
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
  import type { SelectionRange } from "@codemirror/state";
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
  import { Debounce } from "@rilldata/web-common/features/models/utils/Debounce";
  import { createResizeListenerActionFactory } from "@rilldata/web-common/lib/actions/create-resize-listener-factory";
  import {
    useRuntimeServiceGetCatalogEntry,
    useRuntimeServiceListCatalogEntries,
    V1Model,
  } from "@rilldata/web-common/runtime-client";
  import { createEventDispatcher, onMount } from "svelte";

  export let modelName: string;
  export let content: string;
  export let editorHeight = 0;
  export let selections: SelectionRange[] = [];
  export let focusOnMount = false;

  const dispatch = createEventDispatcher();

  const QUERY_UPDATE_DEBOUNCE_TIMEOUT = 0; // disables debouncing
  // const QUERY_SYNC_DEBOUNCE_TIMEOUT = 1000;

  $: getModel = useRuntimeServiceGetCatalogEntry(modelName);
  let model: V1Model;
  $: model = $getModel?.data?.entry?.model;

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
      overflowX: "hidden",
      width: "100%",
      height: "100%",
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
      backgroundColor: "white",
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

  $: sourceCatalogsQuery = useRuntimeServiceListCatalogEntries({
    type: "OBJECT_TYPE_SOURCE",
  });

  let schema: { [table: string]: string[] };

  /** Track embedded sources separately*/
  let embeddedSources = [];
  $: if ($sourceCatalogsQuery?.data?.entries) {
    schema = {};
    embeddedSources = [];
    for (const sourceTable of $sourceCatalogsQuery.data.entries) {
      const sourceIdentifier = sourceTable?.embedded
        ? sourceTable?.source?.properties?.path
        : sourceTable?.name;
      if (sourceTable?.embedded) embeddedSources.push(sourceIdentifier);
      schema[sourceIdentifier] =
        sourceTable.source?.schema?.fields?.map((field) => field.name) ?? [];
    }
  }

  const DuckDBSQL: SQLDialect = SQLDialect.define({
    keywords:
      "select from where group by all having order limit sample unnest with window qualify values filter exclude replace like ilike glob as case when then end in cast left join on not desc asc sum union",
  });

  function makeAutocompleteConfig(
    schema: { [table: string]: string[] },
    embeddedSources: string[]
  ) {
    return autocompletion({
      override: [
        keywordCompletionSource(DuckDBSQL),
        schemaCompletionSource({ schema }),
      ],
      icons: false,
    });
  }

  // UNDERLINES

  const addUnderline = StateEffect.define<{
    from: number;
    to: number;
  }>();
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
          autocompleteCompartment.of(
            makeAutocompleteConfig(schema, embeddedSources)
          ), // a compartment makes the config dynamic
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

  function updateAutocompleteSources(
    schema: { [table: string]: string[] },
    embeddedSources
  ) {
    if (editor) {
      editor.dispatch({
        effects: autocompleteCompartment.reconfigure(
          makeAutocompleteConfig(schema, embeddedSources)
        ),
      });
    }
  }

  // FIXME: resolve type issues incurred when we type selections as SelectionRange[]
  function underlineSelection(selections: any) {
    if (editor) {
      const effects = selections.map(({ from, to }) =>
        addUnderline.of({ from, to })
      );

      if (!editor.state.field(underlineField, false))
        effects.push(StateEffect.appendConfig.of([underlineField]));
      editor.dispatch({ effects });
      return true;
    }
  }

  // reactive statements to dynamically update the editor when inputs change
  $: updateEditorContents(content);
  $: updateAutocompleteSources(schema, embeddedSources);
  $: underlineSelection(selections || []);
</script>

<div class="h-full w-full overflow-x-auto" use:listenToNodeResize>
  <div
    bind:this={editorContainer}
    class="editor-container  h-full w-full overflow-x-auto"
  >
    <div
      class="w-full overflow-x-auto h-full"
      bind:this={editorContainerComponent}
      on:click={() => {
        /** give the editor focus no matter where we click */
        if (!editor.hasFocus) editor.focus();
      }}
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
