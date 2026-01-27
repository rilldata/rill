<script lang="ts">
  import { goto } from "$app/navigation";
  import { onMount } from "svelte";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import InputWithConfirm from "@rilldata/web-common/components/forms/InputWithConfirm.svelte";
  import { FileSpreadsheet, Plus, Trash2, Save } from "lucide-svelte";
  import WorkspaceContainer from "@rilldata/web-common/layout/workspace/WorkspaceContainer.svelte";
  import { createRuntimeServicePutFile } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { getFileNamesInDirectory } from "@rilldata/web-common/features/entity-management/file-selectors";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";

  export let initialName = "mapping.csv";
  export let initialColumns: string[] | undefined = undefined;
  export let initialRows: string[][] | undefined = undefined;
  export let isEditing = false;

  const createFile = createRuntimeServicePutFile();

  let fileName = initialName;
  let hasUnsavedChanges = !isEditing;
  let editing = false;
  let isSaving = false;

  // Table state
  let columns: string[] = initialColumns ?? ["column1", "column2"];
  let rows: string[][] = initialRows ?? [["", ""]];

  $: ({ instanceId } = $runtime);

  // Compute base name without .csv extension for display
  $: baseName = fileName.endsWith(".csv") ? fileName.slice(0, -4) : fileName;

  // On mount, check for unique name if creating new mapping
  onMount(async () => {
    if (!isEditing && instanceId) {
      const uniqueName = await getUniqueBaseName(baseName);
      if (uniqueName !== baseName) {
        fileName = `${uniqueName}.csv`;
      }
    }
  });

  function addColumn() {
    const newColName = `column${columns.length + 1}`;
    columns = [...columns, newColName];
    rows = rows.map((row) => [...row, ""]);
    hasUnsavedChanges = true;
  }

  function removeColumn(index: number) {
    if (columns.length <= 1) return;
    columns = columns.filter((_, i) => i !== index);
    rows = rows.map((row) => row.filter((_, i) => i !== index));
    hasUnsavedChanges = true;
  }

  function addRow() {
    rows = [...rows, new Array(columns.length).fill("")];
    hasUnsavedChanges = true;
  }

  function removeRow(index: number) {
    if (rows.length <= 1) return;
    rows = rows.filter((_, i) => i !== index);
    hasUnsavedChanges = true;
  }

  function updateCell(rowIndex: number, colIndex: number, value: string) {
    rows[rowIndex][colIndex] = value;
    rows = rows;
    hasUnsavedChanges = true;
  }

  function updateColumnName(index: number, value: string) {
    columns[index] = value;
    columns = columns;
    hasUnsavedChanges = true;
  }

  function handleKeydown(
    e: KeyboardEvent,
    rowIndex?: number,
    colIndex?: number,
  ) {
    // Handle Ctrl+A or Cmd+A to select all text in the current cell
    if ((e.ctrlKey || e.metaKey) && e.key === "a") {
      e.preventDefault();
      const input = e.target as HTMLInputElement;
      input.select();
    }

    // Handle Tab at last cell of last row - auto create new row
    if (
      e.key === "Tab" &&
      !e.shiftKey &&
      rowIndex !== undefined &&
      colIndex !== undefined &&
      rowIndex === rows.length - 1 &&
      colIndex === columns.length - 1
    ) {
      e.preventDefault();
      addRow();
      // Focus first cell of new row after DOM updates
      setTimeout(() => {
        const newRowInputs = document.querySelectorAll(
          `[data-row="${rows.length - 1}"]`,
        );
        if (newRowInputs.length > 0) {
          (newRowInputs[0] as HTMLInputElement).focus();
        }
      }, 0);
    }
  }

  function generateCSV(): string {
    const header = columns.join(",");
    const dataRows = rows.map((row) =>
      row
        .map((cell) => {
          // Escape quotes and wrap in quotes if contains comma or quote
          if (cell.includes(",") || cell.includes('"') || cell.includes("\n")) {
            return `"${cell.replace(/"/g, '""')}"`;
          }
          return cell;
        })
        .join(","),
    );
    return [header, ...dataRows].join("\n");
  }

  function generateModelYAML(csvPath: string): string {
    return `# Model YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/models

type: model
materialize: true

connector: duckdb
sql: "select * from read_csv('${csvPath}', auto_detect=true, ignore_errors=1, header=true)"
`;
  }

  async function getUniqueBaseName(name: string): Promise<string> {
    // Get existing files in both directories
    const [dataFiles, modelFiles] = await Promise.all([
      getFileNamesInDirectory(queryClient, instanceId, "/data"),
      getFileNamesInDirectory(queryClient, instanceId, "/models"),
    ]);

    const existingFiles = new Set([...dataFiles, ...modelFiles]);

    // Check if base name files exist
    if (
      !existingFiles.has(`${name}.csv`) &&
      !existingFiles.has(`${name}.yaml`)
    ) {
      return name;
    }

    // Auto-increment with _# suffix
    let counter = 1;
    while (
      existingFiles.has(`${name}_${counter}.csv`) ||
      existingFiles.has(`${name}_${counter}.yaml`)
    ) {
      counter++;
    }

    return `${name}_${counter}`;
  }

  async function handleSave() {
    if (isSaving) return;
    isSaving = true;

    try {
      // Get unique base name (auto-increment if files exist)
      const uniqueBaseName = isEditing
        ? baseName
        : await getUniqueBaseName(baseName);

      const csvPath = `data/${uniqueBaseName}.csv`;
      const modelPath = `models/${uniqueBaseName}.yaml`;

      const csvContent = generateCSV();
      const modelContent = generateModelYAML(csvPath);

      // Create CSV file
      await $createFile.mutateAsync({
        instanceId,
        data: {
          path: csvPath,
          blob: csvContent,
          create: true,
          createOnly: false,
        },
      });

      // Create model YAML file
      await $createFile.mutateAsync({
        instanceId,
        data: {
          path: modelPath,
          blob: modelContent,
          create: true,
          createOnly: false,
        },
      });

      hasUnsavedChanges = false;

      // Navigate to the model file
      await goto(`/files/${modelPath}`);
    } catch (error) {
      console.error("Failed to save mapping files:", error);
    } finally {
      isSaving = false;
    }
  }

  function handleNameChange(newName: string) {
    fileName = newName;
    hasUnsavedChanges = true;
  }
