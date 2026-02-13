<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";

  import SubmissionError from "@rilldata/web-common/components/forms/SubmissionError.svelte";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { type V1ConnectorDriver } from "@rilldata/web-common/runtime-client";
  import type { ActionResult } from "@sveltejs/kit";
  import type { SuperValidated } from "sveltekit-superforms";

  import type { AddDataFormType } from "./types";
  import MultiStepConnectorFlow from "./MultiStepConnectorFlow.svelte";
  import NeedHelpText from "./NeedHelpText.svelte";
  import { isEmpty } from "./utils";
  import JSONSchemaFormRenderer from "../../templates/JSONSchemaFormRenderer.svelte";
  import { connectorStepStore } from "./connectorStepStore";
  import YamlPreview from "./YamlPreview.svelte";
  import { AddDataFormManager } from "./AddDataFormManager";
  import { createConnectorForm } from "./FormValidation";
  import AddDataFormSection from "./AddDataFormSection.svelte";
  import { get } from "svelte/store";
  import { getConnectorSchema } from "./connector-schemas";
  import {
    getRequiredFieldsForValues,
    getSchemaButtonLabels,
    isVisibleForValues,
  } from "../../templates/schema-utils";
  import { ICONS } from "./icons";

  export let connector: V1ConnectorDriver;
  export let schemaName: string;
  export let formType: AddDataFormType;
  export let isSubmitting: boolean;
  export let onBack: () => void;
  export let onClose: () => void;
  export let onCloseAfterNavigation: () => void = onClose;

  let saveAnyway = false;
  let showSaveAnyway = false;

  // Wire manager-provided onUpdate after declaration below
  let handleOnUpdate: (event: {
    form: SuperValidated<
      Record<string, unknown>,
      string,
      Record<string, unknown>
    >;
    result?: Extract<ActionResult, { type: "success" | "failure" }>;
    cancel?: () => void;
  }) => Promise<void> = async (_event) => {};

  // Create form directly from schema using factory function
  const {
    form: form,
    errors: paramsErrors,
    enhance: paramsEnhance,
    tainted: paramsTainted,
    submit: paramsSubmit,
    submitting: paramsSubmitting,
  } = createConnectorForm({
    schemaName,
    formType,
    onUpdate: (e: any) => handleOnUpdate(e),
  });

  // Create manager with form stores (manager handles orchestration, not form creation)
  const formManager = new AddDataFormManager({
    connector,
    formType,
    formStore: form,
    errorsStore: paramsErrors,
    getSelectedAuthMethod: () =>
      get(connectorStepStore).selectedAuthMethod ?? undefined,
    schemaName,
  });

  const isMultiStepConnector = formManager.isMultiStepConnector;
  const hasExplorerStep = formManager.hasExplorerStep;
  const isStepFlowConnector = isMultiStepConnector || hasExplorerStep;
  const isSourceForm = formManager.isSourceForm;
  const isConnectorForm = formManager.isConnectorForm;
  let activeAuthMethod: string | null = null;
  let prevAuthMethod: string | null = null;
  let stepState = $connectorStepStore;
  let multiStepSubmitDisabled = false;
  let multiStepButtonLabel = "";
  let multiStepLoadingCopy = "";
  let shouldShowSkipLink = false;
  let primaryButtonLabel = "";
  let primaryLoadingCopy = "";

  $: stepState = $connectorStepStore;

  // Form IDs
  const baseFormId = formManager.formId;
  let multiStepFormId = baseFormId;
  let paramsError: string | null = null;
  let paramsErrorDetails: string | undefined = undefined;
  let prevDeploymentType: string | undefined = undefined;

  const connectorSchema = getConnectorSchema(schemaName);

  // Clear errors when connection type changes
  $: {
    const currentDeploymentType = $form.deployment_type as string | undefined;
    if (
      prevDeploymentType !== undefined &&
      currentDeploymentType !== prevDeploymentType
    ) {
      paramsError = null;
      showSaveAnyway = false;
    }
    prevDeploymentType = currentDeploymentType;
  }

  /**
   * Clears error state when user modifies form input.
   * Called from onStringInputChange for text inputs.
   *
   * Note: Select/dropdown changes do NOT trigger this - errors only clear on:
   * - Text input changes (via onStringInputChange)
   * - Deployment type changes (via reactive statement above)
   * This is intentional: changing a dropdown option (other than deployment_type)
   * typically doesn't fix connection errors, so we keep the error visible.
   */
  function clearErrorOnInput() {
    if (paramsError) {
      paramsError = null;
      paramsErrorDetails = undefined;
    }
  }

  // Hide Save Anyway once we advance to the model step in step flow connectors.
  $: if (
    isStepFlowConnector &&
    (stepState.step === "source" || stepState.step === "explorer")
  ) {
    showSaveAnyway = false;
  }

  $: isSubmitDisabled = (() => {
    if (isStepFlowConnector) {
      return multiStepSubmitDisabled;
    }

    // No schema = disable submit (schema is required for all connectors)
    if (!connectorSchema) {
      return true;
    }

    const requiredFields = getRequiredFieldsForValues(
      connectorSchema,
      $form,
      isConnectorForm ? "connector" : "source",
    );
    for (const field of requiredFields) {
      if (!isVisibleForValues(connectorSchema, field, $form)) continue;
      const value = $form[field];
      const errorsForField = $paramsErrors[field] as any;
      if (isEmpty(value) || errorsForField?.length) return true;
    }
    return false;
  })();

  $: formId = isStepFlowConnector ? multiStepFormId || baseFormId : baseFormId;

  $: submitting = $paramsSubmitting;

  // Get button labels from schema (e.g., rill-managed ClickHouse uses "Connect")
  $: schemaButtonLabels = getSchemaButtonLabels(connectorSchema, $form);

  $: primaryButtonLabel = isStepFlowConnector
    ? multiStepButtonLabel
    : formManager.getPrimaryButtonLabel({
        isConnectorForm,
        step: stepState.step,
        submitting,
        schemaButtonLabels,
        selectedAuthMethod: activeAuthMethod ?? undefined,
      });

  $: primaryLoadingCopy = (() => {
    if (isStepFlowConnector) return multiStepLoadingCopy;
    if (schemaButtonLabels?.loading) return schemaButtonLabels.loading;
    return activeAuthMethod === "public"
      ? "Continuing..."
      : "Testing connection...";
  })();

  // Clear Save Anyway state whenever auth method changes (any direction).
  $: if (activeAuthMethod !== prevAuthMethod) {
    prevAuthMethod = activeAuthMethod;
    showSaveAnyway = false;
    saveAnyway = false;
  }

  $: isSubmitting = submitting;

  async function handleSaveAnyway() {
    // Save Anyway should only work for connector forms
    if (!isConnectorForm) {
      return;
    }

    saveAnyway = true;
    const result = await formManager.saveConnectorAnyway({
      queryClient,
      values: $form,
    });
    if (result.ok) {
      // Use quiet close — saveConnectorAnyway already navigated via goto().
      // The normal resetModal() fires a synthetic popstate that races with
      // SvelteKit's router and can revert the navigation.
      onCloseAfterNavigation();
    } else {
      paramsError = result.message;
      paramsErrorDetails = result.details;
    }
    saveAnyway = false;
  }

  $: yamlPreview = formManager.computeYamlPreview({
    stepState,
    isMultiStepConnector: isStepFlowConnector,
    isConnectorForm,
    formValues: $form,
  });
  $: shouldShowSaveAnywayButton = isConnectorForm && showSaveAnyway;
  $: saveAnywayLoading = submitting && saveAnyway;

  handleOnUpdate = formManager.makeOnUpdate({
    onClose,
    queryClient,
    getSelectedAuthMethod: () => activeAuthMethod || undefined,
    setParamsError: (message: string | null, details?: string) => {
      paramsError = message;
      paramsErrorDetails = details;
    },
    setShowSaveAnyway: (value: boolean) => {
      showSaveAnyway = value;
    },
  });

  async function handleFileUpload(
    file: File,
    fieldKey: string,
  ): Promise<string> {
    return formManager.handleFileUpload(file, fieldKey);
  }

  function onStringInputChange(event: Event) {
    clearErrorOnInput();
    formManager.onStringInputChange(
      event,
      $paramsTainted as Record<string, boolean> | null | undefined,
    );
  }
