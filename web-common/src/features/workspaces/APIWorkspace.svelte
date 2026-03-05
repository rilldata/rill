<script lang="ts">
  import { goto } from "$app/navigation";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { getNameFromFile } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import {
    ResourceKind,
    resourceIsLoading,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { handleEntityRename } from "@rilldata/web-common/features/entity-management/ui-actions";
  import { mapParseErrorsToLines } from "@rilldata/web-common/features/metrics-views/errors";
  import APIEditor from "@rilldata/web-common/features/apis/editor/APIEditor.svelte";
  import type { Arg } from "@rilldata/web-common/features/apis/editor/types";
  import WorkspaceContainer from "@rilldata/web-common/layout/workspace/WorkspaceContainer.svelte";
  import WorkspaceHeader from "@rilldata/web-common/layout/workspace/WorkspaceHeader.svelte";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { ChevronDownIcon } from "lucide-svelte";

  export let fileArtifact: FileArtifact;

  $: ({ instanceId } = $runtime);
  $: ({
    hasUnsavedChanges,
    autoSave,
    path: filePath,
    resourceName,
    remoteContent,
    fileName,
  } = fileArtifact);

  $: apiName = $resourceName?.name ?? getNameFromFile(filePath);
  $: host = $runtime.host || "http://localhost:9009";

  $: allErrorsQuery = fileArtifact.getAllErrors(queryClient, instanceId);
  $: allErrors = $allErrorsQuery;
  $: resourceQuery = fileArtifact.getResource(queryClient, instanceId);
  $: ({ data: resource } = $resourceQuery);
  $: isReconciling = resourceIsLoading($resourceQuery.data);

  async function onChangeCallback(newTitle: string) {
    const newRoute = await handleEntityRename(
      instanceId,
      newTitle,
      filePath,
      fileName,
    );
    if (newRoute) await goto(newRoute);
  }

  $: errors = mapParseErrorsToLines(allErrors, $remoteContent ?? "");

  let args: Arg[] = [];

  // Templates that modify the SQL/metrics_sql value in the YAML
  const templates = [
    {
      label: "Filter Template",
      clause: `where dimension = '{{ .args.filter }}'`,
    },
    {
      label: "Limit Template",
      clause: "limit {{ .args.limit }}",
    },
    {
      label: "Offset Template",
      clause: "offset {{ .args.offset }}",
    },
    {
      label: "Sort Template",
      clause: "order by {{ .args.sort }} {{ .args.order }}",
    },
    {
      label: "Time range Template",
      clause: `where time >= '{{ .args.start }}' and time < '{{ .args.end }}'`,
    },
  ];

  $: ({ editorContent, updateEditorContent } = fileArtifact);

  // Matches FROM <table> and captures everything after it as trailing clauses
  const fromPattern = /\bfrom\s+\S+/i;

  function applyTemplate(clause: string) {
    const content = $editorContent ?? "";

    // Find which SQL key is present
    const sqlKeyMatch = content.match(/^(metrics_sql|sql)\s*:/m);
    if (!sqlKeyMatch) return;

    const keyIndex = content.indexOf(sqlKeyMatch[0]);
    const afterKey = content.slice(keyIndex + sqlKeyMatch[0].length);

    // Determine if block scalar (| or >) or inline
    const isBlock = /^\s*\|/.test(afterKey);

    if (isBlock) {
      // Block scalar: collect indented lines after "|\n"
      const blockStart = afterKey.indexOf("\n") + 1;
      const fullBlockStart = keyIndex + sqlKeyMatch[0].length + blockStart;

      const restLines = content.slice(fullBlockStart).split("\n");
      const blockLines: string[] = [];
      let blockEnd = fullBlockStart;
      for (const line of restLines) {
        if (line.match(/^\s+\S/) || line.trim() === "") {
          blockLines.push(line);
          blockEnd += line.length + 1;
        } else {
          break;
        }
      }

      const blockText = blockLines.join("\n");
      const match = fromPattern.exec(blockText);
      if (!match) return;

      // Keep everything up to and including "FROM <table>", replace rest
      const indent = blockLines[0]?.match(/^(\s+)/)?.[1] ?? "  ";
      const baseQuery = blockText.slice(0, match.index + match[0].length);
      const newBlock = baseQuery + "\n" + indent + clause;

      const before = content.slice(0, fullBlockStart);
      const after = content.slice(blockEnd);
      updateEditorContent(before + newBlock + "\n" + after);
    } else {
      // Inline: sql: select ... from table where ...
      const valueStart = keyIndex + sqlKeyMatch[0].length;
      const lineEnd = content.indexOf("\n", valueStart);
      const sqlValue = content.slice(
        valueStart,
        lineEnd === -1 ? undefined : lineEnd,
      );

      const match = fromPattern.exec(sqlValue);
      if (!match) return;

      // Keep everything up to and including "FROM <table>", replace rest
      const baseQuery = sqlValue.slice(0, match.index + match[0].length);
      const newValue = baseQuery + " " + clause;

      const before = content.slice(0, valueStart);
      const after = lineEnd === -1 ? "" : content.slice(lineEnd);
      updateEditorContent(before + newValue + after);
    }
  }
</script>

<WorkspaceContainer inspector={false}>
  <WorkspaceHeader
    {filePath}
    {resource}
    resourceKind={ResourceKind.API}
    hasUnsavedChanges={$hasUnsavedChanges}
    onTitleChange={onChangeCallback}
    slot="header"
    showInspectorToggle={false}
    titleInput={fileName}
  >
    <svelte:fragment slot="workspace-controls">
      <DropdownMenu.Root>
        <DropdownMenu.Trigger asChild let:builder>
          <Button type="text" compact small builders={[builder]}>
            Templates
            <ChevronDownIcon size="12px" />
          </Button>
        </DropdownMenu.Trigger>
        <DropdownMenu.Content align="end" class="w-56">
          {#each templates as template}
            <DropdownMenu.Item
              on:click={() => applyTemplate(template.clause)}
            >
              <span class="text-sm">{template.label}</span>
            </DropdownMenu.Item>
          {/each}
        </DropdownMenu.Content>
      </DropdownMenu.Root>
    </svelte:fragment>
  </WorkspaceHeader>

  <svelte:fragment slot="body">
    <APIEditor
      bind:autoSave={$autoSave}
      {fileArtifact}
      {errors}
      {apiName}
      {isReconciling}
      {host}
      {instanceId}
      bind:args
    />
  </svelte:fragment>
</WorkspaceContainer>
