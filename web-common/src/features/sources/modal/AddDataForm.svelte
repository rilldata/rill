<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";

  import SubmissionError from "@rilldata/web-common/components/forms/SubmissionError.svelte";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { type V1ConnectorDriver } from "@rilldata/web-common/runtime-client";
  import type { ActionResult } from "@sveltejs/kit";
  import type { SuperValidated } from "sveltekit-superforms";

  import type { AddDataFormType, ConnectorType } from "./types";
  import AddClickHouseForm from "./AddClickHouseForm.svelte";
  import MultiStepConnectorFlow from "./MultiStepConnectorFlow.svelte";
  import NeedHelpText from "./NeedHelpText.svelte";
  import Tabs from "@rilldata/web-common/components/forms/Tabs.svelte";
  import { TabsContent } from "@rilldata/web-common/components/tabs";
  import { hasOnlyDsn, isEmpty } from "./utils";
  import {
    CONNECTION_TAB_OPTIONS,
    type ClickHouseConnectorType,
  } from "./constants";

  import { connectorStepStore } from "./connectorStepStore";
  import FormRenderer from "./FormRenderer.svelte";
  import YamlPreview from "./YamlPreview.svelte";

  import { AddDataFormManager } from "./AddDataFormManager";
  import AddDataFormSection from "./AddDataFormSection.svelte";
  import { get } from "svelte/store";

  export let connector: V1ConnectorDriver;
  export let formType: AddDataFormType;
  export let isSubmitting: boolean;
  export let onBack: () => void;
  export let onClose: () => void;

  let saveAnyway = false;
  let showSaveAnyway = false;
  let connectionTab: ConnectorType = "parameters";

  // Wire manager-provided onUpdate after declaration below
  let handleOnUpdate: <
    T extends Record<string, unknown>,
    M = any,
    In extends Record<string, unknown> = T,
  >(event: {
    form: SuperValidated<T, M, In>;
    formEl: HTMLFormElement;
    cancel: () => void;
    result: Extract<ActionResult, { type: "success" | "failure" }>;
  }) => Promise<void> = async (_event) => {};

  const formManager = new AddDataFormManager({
    connector,
    formType,
    onParamsUpdate: (e: any) => handleOnUpdate(e),
    onDsnUpdate: (e: any) => handleOnUpdate(e),
    getSelectedAuthMethod: () =>
      get(connectorStepStore).selectedAuthMethod ?? undefined,
  });

  const isMultiStepConnector = formManager.isMultiStepConnector;
  const isSourceForm = formManager.isSourceForm;
  const isConnectorForm = formManager.isConnectorForm;
  const onlyDsn = hasOnlyDsn(connector, isConnectorForm);
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

  // Form 1: Individual parameters
  const paramsFormId = formManager.paramsFormId;
  const properties = formManager.properties;
  const filteredParamsProperties = formManager.filteredParamsProperties;
  let multiStepFormId = paramsFormId;
  const {
    form: paramsForm,
    errors: paramsErrors,
    enhance: paramsEnhance,
    tainted: paramsTainted,
    submit: paramsSubmit,
    submitting: paramsSubmitting,
  } = formManager.params;
  let paramsError: string | null = null;
  let paramsErrorDetails: string | undefined = undefined;

  // Form 2: DSN
  // SuperForms are not meant to have dynamic schemas, so we use a different form instance for the DSN form
  const hasDsnFormOption = formManager.hasDsnFormOption;
  const dsnFormId = formManager.dsnFormId;
  const dsnProperties = formManager.dsnProperties;
  const filteredDsnProperties = formManager.filteredDsnProperties;
  const {
    form: dsnForm,
    errors: dsnErrors,
    enhance: dsnEnhance,
    tainted: dsnTainted,
    submit: dsnSubmit,
    submitting: dsnSubmitting,
  } = formManager.dsn;
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
  let clickhouseShowSaveAnyway: boolean = false;

  // Hide Save Anyway once we advance to the model step in multi-step flows.
  $: if (isMultiStepConnector && stepState.step === "source") {
    showSaveAnyway = false;
  }

  $: isSubmitDisabled = (() => {
    if (isMultiStepConnector) {
      return multiStepSubmitDisabled;
    }

    if (onlyDsn || connectionTab === "dsn") {
      // DSN form: check required DSN properties
      for (const property of dsnProperties) {
        const key = String(property.key);
        const value = $dsnForm[key];
        // DSN should be present even if not marked required in metadata
        const mustBePresent = property.required || key === "dsn";
        if (
          mustBePresent &&
          (isEmpty(value) || /**/ ($dsnErrors[key] as any)?.length)
        ) {
          return true;
        }
      }
      return false;
    } else {
      // Parameters form: check required properties
      for (const property of properties) {
        if (property.required) {
          const key = String(property.key);
          const value = $paramsForm[key];

          // Normal validation for all properties
          if (isEmpty(value) || /**/ ($paramsErrors[key] as any)?.length)
            return true;
        }
      }
      return false;
    }
  })();

  $: formId = isMultiStepConnector
    ? multiStepFormId || formManager.getActiveFormId({ connectionTab, onlyDsn })
    : formManager.getActiveFormId({ connectionTab, onlyDsn });

  $: submitting = (() => {
    if (onlyDsn || connectionTab === "dsn") {
      return $dsnSubmitting;
    } else {
      return $paramsSubmitting;
    }
  })();

  $: primaryButtonLabel = isMultiStepConnector
    ? multiStepButtonLabel
    : formManager.getPrimaryButtonLabel({
        isConnectorForm,
        step: stepState.step,
        submitting,
        clickhouseConnectorType,
        clickhouseSubmitting,
        selectedAuthMethod: activeAuthMethod ?? undefined,
      });

  $: primaryLoadingCopy = (() => {
    if (connector.name === "clickhouse") return "Connecting...";
    if (isMultiStepConnector) return multiStepLoadingCopy;
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

  // Reset errors when form is modified
  $: (() => {
    if (onlyDsn || connectionTab === "dsn") {
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

  async function handleSaveAnyway() {
    // Save Anyway should only work for connector forms
    if (!isConnectorForm) {
      return;
    }

    // For other connectors, use manager helper
    saveAnyway = true;
    const values =
      connector.name === "clickhouse"
        ? connectionTab === "dsn"
          ? $clickhouseDsnForm
          : $clickhouseParamsForm
        : onlyDsn || connectionTab === "dsn"
          ? $dsnForm
          : $paramsForm;
    if (connector.name === "clickhouse") {
      clickhouseSubmitting = true;
    }
    const result = await formManager.saveConnectorAnyway({
      queryClient,
      values,
      clickhouseConnectorType,
    });
    if (result.ok) {
      onClose();
    } else {
      if (connector.name === "clickhouse") {
        if (connectionTab === "dsn") {
          dsnError = result.message;
          dsnErrorDetails = result.details;
        } else {
          paramsError = result.message;
          paramsErrorDetails = result.details;
        }
      } else if (onlyDsn || connectionTab === "dsn") {
        dsnError = result.message;
        dsnErrorDetails = result.details;
      } else {
        paramsError = result.message;
        paramsErrorDetails = result.details;
      }
    }
    saveAnyway = false;
    if (connector.name === "clickhouse") {
      clickhouseSubmitting = false;
    }
  }

  $: yamlPreview = formManager.computeYamlPreview({
    connectionTab,
    onlyDsn,
    filteredParamsProperties,
    filteredDsnProperties,
    stepState,
    isMultiStepConnector,
    isConnectorForm,
    paramsFormValues: $paramsForm,
    dsnFormValues: $dsnForm,
    clickhouseConnectorType,
    clickhouseParamsValues: $clickhouseParamsForm,
    clickhouseDsnValues: $clickhouseDsnForm,
  });
  $: isClickhouse = connector.name === "clickhouse";
  $: shouldShowSaveAnywayButton =
    isConnectorForm && (showSaveAnyway || clickhouseShowSaveAnyway);
  $: saveAnywayLoading = isClickhouse
    ? clickhouseSubmitting && saveAnyway
    : submitting && saveAnyway;

  handleOnUpdate = formManager.makeOnUpdate({
    onClose,
    queryClient,
    getConnectionTab: () => connectionTab,
    getSelectedAuthMethod: () => activeAuthMethod || undefined,
    setParamsError: (message: string | null, details?: string) => {
      paramsError = message;
      paramsErrorDetails = details;
    },
    setDsnError: (message: string | null, details?: string) => {
      dsnError = message;
      dsnErrorDetails = details;
    },
    setShowSaveAnyway: (value: boolean) => {
      showSaveAnyway = value;
    },
  });

  async function handleFileUpload(file: File): Promise<string> {
    return formManager.handleFileUpload(file);
  }

  function onStringInputChange(event: Event) {
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
      {#if connector.name === "clickhouse"}
        <AddClickHouseForm
          {connector}
          {onClose}
          setError={(error, details) => {
            clickhouseError = error;
            clickhouseErrorDetails = details;
          }}
          bind:formId={clickhouseFormId}
          bind:isSubmitting={clickhouseSubmitting}
          bind:isSubmitDisabled={clickhouseIsSubmitDisabled}
          bind:connectorType={clickhouseConnectorType}
          bind:connectionTab
          bind:paramsForm={clickhouseParamsForm}
          bind:dsnForm={clickhouseDsnForm}
          bind:showSaveAnyway={clickhouseShowSaveAnyway}
        />
      {:else if hasDsnFormOption}
        <Tabs
          bind:value={connectionTab}
          options={CONNECTION_TAB_OPTIONS}
          disableMarginTop
        >
          <TabsContent value="parameters">
            <AddDataFormSection
              id={paramsFormId}
              enhance={paramsEnhance}
              onSubmit={paramsSubmit}
            >
              <FormRenderer
                properties={filteredParamsProperties}
                form={paramsForm}
                errors={$paramsErrors}
                {onStringInputChange}
                uploadFile={handleFileUpload}
              />
            </AddDataFormSection>
          </TabsContent>
          <TabsContent value="dsn">
            <AddDataFormSection
              id={dsnFormId}
              enhance={dsnEnhance}
              onSubmit={dsnSubmit}
            >
              <FormRenderer
                properties={filteredDsnProperties}
                form={dsnForm}
                errors={$dsnErrors}
                {onStringInputChange}
                uploadFile={handleFileUpload}
              />
            </AddDataFormSection>
          </TabsContent>
        </Tabs>
      {:else if isConnectorForm && connector.configProperties?.some((property) => property.key === "dsn")}
        <!-- Connector with only DSN - show DSN form directly -->
        <AddDataFormSection
          id={dsnFormId}
          enhance={dsnEnhance}
          onSubmit={dsnSubmit}
        >
          <FormRenderer
            properties={filteredDsnProperties}
            form={dsnForm}
            errors={$dsnErrors}
            {onStringInputChange}
            uploadFile={handleFileUpload}
          />
        </AddDataFormSection>
      {:else if isMultiStepConnector}
        <MultiStepConnectorFlow
          {connector}
          {formManager}
          {paramsForm}
          {paramsErrors}
          {paramsEnhance}
          {paramsSubmit}
          {paramsFormId}
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
      {:else}
        <AddDataFormSection
          id={paramsFormId}
          enhance={paramsEnhance}
          onSubmit={paramsSubmit}
        >
          <FormRenderer
            properties={filteredParamsProperties}
            form={paramsForm}
            errors={$paramsErrors}
            {onStringInputChange}
            uploadFile={handleFileUpload}
          />
        </AddDataFormSection>
      {/if}
    </div>

    <!-- LEFT FOOTER -->
    <div
      class="w-full bg-surface border-t border-gray-200 p-6 flex justify-between gap-2"
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
          disabled={connector.name === "clickhouse"
            ? clickhouseSubmitting || clickhouseIsSubmitDisabled
            : submitting || isSubmitDisabled}
          loading={connector.name === "clickhouse"
            ? clickhouseSubmitting
            : submitting}
          loadingCopy={primaryLoadingCopy}
          form={connector.name === "clickhouse" ? clickhouseFormId : formId}
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
      {#if dsnError || paramsError || clickhouseError}
        <SubmissionError
          message={clickhouseError ??
            (onlyDsn || connectionTab === "dsn" ? dsnError : paramsError) ??
            ""}
          details={clickhouseErrorDetails ??
            (onlyDsn || connectionTab === "dsn"
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
