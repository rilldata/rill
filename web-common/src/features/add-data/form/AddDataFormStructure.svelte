<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import { defaultFormLabels } from "@rilldata/web-common/features/add-data/form/form-labels.ts";
  import SubmissionError from "@rilldata/web-common/components/forms/SubmissionError.svelte";
  import type { createConnectorForm } from "@rilldata/web-common/features/sources/modal/FormValidation.ts";
  import YamlPreview from "@rilldata/web-common/features/sources/modal/YamlPreview.svelte";
  import { ICONS } from "@rilldata/web-common/features/sources/modal/icons.ts";
  import JSONSchemaFormRenderer from "@rilldata/web-common/features/templates/JSONSchemaFormRenderer.svelte";
  import type { MultiStepFormSchema } from "@rilldata/web-common/features/templates/schemas/types.ts";
  import { processFileContent } from "@rilldata/web-common/features/templates/file-encoding.ts";
  import {
    inferModelNameFromSQL,
    inferSourceName,
  } from "@rilldata/web-common/features/sources/sourceUtils.ts";
  import type { V1ConnectorDriver } from "@rilldata/web-common/runtime-client";

  export let connectorDriver: V1ConnectorDriver;
  export let schema: MultiStepFormSchema | null;
  export let superFormsParams: ReturnType<typeof createConnectorForm>;
  export let labels = defaultFormLabels;
  export let yamlPreview: string;
  export let step: "connector" | "source";
  export let onBack: () => void;

  $: ({ form, formId, tainted, submit, submitting, errors, enhance } =
    superFormsParams);
  $: taintedFields = $tainted;

  let shouldShowSaveAnywayButton = false;
  let shouldShowSkipLink = false;
  let isSubmitDisabled = false; // TODO

  $: error = $errors._errors?.[0]; // TODO

  function onStringInputChange(e: Event) {
    const target = e.target as HTMLInputElement;
    const { name, value } = target;

    if (name !== "path" && name !== "sql") return;

    const nameFieldTainted =
      taintedFields && typeof taintedFields === "object"
        ? Boolean(taintedFields?.name)
        : false;
    if (nameFieldTainted) return;

    const inferredName =
      name === "sql"
        ? inferModelNameFromSQL(value)
        : inferSourceName(connectorDriver, value);
    if (!inferredName) return;

    form.update(
      ($form) => {
        $form.name = inferredName;
        return $form;
      },
      { taint: false },
    );
  }

  async function handleFileUpload(
    file: File,
    fieldKey?: string,
  ): Promise<string> {
    const content = await file.text();

    if (fieldKey) {
      const field = schema?.properties?.[fieldKey];
      if (field?.["x-file-encoding"]) {
        const result = processFileContent(content, field);

        if (Object.keys(result.extractedValues).length > 0) {
          form.update(
            ($form) => {
              for (const [key, value] of Object.entries(
                result.extractedValues,
              )) {
                $form[key] = value;
              }
              return $form;
            },
            { taint: false },
          );
        }

        return result.encodedContent;
      }
    }

    return content;
  }
</script>

<div class="flex flex-col h-full w-full md:flex-row overflow-y-auto">
  <!-- LEFT SIDE PANEL -->
  <div class="flex-1 flex flex-col min-w-0 h-full md:pr-0 pr-0 relative">
    <div class="flex flex-col flex-grow overflow-y-auto p-6">
      <form
        id={$formId}
        class="pb-5"
        use:enhance
        on:submit|preventDefault={submit}
      >
        <JSONSchemaFormRenderer
          {schema}
          {step}
          {form}
          {errors}
          {onStringInputChange}
          {handleFileUpload}
          iconMap={ICONS}
        />
      </form>
    </div>

    <!-- LEFT FOOTER -->
    <div
      class="w-full bg-surface-subtle border-t border-gray-200 p-6 flex justify-between gap-2"
    >
      <Button onClick={onBack} type="secondary">Back</Button>

      <div class="flex gap-2">
        {#if shouldShowSaveAnywayButton}
          <Button type="secondary">Save (TODO)</Button>
        {/if}

        <Button
          disabled={$submitting || isSubmitDisabled}
          loading={$submitting}
          loadingCopy={labels.primaryLoadingCopy}
          form={$formId}
          submitForm
          type="primary"
        >
          {labels.primaryButtonLabel}
        </Button>
      </div>
    </div>
  </div>

  <!-- RIGHT SIDE PANEL -->
  <div
    class="add-data-side-panel flex flex-col gap-6 p-6 bg-surface w-full max-w-full border-l-0 border-t mt-6 pl-0 pt-6 md:w-96 md:min-w-[320px] md:max-w-[400px] md:border-l md:border-t-0 md:mt-0 md:pl-6 justify-between"
  >
    <div class="flex flex-col gap-6 flex-1 overflow-y-auto">
      {#if error}
        <SubmissionError message={error} />
      {/if}

      <YamlPreview title={labels.yamlPreviewTitle} yaml={yamlPreview} />

      {#if shouldShowSkipLink}
        <div class="text-sm leading-normal font-medium text-muted-foreground">
          Already connected? <button
            type="button"
            class="text-sm leading-normal text-primary-500 hover:text-primary-600 font-medium hover:underline break-all"
          >
            Import your data (TODO)
          </button>
        </div>
      {/if}
    </div>
  </div>
</div>
