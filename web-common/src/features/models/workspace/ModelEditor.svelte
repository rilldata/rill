<script lang="ts">
  import { autocompletion } from "@codemirror/autocomplete";
  import {
    keywordCompletionSource,
    schemaCompletionSource,
    sql,
  } from "@codemirror/lang-sql";
  import type { SelectionRange } from "@codemirror/state";
  import { Compartment, StateEffect, StateField } from "@codemirror/state";
  import { Decoration, DecorationSet, EditorView } from "@codemirror/view";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { DuckDBSQL } from "../../../components/editor/presets/duckDBDialect";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { useAllSourceColumns } from "../../sources/selectors";
  import { useAllModelColumns } from "../selectors";
  import Editor from "../../editor/Editor.svelte";
  import { FileArtifact } from "../../entity-management/file-artifacts";

  const schema: { [table: string]: string[] } = {};

  export let selections: SelectionRange[] = [];
  export let autoSave = true;
  export let fileArtifact: FileArtifact;
  export let onSave: (content: string) => void = () => {};

  $: ({ remoteContent } = fileArtifact);

  let editor: EditorView;
  let autocompleteCompartment = new Compartment();

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

  $: defaultTable = getTableNameFromFromClause($remoteContent, schema);
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
</script>

<Editor
  {onSave}
  bind:autoSave
  bind:editor
  {fileArtifact}
  extensions={[
    autocompleteCompartment.of(makeAutocompleteConfig(schema, defaultTable)),
    sql({ dialect: DuckDBSQL }),
  ]}
/>
