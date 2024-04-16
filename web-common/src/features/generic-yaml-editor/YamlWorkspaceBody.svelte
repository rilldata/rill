<script lang="ts">
  import type { EditorView } from "@codemirror/view";
  import { removeLeadingSlash } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { parse } from "yaml";
  import YAMLEditor from "../../components/editor/YAMLEditor.svelte";
  import {
    createRuntimeServiceGetFile,
    runtimeServicePutFile,
  } from "../../runtime-client";
  import { runtime } from "../../runtime-client/runtime-store";
  import ErrorPane from "./ErrorPane.svelte";
  import { debounce } from "@rilldata/web-common/lib/create-debouncer";

  export let filePath: string;

  let editor: YAMLEditor;
  let view: EditorView;
  let error: Error | undefined;

  $: file = createRuntimeServiceGetFile(
    $runtime.instanceId,
    removeLeadingSlash(filePath),
    {
      query: {
        // this will ensure that any changes done outside our app is pulled in.
        refetchOnWindowFocus: true,
      },
    },
  );

  let content = "";
  $: content = $file?.data?.blob ?? content;

  const debouncedUpdate = debounce(handleUpdate, 300);

  async function handleUpdate(e: CustomEvent<{ content: string }>) {
    const blob = e.detail.content;
    await runtimeServicePutFile(
      $runtime.instanceId,
      removeLeadingSlash(filePath),
      {
        blob: blob,
      },
    );
    error = validateYAMLAndReturnError(blob);
  }

  function validateYAMLAndReturnError(blob: string): Error | undefined {
    try {
      parse(blob);
      return undefined;
    } catch (e) {
      return e;
    }
  }

  function cleanErrorMessage(message: string): string {
    return message?.replace("YAMLParseError: ", "");
  }
</script>

<div
  class="flex flex-col w-full h-full content-stretch"
  style:height={"calc(100vh - var(--header-height))"}
>
  <div class="grow bg-white overflow-y-auto">
    <div
      class="border-white w-full overflow-y-auto h-full"
      class:border-b-hidden={error}
      class:border-red-500={error}
    >
      <YAMLEditor
        bind:this={editor}
        bind:view
        {content}
        whenFocused
        on:update={debouncedUpdate}
      />
    </div>
  </div>
  {#if error}
    <ErrorPane errorMessage={cleanErrorMessage(error.message)} />
  {/if}
</div>
