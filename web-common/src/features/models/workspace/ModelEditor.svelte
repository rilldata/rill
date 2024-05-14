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
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Label from "@rilldata/web-common/components/forms/Label.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import Check from "@rilldata/web-common/components/icons/Check.svelte";
  import UndoIcon from "@rilldata/web-common/components/icons/UndoIcon.svelte";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { createEventDispatcher, onMount } from "svelte";
  import { DuckDBSQL } from "../../../components/editor/presets/duckDBDialect";
  import { editorTheme } from "../../../components/editor/theme";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { useAllSourceColumns } from "../../sources/selectors";
  import { useAllModelColumns } from "../selectors";

  const dispatch = createEventDispatcher();
  const schema: { [table: string]: string[] } = {};

  export let blob: string;
  export let localContent: string | null;
  export let selections: SelectionRange[] = [];
  export let autoSave = true;
  export let hasUnsavedChanges: boolean;
  export let key: string;

  let editor: EditorView;
  let editorContainerComponent: HTMLDivElement;
  let autocompleteCompartment = new Compartment();

  $: if (key) {
    // When the key changes, unfocus the Editor so that the update is dispatched
    editor?.contentDOM.blur();
  }

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

  // reactive statements to dynamically update the editor when inputs change
  $: updateEditorContents(blob);
  $: defaultTable = getTableNameFromFromClause(blob, schema);
  $: updateAutocompleteSources(schema, defaultTable);
  $: underlineSelection(selections || []);

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
        doc: blob,
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
              localContent = v.state.doc.toString();

              if (autoSave) saveContent();
            }
          }),
        ],
      }),
      parent: editorContainerComponent,
    });
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

  function saveContent() {
    dispatch("save");
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === "s" && (e.ctrlKey || e.metaKey)) {
      e.preventDefault();
      saveContent();
    }
  }

  function revertContent() {
    dispatch("revert");
  }
</script>

<svelte:window on:keydown={handleKeydown} />

<section>
  <div class="editor-container">
    <div
      bind:this={editorContainerComponent}
      class="size-full"
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

  <footer>
    <div class="flex gap-x-3">
      {#if !autoSave}
        <Button disabled={!hasUnsavedChanges} on:click={saveContent}>
          <Check size="14px" />
          Save
        </Button>

        <Button
          type="text"
          disabled={!hasUnsavedChanges}
          on:click={revertContent}
        >
          <UndoIcon size="14px" />
          Revert changes
        </Button>
      {/if}
    </div>
    <div class="flex gap-x-1 items-center h-full bg-white rounded-full">
      <Switch
        bind:checked={autoSave}
        id="auto-save"
        small
        on:click={() => {
          if (!autoSave) saveContent();
        }}
      />
      <Label class="font-normal text-xs" for="auto-save">Auto-save</Label>
    </div>
  </footer>
</section>

<style lang="postcss">
  .editor-container {
    @apply size-full overflow-auto p-2 pb-0 flex flex-col;
  }

  footer {
    @apply justify-between items-center flex flex-none;
    @apply h-10 p-2 w-full rounded-b-sm border-t bg-white;
  }

  section {
    @apply size-full flex-col rounded-sm bg-white flex overflow-hidden relative;
  }
</style>
