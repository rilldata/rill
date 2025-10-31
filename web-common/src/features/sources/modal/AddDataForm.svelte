<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";

  import SubmissionError from "@rilldata/web-common/components/forms/SubmissionError.svelte";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { type V1ConnectorDriver } from "@rilldata/web-common/runtime-client";
  import type { ActionResult } from "@sveltejs/kit";
  import { createEventDispatcher } from "svelte";
  import {
    defaults,
    superForm,
    type SuperValidated,
  } from "sveltekit-superforms";
  import { yup } from "sveltekit-superforms/adapters";
  import {
    inferSourceName,
    prepareSourceFormData,
    compileSourceYAML,
  } from "../sourceUtils";
  import { humanReadableErrorMessage } from "../errors/errors";
  import {
    submitAddConnectorForm,
    submitAddSourceForm,
  } from "./submitAddDataForm";
  import type { AddDataFormType, ConnectorType } from "./types";
  import {
    dsnSchema as dsnValidation,
    getValidationSchemaForConnector,
  } from "./FormValidation";
  import AddClickHouseForm from "./AddClickHouseForm.svelte";
  import NeedHelpText from "./NeedHelpText.svelte";
  import Tabs from "@rilldata/web-common/components/forms/Tabs.svelte";
  import { TabsContent } from "@rilldata/web-common/components/tabs";
  import { isEmpty } from "./utils";
  import {
    CONNECTION_TAB_OPTIONS,
    type ClickHouseConnectorType,
  } from "./constants";
  import { getInitialFormValuesFromProperties } from "../sourceUtils";
  import { compileConnectorYAML } from "../../connectors/code-utils";
  import { MULTI_STEP_CONNECTORS } from "./constants";
  import {
    connectorStepStore,
    setStep,
    setConnectorConfig,
  } from "./connectorStepStore";
  import FormRenderer from "./FormRenderer.svelte";
  import YamlPreview from "./YamlPreview.svelte";
  import GCSMultiStepForm from "./GCSMultiStepForm.svelte";

  const dispatch = createEventDispatcher();

  export let connector: V1ConnectorDriver;
  export let formType: AddDataFormType;
  export let onBack: () => void;
  export let onClose: () => void;

  const isSourceForm = formType === "source";
  const isConnectorForm = formType === "connector";

  let connectionTab: ConnectorType = "parameters";

  // Simple multi-step state management
  const isMultiStepConnector = MULTI_STEP_CONNECTORS.includes(
    connector.name ?? "",
  );
  $: stepState = $connectorStepStore;

  // Reactive properties based on current step
  $: stepProperties =
    isMultiStepConnector && stepState.step === "source"
      ? (connector.sourceProperties ?? [])
      : properties;

  // Update form when transitioning to step 2
  $: if (
    isMultiStepConnector &&
    stepState.step === "source" &&
    stepState.connectorConfig
  ) {
    // Initialize form with source properties and default values
    const sourceProperties = connector.sourceProperties ?? [];
    const initialValues = getInitialFormValuesFromProperties(sourceProperties);

    // Merge with stored connector config
    const combinedValues = { ...stepState.connectorConfig, ...initialValues };

    paramsForm.update(() => combinedValues, { taint: false });
  }

  // Determine effective form type
  $: effectiveFormType =
    isMultiStepConnector && stepState.step === "source" ? "source" : formType;

  $: formHeight = ["clickhouse", "snowflake", "salesforce"].includes(
    connector.name ?? "",
  )
    ? "max-h-[38.5rem] min-h-[38.5rem]"
    : "max-h-[34.5rem] min-h-[34.5rem]";

  // Form 1: Individual parameters
  const paramsFormId = `add-data-${connector.name}-form`;
  const properties =
    (isSourceForm
      ? connector.sourceProperties
      : connector.configProperties?.filter(
          (property) => property.key !== "dsn",
        )) ?? [];

  // Filter properties based on connector type
  const filteredParamsProperties = (() => {
    // FIXME: https://linear.app/rilldata/issue/APP-408/support-ducklake-in-the-ui
    if (connector.name === "duckdb") {
      return properties.filter(
        (property) => property.key !== "attach" && property.key !== "mode",
      );
    }
    // For other connectors, filter out noPrompt properties
    return properties.filter((property) => !property.noPrompt);
  })();
  const schema = yup(getValidationSchemaForConnector(connector.name as string));
  const initialFormValues = getInitialFormValuesFromProperties(properties);
  const {
    form: paramsForm,
    errors: paramsErrors,
    enhance: paramsEnhance,
    tainted: paramsTainted,
    submit: paramsSubmit,
    submitting: paramsSubmitting,
  } = superForm(initialFormValues, {
    SPA: true,
    validators: schema,
    onUpdate: handleOnUpdate,
    resetForm: false,
  });
  let paramsError: string | null = null;
  let paramsErrorDetails: string | undefined = undefined;

  // Form 2: DSN
  // SuperForms are not meant to have dynamic schemas, so we use a different form instance for the DSN form
  const hasDsnFormOption =
    isConnectorForm &&
    connector.configProperties?.some((property) => property.key === "dsn") &&
    connector.configProperties?.some((property) => property.key !== "dsn");
  const dsnFormId = `add-data-${connector.name}-dsn-form`;
  const dsnProperties =
    connector.configProperties?.filter((property) => property.key === "dsn") ??
    [];

  const filteredDsnProperties = dsnProperties;
  const dsnYupSchema = yup(dsnValidation);
  const {
    form: dsnForm,
    errors: dsnErrors,
    enhance: dsnEnhance,
    tainted: dsnTainted,
    submit: dsnSubmit,
    submitting: dsnSubmitting,
  } = superForm(defaults(dsnYupSchema), {
    SPA: true,
    validators: dsnYupSchema,
    onUpdate: handleOnUpdate,
    resetForm: false,
  });
  let dsnError: string | null = null;
  let dsnErrorDetails: string | undefined = undefined;

  let clickhouseError: string | null = null;
  let clickhouseErrorDetails: string | undefined = undefined;

  let clickhouseFormId: string = "";
  let clickhouseSubmitting: boolean;
  let clickhouseIsSubmitDisabled: boolean;
  let clickhouseConnectorType: ClickHouseConnectorType = "self-hosted";
  let clickhouseParamsForm;
  let clickhouseDsnForm;

  // Helper function to check if connector only has DSN (no tabs)
  function hasOnlyDsn() {
    return (
      isConnectorForm &&
      connector.configProperties?.some((property) => property.key === "dsn") &&
      !connector.configProperties?.some((property) => property.key !== "dsn")
    );
  }

  // Compute disabled state for the submit button
  $: isSubmitDisabled = (() => {
    if (hasOnlyDsn() || connectionTab === "dsn") {
      // DSN form: check required DSN properties
      for (const property of dsnProperties) {
        if (property.required) {
          const key = String(property.key);
          const value = $dsnForm[key];
          if (isEmpty(value) || $dsnErrors[key]?.length) return true;
        }
      }
      return false;
    } else {
      // Parameters form: check required properties
      // Use stepProperties for multi-step connectors, otherwise use properties
      const propertiesToCheck = isMultiStepConnector
        ? stepProperties
        : properties;

      for (const property of propertiesToCheck) {
        if (property.required) {
          const key = String(property.key);
          const value = $paramsForm[key];

          // Normal validation for all properties
          if (isEmpty(value) || $paramsErrors[key]?.length) return true;
        }
      }
      return false;
    }
  })();

  $: formId = (() => {
    if (hasOnlyDsn() || connectionTab === "dsn") {
      return dsnFormId;
    } else {
      return paramsFormId;
    }
  })();

  $: submitting = (() => {
    if (hasOnlyDsn() || connectionTab === "dsn") {
      return $dsnSubmitting;
    } else {
      return $paramsSubmitting;
    }
  })();

  // Reset errors when form is modified
  $: (() => {
    if (hasOnlyDsn() || connectionTab === "dsn") {
      if ($dsnTainted) dsnError = null;
    } else {
      if ($paramsTainted) paramsError = null;
    }
  })();

  // Clear errors when switching tabs
  $: (() => {
    if (hasDsnFormOption) {
      if (connectionTab === "dsn") {
        paramsError = null;
        paramsErrorDetails = undefined;
      } else {
        dsnError = null;
        dsnErrorDetails = undefined;
      }
    }
  })();

  // Emit the submitting state to the parent
  $: dispatch("submitting", { submitting });

  function getClickHouseYamlPreview(
    values: Record<string, unknown>,
    connectorType: ClickHouseConnectorType,
  ) {
    // Convert connectorType to managed boolean for YAML compatibility
    const managed = connectorType === "rill-managed";

    // Ensure ClickHouse Cloud specific requirements are met in preview
    const previewValues = { ...values, managed } as Record<string, unknown>;
    if (connectorType === "clickhouse-cloud") {
      previewValues.ssl = true;
      previewValues.port = "8443";
    }

    return compileConnectorYAML(connector, previewValues, {
      fieldFilter: (property) => {
        // When in DSN mode, don't filter out noPrompt properties
        // because the DSN field itself might have noPrompt: true
        if (hasOnlyDsn() || connectionTab === "dsn") {
          return true; // Show all DSN properties
        }
        return !property.noPrompt;
      },
      orderedProperties:
        connectionTab === "dsn"
          ? filteredDsnProperties
          : filteredParamsProperties,
    });
  }

  function getConnectorYamlPreview(values: Record<string, unknown>) {
    return compileConnectorYAML(connector, values, {
      fieldFilter: (property) => {
        // When in DSN mode, don't filter out noPrompt properties
        // because the DSN field itself might have noPrompt: true
        if (hasOnlyDsn() || connectionTab === "dsn") {
          return true; // Show all DSN properties
        }
        return !property.noPrompt;
      },
      orderedProperties:
        hasOnlyDsn() || connectionTab === "dsn"
          ? filteredDsnProperties
          : filteredParamsProperties,
    });
  }

  function getSourceYamlPreview(values: Record<string, unknown>) {
    // For multi-step connectors in step 2, filter out connector properties
    let filteredValues = values;
    if (isMultiStepConnector && stepState.step === "source") {
      // Get connector property keys to filter out
      const connectorPropertyKeys = new Set(
        connector.configProperties?.map((prop) => prop.key).filter(Boolean) ||
          [],
      );

      // Filter out connector properties, keeping only source properties and other necessary fields
      filteredValues = Object.fromEntries(
        Object.entries(values).filter(
          ([key]) => !connectorPropertyKeys.has(key),
        ),
      );
    }

    const [rewrittenConnector, rewrittenFormValues] = prepareSourceFormData(
      connector,
      filteredValues,
    );

    // Check if the connector was rewritten to DuckDB
    const isRewrittenToDuckDb = rewrittenConnector.name === "duckdb";

    if (isRewrittenToDuckDb) {
      return compileSourceYAML(rewrittenConnector, rewrittenFormValues);
    } else {
      return getConnectorYamlPreview(rewrittenFormValues);
    }
  }

  $: yamlPreview = (() => {
    // ClickHouse special case
    if (connector.name === "clickhouse") {
      // Reactive form values
      const values =
        connectionTab === "dsn" ? $clickhouseDsnForm : $clickhouseParamsForm;
      return getClickHouseYamlPreview(values, clickhouseConnectorType);
    }

    // Multi-step connector special case - show different preview based on step
    if (isMultiStepConnector) {
      if (stepState.step === "connector") {
        // Step 1: Show connector preview
        return getConnectorYamlPreview($paramsForm);
      } else {
        // Step 2: Show source preview with stored connector config
        const combinedValues = { ...stepState.connectorConfig, ...$paramsForm };
        return getSourceYamlPreview(combinedValues);
      }
    }

    const values =
      hasOnlyDsn() || connectionTab === "dsn" ? $dsnForm : $paramsForm;

    if (isConnectorForm) {
      // Connector form
      return getConnectorYamlPreview(values);
    } else {
      // Source form
      return getSourceYamlPreview(values);
    }
  })();

  // YAML copy handled in YamlPreview component

  function onStringInputChange(event: Event) {
    const target = event.target as HTMLInputElement;
    const { name, value } = target;

    if (name === "path") {
      if ($paramsTainted?.name) return;
      const name = inferSourceName(connector, value);
      if (name)
        paramsForm.update(
          ($form) => {
            $form.name = name;
            return $form;
          },
          { taint: false },
        );
    }
  }

  async function handleOnUpdate<
    T extends Record<string, unknown>,
    M = any,
    In extends Record<string, unknown> = T,
  >(event: {
    form: SuperValidated<T, M, In>;
    formEl: HTMLFormElement;
    cancel: () => void;
    result: Extract<ActionResult, { type: "success" | "failure" }>;
  }) {
    if (!event.form.valid) return;
    const values = event.form.data;

    try {
      let processedValues = values;

      if (isMultiStepConnector && stepState.step === "source") {
        // Step 2: Create source with stored connector config
        await submitAddSourceForm(queryClient, connector, processedValues);
        onClose();
      } else if (isMultiStepConnector && stepState.step === "connector") {
        // Step 1: Create connector and transition to step 2
        await submitAddConnectorForm(queryClient, connector, processedValues);

        setConnectorConfig({});
        setStep("source");
        return; // Don't close the modal, just transition to step 2
      } else if (effectiveFormType === "source") {
        // Regular source form
        await submitAddSourceForm(queryClient, connector, processedValues);
        onClose();
      } else {
        // Regular connector form
        await submitAddConnectorForm(queryClient, connector, processedValues);
        onClose();
      }
    } catch (e) {
      let error: string;
      let details: string | undefined = undefined;

      // Handle different error types
      if (e instanceof Error) {
        error = e.message;
        details = undefined;
      } else if (e?.message && e?.details) {
        error = e.message;
        details = e.details !== e.message ? e.details : undefined;
      } else if (e?.response?.data) {
        const originalMessage = e.response.data.message;
        const humanReadable = humanReadableErrorMessage(
          connector.name,
          e.response.data.code,
          originalMessage,
        );
        error = humanReadable;
        details =
          humanReadable !== originalMessage ? originalMessage : undefined;
      } else if (e?.message) {
        error = e.message;
        details = undefined;
      } else {
        error = "Unknown error";
        details = undefined;
      }

      // Keep error state for each form - match the display logic
      if (hasOnlyDsn() || connectionTab === "dsn") {
        dsnError = error;
        dsnErrorDetails = details;
      } else {
        paramsError = error;
        paramsErrorDetails = details;
      }
    }
  }

  // Handle file upload for credential files
  async function handleFileUpload(file: File): Promise<string> {
    try {
      const content = await file.text();

      // Parse and re-stringify JSON to sanitize whitespace
      const parsedJson = JSON.parse(content);
      const sanitizedJson = JSON.stringify(parsedJson);

      // For BigQuery, try to extract project_id from the credentials JSON
      if (connector.name === "bigquery" && parsedJson.project_id) {
        // Update the project_id field in the form
        paramsForm.update(
          ($form) => {
            $form.project_id = parsedJson.project_id;
            return $form;
          },
          { taint: false },
        );
      }

      return sanitizedJson;
    } catch (error) {
      if (error instanceof SyntaxError) {
        throw new Error(`Invalid JSON file: ${error.message}`);
      }
      throw new Error(`Failed to read file: ${error.message}`);
    }
  }

  // Handle skip button for multi-step connectors
  function handleSkip() {
    if (!isMultiStepConnector || stepState.step !== "connector") return;
    setConnectorConfig({});
    setStep("source");
  }

  function handleBack() {
    if (isMultiStepConnector && stepState.step === "source") {
      // Go back to step 1 (connector configuration)
      setStep("connector");
    } else {
      // Use the original back behavior for non-multi-step or step 1
      onBack();
    }
  }