</script>

<WorkspaceContainer inspector={false}>
  <header slot="header" class="flex flex-col py-2 gap-y-2">
    <div class="second-level-wrapper">
      <div class="flex gap-x-1 items-center w-full" class:truncate={!editing}>
        <span class="flex-none">
          <FileSpreadsheet size="19px" color="#10B981" />
        </span>

        <InputWithConfirm
          bind:editing
          size="md"
          editable={true}
          id="mapping-title-input"
          textClass="text-xl font-semibold"
          value={fileName}
          onConfirm={handleNameChange}
          showIndicator={hasUnsavedChanges}
        />
      </div>

      <div class="flex items-center gap-x-2 w-fit flex-none">
        <Button type="primary" disabled={isSaving} onClick={handleSave}>
          <Save size="14px" />
          {isSaving ? "Saving..." : "Save"}
        </Button>
      </div>
    </div>
  </header>

  <div slot="body" class="h-full overflow-hidden overflow-y-auto p-4">
    <div class="flex flex-col gap-4">
      <div class="flex gap-2">
        <Button type="secondary" onClick={addColumn}>
          <Plus size="14px" /> Add Column
        </Button>
        <Button type="secondary" onClick={addRow}>
          <Plus size="14px" /> Add Row
        </Button>
      </div>

      <div class="table-container border rounded-md overflow-auto">
        <table class="w-full border-collapse">
          <thead>
            <tr class="bg-gray-50">
              <th
                class="w-10 border-b border-r p-2 text-center text-gray-500 text-sm"
              >
                #
              </th>
              {#each columns as col, colIndex (colIndex)}
                <th class="border-b border-r p-0 min-w-[150px]">
                  <div class="flex items-center">
                    <input
                      type="text"
                      value={col}
                      on:input={(e) =>
                        updateColumnName(colIndex, e.currentTarget.value)}
                      on:keydown={handleKeydown}
                      class="flex-1 p-2 font-semibold text-sm bg-transparent border-none outline-none focus:bg-blue-50"
                    />
                    {#if columns.length > 1}
                      <button
                        class="p-1 text-gray-400 hover:text-red-500"
                        on:click={() => removeColumn(colIndex)}
                      >
                        <Trash2 size="14px" />
                      </button>
                    {/if}
                  </div>
                </th>
              {/each}
              <th class="w-10 border-b p-2"></th>
            </tr>
          </thead>
          <tbody>
            {#each rows as row, rowIndex (rowIndex)}
              <tr class="hover:bg-gray-50">
                <td
                  class="border-b border-r p-2 text-center text-gray-500 text-sm bg-gray-50"
                >
                  {rowIndex + 1}
                </td>
                {#each row as cell, colIndex (colIndex)}
                  <td class="border-b border-r p-0 bg-white">
                    <input
                      type="text"
                      value={cell}
                      data-row={rowIndex}
                      data-col={colIndex}
                      on:input={(e) =>
                        updateCell(rowIndex, colIndex, e.currentTarget.value)}
                      on:keydown={(e) => handleKeydown(e, rowIndex, colIndex)}
                      class="w-full p-2 text-sm bg-white border-none outline-none focus:bg-blue-50"
                    />
                  </td>
                {/each}
                <td class="border-b p-2 text-center">
                  {#if rows.length > 1}
                    <button
                      class="p-1 text-gray-400 hover:text-red-500"
                      on:click={() => removeRow(rowIndex)}
                    >
                      <Trash2 size="14px" />
                    </button>
                  {/if}
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    </div>
  </div>
</WorkspaceContainer>

<style lang="postcss">
  .second-level-wrapper {
    @apply px-4 py-2 w-full h-7;
    @apply flex justify-between gap-x-2;
    @apply items-center;
  }

  table {
    border-spacing: 0;
  }

  input:focus {
    outline: none;
  }
</style>
