<!--
  Tool block for request_connector_fields: collapsible tool call + result, and a
  placeholder for future connector field forms (.env / YAML).
-->
<script lang="ts">
  import type { V1Tool } from "../../../../../runtime-client";
  import ToolCall from "../tools/ToolCall.svelte";
  import type { RequestConnectorFieldsBlock } from "./request-connector-fields-block.ts";
  import JSONSchemaFormRenderer from "@rilldata/web-common/features/templates/JSONSchemaFormRenderer.svelte";
  import { createConnectorForm } from "@rilldata/web-common/features/sources/modal/FormValidation.ts";
  import { ICONS } from "@rilldata/web-common/features/sources/modal/icons.ts";
  import { Button } from "@rilldata/web-common/components/button";
  import type { Conversation } from "@rilldata/web-common/features/chat/core/conversation.ts";

  export let conversation: Conversation;
  export let block: RequestConnectorFieldsBlock;
  export let tools: V1Tool[] | undefined = undefined;

  const {
    form,
    formId,
    errors: paramsErrors,
    enhance,
    submit: paramsSubmit,
  } = createConnectorForm({
    schemaName: block.schemaName,
    formType: "connector",
    onUpdate: ({ form }) => {
      if (!form.valid) return;
      const formattedValues = Object.entries(form.data).map(
        ([key, value]) => `  ${key}: ${value?.toString()}\n`,
      );
      conversation.draftMessage.set(
        `These are the entered values:\n${formattedValues.join("\n")}`,
      );
      void conversation.sendMessage({});
    },
    schemaOverride: block.schema,
  });

  async function handleFileUpload(
    file: File,
    fieldKey: string,
  ): Promise<string> {
    // TODO
  }

  function onStringInputChange(event: Event) {
    // TODO
  }
</script>

<div class="request-connector-fields-block">
  <ToolCall
    message={block.message}
    resultMessage={block.resultMessage}
    {tools}
    variant="block"
  />
  <div class="connector-fields-placeholder">
    {#if block.llmMessage}
      <div class="text-xs text-fg-muted italic pb-2">{block.llmMessage}</div>
    {/if}
    <form
      id={$formId}
      class="p-2 flex-grow overflow-y-auto border rounded-sm"
      use:enhance
      onsubmit={(e) => {
        e.preventDefault();
        paramsSubmit(e);
      }}
    >
      <JSONSchemaFormRenderer
        schema={block.schema}
        step={"connector"}
        {form}
        errors={$paramsErrors}
        {onStringInputChange}
        {handleFileUpload}
        iconMap={ICONS}
      />
      <div class="flex flex-row">
        <div class="grow"></div>
        <Button type="primary" onClick={paramsSubmit}>Submit</Button>
      </div>
    </form>
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