</script>

<div class="add-data-layout flex flex-col h-full w-full md:flex-row">
  <!-- LEFT SIDE PANEL -->
  <div
    class="add-data-form-panel flex-1 flex flex-col min-w-0 md:pr-0 pr-0 relative"
  >
    <div class="flex flex-col flex-grow {formHeight} overflow-y-auto p-6">
      {#if connector.name === "clickhouse"}
        <AddClickHouseForm
          {connector}
          {onClose}
          setError={(error, details) => {
            clickhouseError = error;
            clickhouseErrorDetails = details;
          }}
          bind:formId={clickhouseFormId}
          bind:submitting={clickhouseSubmitting}
          bind:isSubmitDisabled={clickhouseIsSubmitDisabled}
          bind:connectorType={clickhouseConnectorType}
          bind:connectionTab
          bind:paramsForm={clickhouseParamsForm}
          bind:dsnForm={clickhouseDsnForm}
          on:submitting
        />
      {:else if hasDsnFormOption}
        <Tabs
          bind:value={connectionTab}
          options={CONNECTION_TAB_OPTIONS}
          disableMarginTop
        >
          <TabsContent value="parameters">
            <form
              id={paramsFormId}
              class="pb-5 flex-grow overflow-y-auto"
              use:paramsEnhance
              on:submit|preventDefault={paramsSubmit}
            >
              <FormRenderer
                properties={filteredParamsProperties}
                form={paramsForm}
                errors={$paramsErrors}
                {onStringInputChange}
                uploadFile={handleFileUpload}
              />
            </form>
          </TabsContent>
          <TabsContent value="dsn">
            <form
              id={dsnFormId}
              class="pb-5 flex-grow overflow-y-auto"
              use:dsnEnhance
              on:submit|preventDefault={dsnSubmit}
            >
              <FormRenderer
                properties={filteredDsnProperties}
                form={dsnForm}
                errors={$dsnErrors}
                {onStringInputChange}
                uploadFile={handleFileUpload}
              />
            </form>
          </TabsContent>
        </Tabs>
      {:else if isConnectorForm && connector.configProperties?.some((property) => property.key === "dsn")}
        <!-- Connector with only DSN - show DSN form directly -->
        <form
          id={dsnFormId}
          class="pb-5 flex-grow overflow-y-auto"
          use:dsnEnhance
          on:submit|preventDefault={dsnSubmit}
        >
          <FormRenderer
            properties={filteredDsnProperties}
            form={dsnForm}
            errors={$dsnErrors}
            {onStringInputChange}
            uploadFile={handleFileUpload}
          />
        </form>
      {:else if isMultiStepConnector}
        {#if stepState.step === "connector"}
          <!-- GCS Step 1: Connector configuration -->
          <form
            id={paramsFormId}
            class="pb-5 flex-grow overflow-y-auto"
            use:paramsEnhance
            on:submit|preventDefault={paramsSubmit}
          >
            <GCSMultiStepForm
              properties={filteredParamsProperties}
              {paramsForm}
              paramsErrors={$paramsErrors}
              {onStringInputChange}
              {handleFileUpload}
            />
          </form>
        {:else}
          <!-- GCS Step 2: Source configuration -->
          <form
            id={paramsFormId}
            class="pb-5 flex-grow overflow-y-auto"
            use:paramsEnhance
            on:submit|preventDefault={paramsSubmit}
          >
            <FormRenderer
              properties={stepProperties}
              form={paramsForm}
              errors={$paramsErrors}
              {onStringInputChange}
              uploadFile={handleFileUpload}
            />
          </form>
        {/if}
      {:else}
        <form
          id={paramsFormId}
          class="pb-5 flex-grow overflow-y-auto"
          use:paramsEnhance
          on:submit|preventDefault={paramsSubmit}
        >
          <FormRenderer
            properties={filteredParamsProperties}
            form={paramsForm}
            errors={$paramsErrors}
            {onStringInputChange}
            uploadFile={handleFileUpload}
          />
        </form>
      {/if}
    </div>

    <!-- LEFT FOOTER -->
    <div
      class="w-full bg-white border-t border-gray-200 p-6 flex justify-between gap-2"
    >
      <Button onClick={handleBack} type="secondary">Back</Button>

      <div class="flex gap-2">
        {#if isMultiStepConnector && stepState.step === "connector"}
          <Button onClick={handleSkip} type="secondary">Skip</Button>
        {/if}

        <Button
          disabled={connector.name === "clickhouse"
            ? clickhouseSubmitting || clickhouseIsSubmitDisabled
            : submitting || isSubmitDisabled}
          loading={connector.name === "clickhouse"
            ? clickhouseSubmitting
            : submitting}
          loadingCopy={connector.name === "clickhouse"
            ? "Connecting..."
            : "Testing connection..."}
          form={connector.name === "clickhouse" ? clickhouseFormId : formId}
          submitForm
          type="primary"
        >
          {#if connector.name === "clickhouse"}
            {#if clickhouseConnectorType === "rill-managed"}
              {#if clickhouseSubmitting}
                Connecting...
              {:else}
                Connect
              {/if}
            {:else if clickhouseSubmitting}
              Testing connection...
            {:else}
              Test and Connect
            {/if}
          {:else if isConnectorForm}
            {#if isMultiStepConnector && stepState.step === "connector"}
              {#if submitting}
                Testing connection...
              {:else}
                Test and Connect
              {/if}
            {:else if isMultiStepConnector && stepState.step === "source"}
              {#if submitting}
                Creating model...
              {:else}
                Test and Add data
              {/if}
            {:else if submitting}
              Testing connection...
            {:else}
              Test and Connect
            {/if}
          {:else}
            Test and Add data
          {/if}
        </Button>
      </div>
    </div>
  </div>

  <!-- RIGHT SIDE PANEL -->
  <div
    class="add-data-side-panel flex flex-col gap-6 p-6 bg-[#FAFAFA] w-full max-w-full border-l-0 border-t mt-6 pl-0 pt-6 md:w-96 md:min-w-[320px] md:max-w-[400px] md:border-l md:border-t-0 md:mt-0 md:pl-6"
  >
    {#if dsnError || paramsError || clickhouseError}
      <SubmissionError
        message={clickhouseError ??
          (hasOnlyDsn() || connectionTab === "dsn" ? dsnError : paramsError) ??
          ""}
        details={clickhouseErrorDetails ??
          (hasOnlyDsn() || connectionTab === "dsn"
            ? dsnErrorDetails
            : paramsErrorDetails) ??
          ""}
      />
    {/if}

    <YamlPreview
      title={isMultiStepConnector
        ? stepState.step === "connector"
          ? "Connector preview"
          : "Model preview"
        : isSourceForm
          ? "Model preview"
          : "Connector preview"}
      yaml={yamlPreview}
    />

    <NeedHelpText {connector} />
  </div>
</div>
