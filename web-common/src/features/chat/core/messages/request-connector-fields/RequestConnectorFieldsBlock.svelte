<!--
  Tool block for request_connector_fields: collapsible tool call + result, and a
  placeholder for future connector field forms (.env / YAML).
-->
<script lang="ts">
  import { onMount } from "svelte";
  import type { V1Tool } from "../../../../../runtime-client";
  import ToolCall from "../tools/ToolCall.svelte";
  import type { RequestConnectorFieldsBlock } from "./request-connector-fields-block.ts";
  import type { Conversation } from "@rilldata/web-common/features/chat/core/conversation.ts";
  import PartialConnectorForm from "@rilldata/web-common/features/chat/core/messages/request-connector-fields/PartialConnectorForm.svelte";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts.ts";
  import { decompileConnectorYAML } from "@rilldata/web-common/features/connectors/code-utils.ts";
  import { getSchemaFieldMetaList } from "@rilldata/web-common/features/templates/schema-utils.ts";

  export let conversation: Conversation;
  export let block: RequestConnectorFieldsBlock;
  export let tools: V1Tool[] | undefined = undefined;

  let existingData: Record<string, any> = {};
  let existingDataLoaded = false;

  async function loadConnectorData() {
    if (!block.connectorPath) return;
    const connectorFile = fileArtifacts.getFileArtifact(block.connectorPath);
    const connectorYaml = await connectorFile.fetchContent(false);
    existingData = decompileConnectorYAML(
      connectorYaml,
      getSchemaFieldMetaList(block.schema, { step: "connector" }),
    );
  }

  function onSubmit(newConnectorPath: string) {
    conversation.draftMessage.set(
      `Values have been save to "${newConnectorPath}". Continue with the connector creation.`,
    );
    void conversation.sendMessage({});
  }

  onMount(loadConnectorData);
</script>

<div class="request-connector-fields-block">
  <ToolCall
    message={block.message}
    resultMessage={block.resultMessage}
    {tools}
    variant="block"
  />
  <div class="connector-fields-placeholder">
    {#if !block.hasSubmitted}
      {#if block.llmMessage}
        <div class="text-xs text-fg-muted italic pb-2">{block.llmMessage}</div>
      {/if}
      {#if existingDataLoaded}
        <PartialConnectorForm
          schemaName={block.schemaName}
          schema={block.filteredSchema}
          connectorPath={block.connectorPath}
          {existingData}
          {onSubmit}
        />
      {/if}
    {/if}
  </div>
</div>

<style lang="postcss">
  .request-connector-fields-block {
    @apply w-full max-w-full self-start flex flex-col gap-2;
  }

  .connector-fields-placeholder {
    @apply min-h-0;
  }
</style>
