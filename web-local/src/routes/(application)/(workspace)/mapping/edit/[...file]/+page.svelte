<script lang="ts">
  import MappingWorkspace from "@rilldata/web-common/features/workspaces/MappingWorkspace.svelte";
  import DeveloperChat from "@rilldata/web-common/features/chat/DeveloperChat.svelte";
  import type { PageData } from "./$types";

  export let data: PageData;

  $: ({ fileArtifact, filePath } = data);
  $: remoteContent = fileArtifact.remoteContent;

  // Extract filename from path (e.g., "/data/mapping.csv" -> "mapping.csv")
  $: fileName = filePath.split("/").pop() ?? "mapping.csv";

  // Parse CSV content into columns and rows
  function parseCSV(content: string): { columns: string[]; rows: string[][] } {
    if (!content || content.trim() === "") {
      return { columns: ["column1", "column2"], rows: [["", ""]] };
    }

    const lines = content.split("\n").filter((line) => line.trim() !== "");
    if (lines.length === 0) {
      return { columns: ["column1", "column2"], rows: [["", ""]] };
    }

    // Parse header
    const columns = parseCSVLine(lines[0]);

    // Parse data rows
    const rows = lines.slice(1).map((line) => {
      const cells = parseCSVLine(line);
      // Ensure each row has the same number of cells as columns
      while (cells.length < columns.length) {
        cells.push("");
      }
      return cells.slice(0, columns.length);
    });

    // Ensure at least one row
    if (rows.length === 0) {
      rows.push(new Array(columns.length).fill(""));
    }

    return { columns, rows };
  }

  // Parse a single CSV line, handling quoted values
  function parseCSVLine(line: string): string[] {
    const result: string[] = [];
    let current = "";
    let inQuotes = false;

    for (let i = 0; i < line.length; i++) {
      const char = line[i];
      const nextChar = line[i + 1];

      if (inQuotes) {
        if (char === '"' && nextChar === '"') {
          current += '"';
          i++; // Skip next quote
        } else if (char === '"') {
          inQuotes = false;
        } else {
          current += char;
        }
      } else {
        if (char === '"') {
          inQuotes = true;
        } else if (char === ",") {
          result.push(current);
          current = "";
        } else {
          current += char;
        }
      }
    }
    result.push(current);

    return result;
  }

  $: csvData = parseCSV($remoteContent ?? "");
</script>

<svelte:head>
  <title>Rill Developer | Edit {fileName}</title>
</svelte:head>

<div class="flex h-full overflow-hidden">
  <div class="flex-1 overflow-hidden">
    <MappingWorkspace
      initialName={fileName}
      initialColumns={csvData.columns}
      initialRows={csvData.rows}
      isEditing={true}
    />
  </div>
  <DeveloperChat />
</div>
