<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";

  import SubmissionError from "@rilldata/web-common/components/forms/SubmissionError.svelte";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import {
    type ConnectorDriverProperty,
    type V1ConnectorDriver,
  } from "@rilldata/web-common/runtime-client";
  import type { ActionResult } from "@sveltejs/kit";
  import type { SuperValidated } from "sveltekit-superforms";

  import type { AddDataFormType, ConnectorType } from "./types";
  import AddClickHouseForm from "./AddClickHouseForm.svelte";
  import ConnectorForm from "./ConnectorForm.svelte";
  import NeedHelpText from "./NeedHelpText.svelte";
  import Tabs from "@rilldata/web-common/components/forms/Tabs.svelte";
  import { TabsContent } from "@rilldata/web-common/components/tabs";
  import { hasOnlyDsn, isEmpty } from "./utils";
  import {
    CONNECTION_TAB_OPTIONS,
    type ClickHouseConnectorType,
    FORM_HEIGHT_DEFAULT,
    MULTI_STEP_CONNECTORS,
  } from "./constants";

  import FormRenderer from "./FormRenderer.svelte";
  import YamlPreview from "./YamlPreview.svelte";

  import { AddDataFormManager } from "./AddDataFormManager";
  import AddDataFormSection from "./AddDataFormSection.svelte";

  export let connector: V1ConnectorDriver;
  export let formType: AddDataFormType;
  export let connectorInstanceName: string | null = null;
  export let isSubmitting: boolean;
  export let onBack: () => void;
  export let onClose: () => void;

  const isMultiStepConnector =
    formType === "connector" &&
    MULTI_STEP_CONNECTORS.includes(connector.name ?? "");

  let saveAnyway = false;
  let showSaveAnyway = false;
  let connectionTab: ConnectorType = "parameters";
  let formHeight = FORM_HEIGHT_DEFAULT;

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

  let formManager: AddDataFormManager | null = null;

  $: if (!isMultiStepConnector) {
    formManager = new AddDataFormManager({
      connector,
      formType,
      connectorInstanceName,
      onParamsUpdate: (e: any) => handleOnUpdate(e),
      onDsnUpdate: (e: any) => handleOnUpdate(e),
      getSelectedAuthMethod: () => undefined,
    });
    formHeight = formManager.formHeight;
  }

  let isSourceForm = formType === "source";
  let isConnectorForm = formType === "connector";
  let onlyDsn = false;
  let activeAuthMethod: string | null = null;
  let prevAuthMethod: string | null = null;
  let multiStepSubmitDisabled = false;
  let multiStepButtonLabel = "";
  let multiStepLoadingCopy = "";
  let primaryButtonLabel = "";
  let primaryLoadingCopy = "";
  let shouldShowSkipLink = false;
  let multiStepYamlPreview = "";
  let multiStepYamlPreviewTitle = "Connector preview";
  let multiStepSubmitting = false;
  let multiStepShowSaveAnyway = false;
  let multiStepSaveAnywayLoading = false;
  let multiStepSaveAnywayHandler: () => Promise<void> = async () => {};
  let multiStepHandleBack: () => void = () => onBack();
  let multiStepHandleSkip: () => void = () => {};

  // Form 1: Individual parameters (non-multi-step)
  let paramsFormId = "";
  let properties: ConnectorDriverProperty[] = [];
  let filteredParamsProperties: ConnectorDriverProperty[] = [];
  let multiStepFormId = "";
  let paramsForm: any = null;
  let paramsErrors: any = null;
  let paramsEnhance: any = null;
  let paramsTainted: any = null;
  let paramsSubmit: any = null;
  let paramsSubmitting: any = null;
  let paramsError: string | null = null;
  let paramsErrorDetails: string | undefined = undefined;

  // Form 2: DSN (non-multi-step)
  let hasDsnFormOption = false;
  let dsnFormId = "";
  let dsnProperties: ConnectorDriverProperty[] = [];
  let filteredDsnProperties: ConnectorDriverProperty[] = [];
  let dsnForm: any = null;
  let dsnErrors: any = null;
  let dsnEnhance: any = null;
  let dsnTainted: any = null;
  let dsnSubmit: any = null;
  let dsnSubmitting: any = null;
  let dsnError: string | null = null;
  let dsnErrorDetails: string | undefined = undefined;

  $: if (formManager) {
    paramsFormId = formManager.paramsFormId;
    properties = formManager.properties;
    filteredParamsProperties = formManager.filteredParamsProperties;
    multiStepFormId = paramsFormId;
    ({
      form: paramsForm,
      errors: paramsErrors,
      enhance: paramsEnhance,
      tainted: paramsTainted,
      submit: paramsSubmit,
      submitting: paramsSubmitting,
    } = formManager.params);

    hasDsnFormOption = formManager.hasDsnFormOption;
    dsnFormId = formManager.dsnFormId;
    dsnProperties = formManager.dsnProperties;
    filteredDsnProperties = formManager.filteredDsnProperties;
    ({
      form: dsnForm,
      errors: dsnErrors,
      enhance: dsnEnhance,
      tainted: dsnTainted,
      submit: dsnSubmit,
      submitting: dsnSubmitting,
    } = formManager.dsn);

    isSourceForm = formManager.isSourceForm;
    isConnectorForm = formManager.isConnectorForm;
    hasDsnFormOption = formManager.hasDsnFormOption;
    onlyDsn = hasOnlyDsn(connector, isConnectorForm);
  } else {
    isSourceForm = formType === "source";
    isConnectorForm = formType === "connector";
    hasDsnFormOption = false;
    onlyDsn = false;
  }

  let clickhouseError: string | null = null;
  let clickhouseErrorDetails: string | undefined = undefined;
  let clickhouseFormId: string = "";
  let clickhouseSubmitting: boolean;
  let clickhouseIsSubmitDisabled: boolean;
  let clickhouseConnectorType: ClickHouseConnectorType = "self-hosted";
  let clickhouseParamsForm;
  let clickhouseDsnForm;
  let clickhouseShowSaveAnyway: boolean = false;

  $: isSubmitDisabled = (() => {
    if (isMultiStepConnector) {
      return multiStepSubmitDisabled;
    }

    if (!formManager) return true;

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
    ? multiStepFormId
    : (formManager?.getActiveFormId({ connectionTab, onlyDsn }) ?? "");

  $: submitting = (() => {
    if (isMultiStepConnector) return multiStepSubmitting;
    if (!formManager) return false;
    if (onlyDsn || connectionTab === "dsn") {
      return $dsnSubmitting;
    } else {
      return $paramsSubmitting;
    }
  })();

  $: primaryButtonLabel = isMultiStepConnector
    ? multiStepButtonLabel
    : formManager
      ? formManager.getPrimaryButtonLabel({
          isConnectorForm,
          step: isSourceForm ? "source" : "connector",
          submitting,
          clickhouseConnectorType,
          clickhouseSubmitting,
          selectedAuthMethod: activeAuthMethod ?? undefined,
        })
      : "";

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

  $: ctaDisabled =
    connector.name === "clickhouse"
      ? clickhouseSubmitting || clickhouseIsSubmitDisabled
      : isMultiStepConnector
        ? multiStepSubmitting || multiStepSubmitDisabled
        : submitting || isSubmitDisabled;
  $: ctaLoading =
    connector.name === "clickhouse"
      ? clickhouseSubmitting
      : isMultiStepConnector
        ? multiStepSubmitting
        : submitting;
  $: ctaLoadingCopy = isMultiStepConnector
    ? multiStepLoadingCopy
    : primaryLoadingCopy;
  $: ctaLabel = isMultiStepConnector
    ? multiStepButtonLabel
    : primaryButtonLabel;
  $: ctaFormId =
    connector.name === "clickhouse"
      ? clickhouseFormId
      : isMultiStepConnector
        ? multiStepFormId
        : formId;

  $: effectiveYaml = isMultiStepConnector ? multiStepYamlPreview : yamlPreview;
  $: effectiveYamlTitle = isMultiStepConnector
    ? multiStepYamlPreviewTitle
    : isSourceForm
      ? "Model preview"
      : "Connector preview";

  // Reset errors when form is modified (non-multi-step paths)
  $: (() => {
    if (isMultiStepConnector || !formManager) return;
    if (onlyDsn || connectionTab === "dsn") {
      if ($dsnTainted) dsnError = null;
    } else {
      if ($paramsTainted) paramsError = null;
    }
  })();

  // Clear errors when switching tabs (non-multi-step paths)
  $: (() => {
    if (isMultiStepConnector || !formManager) return;
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

    // Multi-step connectors delegate to the container handler
    if (isMultiStepConnector) {
      await multiStepSaveAnywayHandler();
      return;
    }

    if (!formManager) return;

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

  $: yamlPreview = isMultiStepConnector
    ? multiStepYamlPreview
    : formManager
      ? formManager.computeYamlPreview({
          connectionTab,
          onlyDsn,
          filteredParamsProperties,
          filteredDsnProperties,
          stepState: undefined,
          isMultiStepConnector,
          isConnectorForm,
          paramsFormValues: $paramsForm,
          dsnFormValues: $dsnForm,
          clickhouseConnectorType,
          clickhouseParamsValues: $clickhouseParamsForm,
          clickhouseDsnValues: $clickhouseDsnForm,
        })
      : "";
  $: isClickhouse = connector.name === "clickhouse";
  $: shouldShowSaveAnywayButton =
    isConnectorForm &&
    (clickhouseShowSaveAnyway ||
      (isMultiStepConnector ? multiStepShowSaveAnyway : showSaveAnyway));
  $: saveAnywayLoading = isMultiStepConnector
    ? multiStepSaveAnywayLoading
    : isClickhouse
      ? clickhouseSubmitting && saveAnyway
      : submitting && saveAnyway;

  if (formManager) {
    const fm = formManager as AddDataFormManager;
    handleOnUpdate =
      fm.makeOnUpdate({
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
      }) ?? (async () => {});
  } else {
    handleOnUpdate = async () => {};
  }

  async function handleFileUpload(file: File): Promise<string> {
    return formManager ? formManager.handleFileUpload(file) : "";
  }

  function onStringInputChange(event: Event) {
    if (!formManager) return;
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
          bind:isSubmitting={clickhouseSubmitting}
          bind:isSubmitDisabled={clickhouseIsSubmitDisabled}
          bind:connectorType={clickhouseConnectorType}
          bind:connectionTab
          bind:paramsForm={clickhouseParamsForm}
          bind:dsnForm={clickhouseDsnForm}
          bind:showSaveAnyway={clickhouseShowSaveAnyway}
        />
      {:else if isMultiStepConnector}
        {#key connector.name}
          <ConnectorForm
            {connector}
            {connectorInstanceName}
            {onClose}
            {onBack}
            bind:isSubmitDisabled={multiStepSubmitDisabled}
            bind:primaryButtonLabel={multiStepButtonLabel}
            bind:primaryLoadingCopy={multiStepLoadingCopy}
            bind:formId={multiStepFormId}
            bind:yamlPreview={multiStepYamlPreview}
            bind:yamlPreviewTitle={multiStepYamlPreviewTitle}
            bind:isSubmitting={multiStepSubmitting}
            bind:showSaveAnyway={multiStepShowSaveAnyway}
            bind:saveAnywayLoading={multiStepSaveAnywayLoading}
            bind:saveAnywayHandler={multiStepSaveAnywayHandler}
            bind:handleBack={multiStepHandleBack}
            bind:handleSkip={multiStepHandleSkip}
            bind:shouldShowSkipLink
            bind:paramsError
            bind:paramsErrorDetails
          />
        {/key}
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
      class="w-full bg-surface-subtle border-t border-gray-200 p-6 flex justify-between gap-2"
    >
      <Button
        onClick={() =>
          isMultiStepConnector
            ? multiStepHandleBack()
            : formManager?.handleBack(onBack)}
        type="tertiary"
      >
        Back
      </Button>

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
          disabled={ctaDisabled}
          loading={ctaLoading}
          loadingCopy={ctaLoadingCopy}
          form={ctaFormId}
          submitForm
          type="primary"
        >
          {ctaLabel}
        </Button>
      </div>
    </div>
  </div>

  <!-- RIGHT SIDE PANEL -->
  <div
    class="add-data-side-panel flex flex-col gap-6 p-6 bg-surface-subtle w-full max-w-full border-l-0 border-t mt-6 pl-0 pt-6 md:w-96 md:min-w-[320px] md:max-w-[400px] md:border-l md:border-t-0 md:mt-0 md:pl-6 justify-between"
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

      <YamlPreview title={effectiveYamlTitle} yaml={effectiveYaml} />

      {#if shouldShowSkipLink}
        <div class="text-sm leading-normal font-medium text-muted-foreground">
          Already connected? <button
            type="button"
            class="text-sm leading-normal text-primary-500 hover:text-primary-600 font-medium hover:underline break-all"
            on:click={() =>
              isMultiStepConnector
                ? multiStepHandleSkip()
                : formManager?.handleSkip()}
          >
            Import your data
          </button>
        </div>
      {/if}
    </div>

    <NeedHelpText {connector} />
  </div>
</div>
