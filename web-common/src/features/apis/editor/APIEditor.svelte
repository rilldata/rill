<script lang="ts">
  import type { EditorView } from "@codemirror/view";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { setLineStatuses } from "@rilldata/web-common/components/editor/line-status";
  import type { LineStatus } from "@rilldata/web-common/components/editor/line-status/state";
  import { customYAMLwithJSONandSQL } from "@rilldata/web-common/components/editor/presets/yamlWithJsonAndSql";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import Trash from "@rilldata/web-common/components/icons/Trash.svelte";
  import Editor from "@rilldata/web-common/features/editor/Editor.svelte";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import Resizer from "@rilldata/web-common/layout/Resizer.svelte";
  import WorkspaceEditorContainer from "@rilldata/web-common/layout/workspace/WorkspaceEditorContainer.svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { PlusIcon, PlayIcon, ChevronDownIcon } from "lucide-svelte";
  import APIResponsePreview from "./APIResponsePreview.svelte";

  export let apiName: string;
  export let errors: LineStatus[];
  export let fileArtifact: FileArtifact;
  export let autoSave: boolean;

  interface Arg {
    id: string;
    key: string;
    value: string;
  }

  let editor: EditorView;
  let args: Arg[] = [];
  let apiResponse: unknown[] | null = null;
  let responseError: string | null = null;
  let isLoading = false;
  let previewHeight = 250;

  $: ({ instanceId } = $runtime);
  $: if (editor) setLineStatuses(errors, editor);
  $: mainError = errors?.at(0);
  $: host = $runtime.host || "http://localhost:9009";
  $: baseUrl = `${host}/v1/instances/${instanceId}/api/${apiName}`;
  $: fullUrl = buildFullUrl(baseUrl, args);
  $: hasErrors = errors.length > 0;

  function buildFullUrl(base: string, params: Arg[]): string {
    const url = new URL(base);
    params.forEach((arg) => {
      if (arg.key.trim()) {
        url.searchParams.set(arg.key, arg.value);
      }
    });
    return url.toString();
  }

  function addArg() {
    args = [...args, { id: crypto.randomUUID(), key: "", value: "" }];
  }

  function removeArg(id: string) {
    args = args.filter((arg) => arg.id !== id);
  }

  async function testAPI() {
    isLoading = true;
    responseError = null;
    apiResponse = null;

    try {
      const response = await fetch(fullUrl);

      if (!response.ok) {
        const errorText = await response.text();
        try {
          const errorJson = JSON.parse(errorText);
          responseError = errorJson.message || errorJson.error || errorText;
        } catch {
          responseError = errorText;
        }
        return;
      }

      const data = await response.json();
      apiResponse = Array.isArray(data) ? data : [data];
    } catch (e) {
      responseError = e instanceof Error ? e.message : "Unknown error occurred";
    } finally {
      isLoading = false;
    }
  }
</script>

<div class="flex flex-col h-full overflow-hidden p-2">
  <div class="editor-panel">
    <WorkspaceEditorContainer error={mainError}>
      <Editor
        bind:autoSave
        bind:editor
        onSave={(content) => {
          if (!content?.length) {
            setLineStatuses([], editor);
          }
        }}
        {fileArtifact}
        extensions={[customYAMLwithJSONandSQL]}
      />
    </WorkspaceEditorContainer>
  </div>

  <div
    class="preview-panel"
    style:height="{previewHeight}px"
    style:min-height="100px"
    style:max-height="60%"
  >
    <Resizer
      max={500}
      direction="NS"
      side="top"
      bind:dimension={previewHeight}
    />

    <div class="flex items-center gap-x-3 px-3 py-2 border-b">
      <div class="flex items-center gap-x-2 flex-1 min-w-0">
        <span class="text-xs font-medium text-fg-secondary shrink-0">GET</span>
        <span class="text-xs font-mono text-fg-muted truncate">{fullUrl}</span>
      </div>

      <div class="flex items-center gap-x-2 shrink-0">
        <DropdownMenu.Root closeOnItemClick={false}>
          <DropdownMenu.Trigger asChild let:builder>
            <Button type="text" compact small builders={[builder]}>
              Args
              {#if args.length > 0}
                <span
                  class="inline-flex items-center justify-center w-4 h-4 text-[10px] font-medium bg-surface-active text-fg-accent rounded-full"
                >
                  {args.length}
                </span>
              {/if}
              <ChevronDownIcon size="10px" />
            </Button>
          </DropdownMenu.Trigger>
          <DropdownMenu.Content align="end" class="w-72 p-2">
            <div class="flex flex-col gap-y-2">
              {#if args.length === 0}
                <p class="text-xs text-fg-muted px-1 py-2">
                  No arguments. Click "Add" below.
                </p>
              {:else}
                {#each args as arg (arg.id)}
                  <div class="flex items-center gap-x-1">
                    <Input
                      bind:value={arg.key}
                      placeholder="key"
                      size="sm"
                      width="100px"
                    />
                    <Input
                      bind:value={arg.value}
                      placeholder="value"
                      size="sm"
                      full
                    />
                    <Button
                      type="ghost"
                      square
                      small
                      compact
                      onClick={() => removeArg(arg.id)}
                    >
                      <Trash size="12px" />
                    </Button>
                  </div>
                {/each}
              {/if}
              <Button type="text" compact small onClick={addArg}>
                <PlusIcon size="12px" />
                Add
              </Button>
            </div>
          </DropdownMenu.Content>
        </DropdownMenu.Root>

        <Button
          type="primary"
          small
          onClick={testAPI}
          disabled={hasErrors}
          loading={isLoading}
          loadingCopy="Testing"
        >
          <PlayIcon size="12px" />
          Test API
        </Button>
      </div>
    </div>

    <div class="flex-1 overflow-auto">
      <APIResponsePreview
        response={apiResponse}
        error={responseError}
        {isLoading}
        {apiName}
      />
    </div>
  </div>
</div>

<style lang="postcss">
  .editor-panel {
    @apply flex-1 overflow-hidden min-h-0 border rounded-[2px] rounded-b-none border-b-0;
  }

  .preview-panel {
    @apply relative flex flex-col bg-surface-background border rounded-[2px] rounded-t-none overflow-hidden;
  }
</style>
