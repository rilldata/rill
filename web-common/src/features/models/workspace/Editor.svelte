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
  import { useQueryClient } from "@tanstack/svelte-query";
  import { createEventDispatcher, onMount } from "svelte";
  import { DuckDBSQL } from "../../../components/editor/presets/duckDBDialect";
  import { editorTheme } from "../../../components/editor/theme";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { useAllSourceColumns } from "../../sources/selectors";
  import { useAllModelColumns } from "../selectors";

  export let content: string;
  export let editorHeight = 0;
  export let selections: SelectionRange[] = [];
  export let focusOnMount = false;

  const queryClient = useQueryClient();
  const dispatch = createEventDispatcher();

  let latestContent = content;

  let editor: EditorView;
  let editorContainer;
  let editorContainerComponent;

  // AUTOCOMPLETE

  let autocompleteCompartment = new Compartment();

  const schema: { [table: string]: string[] } = {};

  // Autocomplete: source tables
  $: allSourceColumns = useAllSourceColumns(queryClient, $runtime?.instanceId);
  $: if ($allSourceColumns?.length) {
    for (const sourceTable of $allSourceColumns) {
      const sourceIdentifier = sourceTable?.tableName;
      schema[sourceIdentifier] = sourceTable.profileColumns
        ?.filter((c) => c.name !== undefined)
        // CAST SAFETY: already filtered out undefined values
        .map((c) => c.name as string);
    }
  }

  //Auto complete: model tables
  $: allModelColumns = useAllModelColumns(queryClient, $runtime?.instanceId);
  $: if ($allModelColumns?.length) {
    for (const modelTable of $allModelColumns) {
      const modelIdentifier = modelTable?.tableName;
      schema[modelIdentifier] = modelTable.profileColumns
        ?.filter((c) => c.name !== undefined)
        // CAST SAFETY: already filtered out undefined values
        ?.map((c) => c.name as string);
    }
  }

  function getTableNameFromFromClause(
    sql: string,
    schema: { [table: string]: string[] },
  ): string | undefined {
    if (!sql || !schema) return undefined;

    const fromMatch = sql.toUpperCase().match(/\bFROM\b\s+(\w+)/);
    const tableName = fromMatch ? fromMatch[1] : undefined;

    // Get the tableName from the schema map, so we can use the correct case
    for (const schemaTableName of Object.keys(schema)) {
      if (schemaTableName.toUpperCase() === tableName) {
        return schemaTableName;
      }
    }

    return undefined;
  }

  function makeAutocompleteConfig(
    schema: { [table: string]: string[] },
    defaultTable?: string,
  ) {
    return autocompletion({
      override: [
        keywordCompletionSource(DuckDBSQL),
        schemaCompletionSource({ schema, defaultTable }),
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
          editorTheme(),
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
            makeAutocompleteConfig(schema, defaultTable),
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
          sql({ dialect: DuckDBSQL }),
          keymap.of([indentWithTab]),
          EditorView.updateListener.of((v) => {
            if (v.focusChanged && v.view.hasFocus) {
              dispatch("receive-focus");
            }
            if (v.docChanged) {
              latestContent = v.state.doc.toString();

              dispatch("write", {
                content: latestContent,
              });
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
    defaultTable?: string,
  ) {
    if (editor) {
      editor.dispatch({
        effects: autocompleteCompartment.reconfigure(
          makeAutocompleteConfig(schema, defaultTable),
        ),
      });
    }
  }

  // FIXME: resolve type issues incurred when we type selections as SelectionRange[]
  function underlineSelection(selections: any) {
    if (editor) {
      const effects = selections.map(({ from, to }) =>
        addUnderline.of({ from, to }),
      );

      if (!editor.state.field(underlineField, false))
        effects.push(StateEffect.appendConfig.of([underlineField]));
      editor.dispatch({ effects });
      return true;
    }
  }

  // reactive statements to dynamically update the editor when inputs change
  $: updateEditorContents(content);
  $: defaultTable = getTableNameFromFromClause(content, schema);
  $: updateAutocompleteSources(schema, defaultTable);
  $: underlineSelection(selections || []);
</script>

<div class="h-full w-full overflow-x-auto" bind:clientHeight={editorHeight}>
  <div
    bind:this={editorContainer}
    class="editor-container h-full w-full overflow-x-auto"
  >
    <div
      role="textbox"
      tabindex="0"
      bind:this={editorContainerComponent}
      class="w-full overflow-x-auto h-full"
      on:click={() => {
        /** give the editor focus no matter where we click */
        if (!editor.hasFocus) editor.focus();
      }}
      on:keydown={() => {
        /** no op for now */
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