</script>

<div class="add-data-layout flex flex-col h-full w-full md:flex-row">
  <!-- LEFT SIDE PANEL -->
  <div
    class="add-data-form-panel flex-1 flex flex-col min-w-0 md:pr-0 pr-0 relative"
  >
    <div
      class="flex flex-col flex-grow {formManager.formHeight} overflow-y-auto p-6"
    >
      {#if isStepFlowConnector}
        <MultiStepConnectorFlow
          {connector}
          {formManager}
          {form}
          {paramsErrors}
          {paramsEnhance}
          {paramsSubmit}
          {baseFormId}
          {onStringInputChange}
          {handleFileUpload}
          submitting={$paramsSubmitting}
          bind:activeAuthMethod
          bind:isSubmitDisabled={multiStepSubmitDisabled}
          bind:primaryButtonLabel={multiStepButtonLabel}
          bind:primaryLoadingCopy={multiStepLoadingCopy}
          bind:formId={multiStepFormId}
          bind:shouldShowSkipLink
        />
      {:else if connectorSchema}
        <AddDataFormSection
          id={baseFormId}
          enhance={paramsEnhance}
          onSubmit={paramsSubmit}
        >
          <JSONSchemaFormRenderer
            schema={connectorSchema}
            step={isConnectorForm ? "connector" : "source"}
            {form}
            errors={$paramsErrors}
            {onStringInputChange}
            {handleFileUpload}
            iconMap={ICONS}
          />
        </AddDataFormSection>
      {:else}
        <div class="p-4 bg-red-50 border border-red-200 rounded-md">
          <p class="text-red-800 font-medium">Missing connector schema</p>
          <p class="text-red-600 text-sm mt-1">
            No schema found for connector "{connector.name}". Please add a
            schema in connector-schemas.ts.
          </p>
        </div>
      {/if}
    </div>

    <!-- LEFT FOOTER -->
    <div
      class="w-full bg-surface-subtle border-t border-gray-200 p-6 flex justify-between gap-2"
    >
      <Button onClick={() => formManager.handleBack(onBack)} type="secondary"
        >Back</Button
      >

      <div class="flex gap-2">
        {#if shouldShowSaveAnywayButton}
          <Button
            disabled={false}
            loading={saveAnywayLoading}
            loadingCopy="Saving..."
            onClick={handleSaveAnyway}
            type="secondary"
          >
            Save Anyway
          </Button>
        {/if}

        <Button
          disabled={submitting || isSubmitDisabled}
          loading={submitting}
          loadingCopy={primaryLoadingCopy}
          form={formId}
          submitForm
          type="primary"
        >
          {primaryButtonLabel}
        </Button>
      </div>
    </div>
  </div>

  <!-- RIGHT SIDE PANEL -->
  <div
    class="add-data-side-panel flex flex-col gap-6 p-6 bg-surface w-full max-w-full border-l-0 border-t mt-6 pl-0 pt-6 md:w-96 md:min-w-[320px] md:max-w-[400px] md:border-l md:border-t-0 md:mt-0 md:pl-6 justify-between"
  >
    <div class="flex flex-col gap-6 flex-1 overflow-y-auto">
      {#if paramsError}
        <SubmissionError
          message={paramsError ?? ""}
          details={paramsErrorDetails ?? ""}
        />
      {/if}

      <YamlPreview
        title={isStepFlowConnector
          ? stepState.step === "connector"
            ? "Connector preview"
            : "Model preview"
          : isSourceForm
            ? "Model preview"
            : "Connector preview"}
        yaml={yamlPreview}
      />

      {#if shouldShowSkipLink}
        <div class="text-sm leading-normal font-medium text-muted-foreground">
          Already connected? <button
            type="button"
            class="text-sm leading-normal text-primary-500 hover:text-primary-600 font-medium hover:underline break-all"
            on:click={() => formManager.handleSkip()}
          >
            Import your data
          </button>
        </div>
      {/if}
    </div>

    <NeedHelpText {connector} />
  </div>
</div>
