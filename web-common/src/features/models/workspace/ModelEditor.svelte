<script lang="ts">
  import { autocompletion } from "@codemirror/autocomplete";
  import {
    keywordCompletionSource,
    schemaCompletionSource,
    sql,
  } from "@codemirror/lang-sql";
  import { Compartment } from "@codemirror/state";
  import { EditorView } from "@codemirror/view";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { DuckDBSQL } from "../../../components/editor/presets/duckDBDialect";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { useAllSourceColumns } from "../../sources/selectors";
  import { useAllModelColumns } from "../selectors";
  import Editor from "../../editor/Editor.svelte";
  import { FileArtifact } from "../../entity-management/file-artifact";

  const schema: { [table: string]: string[] } = {};

  export let autoSave = true;
  export let fileArtifact: FileArtifact;
  export let onSave: (content: string) => void = () => {};

  $: ({ instanceId } = $runtime);

  $: ({ remoteContent } = fileArtifact);

  let editor: EditorView;
  let autocompleteCompartment = new Compartment();

  // Autocomplete: source tables
  $: allSourceColumns = useAllSourceColumns(queryClient, instanceId);
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
  $: allModelColumns = useAllModelColumns(queryClient, instanceId);
  $: if ($allModelColumns?.length) {
    for (const modelTable of $allModelColumns) {
      const modelIdentifier = modelTable?.tableName;
      schema[modelIdentifier] = modelTable.profileColumns
        ?.filter((c) => c.name !== undefined)
        // CAST SAFETY: already filtered out undefined values
        ?.map((c) => c.name as string);
    }
  }

  $: defaultTable = getTableNameFromFromClause($remoteContent ?? "", schema);
  $: updateAutocompleteSources(schema, defaultTable);

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

  function updateAutocompleteSources(
    schema: { [table: string]: string[] },
    defaultTable?: string,
  ) {
    if (editor) {
      queueMicrotask(() => {
        editor.dispatch({
          effects: autocompleteCompartment.reconfigure(
            makeAutocompleteConfig(schema, defaultTable),
          ),
        });
      });
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
