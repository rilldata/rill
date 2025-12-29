<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";

  import SubmissionError from "@rilldata/web-common/components/forms/SubmissionError.svelte";
  import ConnectorTypeSelector from "@rilldata/web-common/components/forms/ConnectorTypeSelector.svelte";
  import InformationalField from "@rilldata/web-common/components/forms/InformationalField.svelte";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import Checkbox from "@rilldata/web-common/components/forms/Checkbox.svelte";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { type V1ConnectorDriver } from "@rilldata/web-common/runtime-client";
  import type { ActionResult } from "@sveltejs/kit";
  import type { SuperValidated } from "sveltekit-superforms";

  import type { AddDataFormType, ConnectorType } from "./types";
  import MultiStepConnectorFlow from "./MultiStepConnectorFlow.svelte";
  import NeedHelpText from "./NeedHelpText.svelte";
  import Tabs from "@rilldata/web-common/components/forms/Tabs.svelte";
  import { TabsContent } from "@rilldata/web-common/components/tabs";
  import { hasOnlyDsn, isEmpty } from "./utils";
  import {
    CONNECTION_TAB_OPTIONS,
    CONNECTOR_TYPE_OPTIONS,
    type ClickHouseConnectorType,
  } from "./constants";
  import { normalizeErrors } from "../../templates/error-utils";

  import { connectorStepStore } from "./connectorStepStore";
  import FormRenderer from "./FormRenderer.svelte";
  import YamlPreview from "./YamlPreview.svelte";

  import { AddDataFormManager } from "./AddDataFormManager";
  import AddDataFormSection from "./AddDataFormSection.svelte";
  import { get, type Writable } from "svelte/store";
  import {
    ConnectorDriverPropertyType,
    type ConnectorDriverProperty,
  } from "@rilldata/web-common/runtime-client";
  import { getInitialFormValuesFromProperties } from "../sourceUtils";

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
  let clickhouseProperties: ConnectorDriverProperty[] = [];
  let clickhouseFilteredProperties: ConnectorDriverProperty[] = [];
  let clickhouseDsnProperties: ConnectorDriverProperty[] = [];
  let clickhouseShowSaveAnyway: boolean = false;
  const clickhouseInitialValues =
    connector.name === "clickhouse"
      ? getInitialFormValuesFromProperties(connector.configProperties ?? [])
      : {};
  let prevClickhouseConnectorType: ClickHouseConnectorType =
    clickhouseConnectorType;
  const paramsFormStore = paramsForm as unknown as Writable<
    Record<string, any>
  >;
  const dsnFormStore = dsnForm as unknown as Writable<Record<string, any>>;

  // Keep ClickHouse connector type in sync with form state and apply defaults
  $: if (connector.name === "clickhouse") {
    // Always set connector_type on the params form
    paramsForm.update(
      ($form: any) => ({
        ...$form,
        connector_type: clickhouseConnectorType,
      }),
      { taint: false } as any,
    );

    if (
      clickhouseConnectorType === "rill-managed" &&
      Object.keys($paramsForm).length > 1
    ) {
      paramsForm.update(
        () => ({ managed: true, connector_type: "rill-managed" }),
        { taint: false } as any,
      );
      clickhouseError = null;
      clickhouseErrorDetails = undefined;
    } else if (
      prevClickhouseConnectorType === "rill-managed" &&
      clickhouseConnectorType === "self-hosted"
    ) {
      paramsForm.update(
        () => ({ ...clickhouseInitialValues, managed: false }),
        { taint: false } as any,
      );
    } else if (
      prevClickhouseConnectorType !== "clickhouse-cloud" &&
      clickhouseConnectorType === "clickhouse-cloud"
    ) {
      paramsForm.update(
        () => ({
          ...clickhouseInitialValues,
          managed: false,
          port: "8443",
          ssl: true,
        }),
        { taint: false } as any,
      );
    } else if (
      prevClickhouseConnectorType === "clickhouse-cloud" &&
      clickhouseConnectorType === "self-hosted"
    ) {
      paramsForm.update(
        () => ({ ...clickhouseInitialValues, managed: false }),
        { taint: false } as any,
      );
    }
    prevClickhouseConnectorType = clickhouseConnectorType;
  }

  // Force parameters tab for Rill-managed
  $: if (
    connector.name === "clickhouse" &&
    clickhouseConnectorType === "rill-managed"
  ) {
    connectionTab = "parameters";
  }

  // Use manager forms for ClickHouse
  $: if (connector.name === "clickhouse") {
    clickhouseParamsForm = paramsForm;
    clickhouseDsnForm = dsnForm;
    clickhouseFormId = connectionTab === "dsn" ? dsnFormId : paramsFormId;
    clickhouseSubmitting =
      connectionTab === "dsn" ? $dsnSubmitting : $paramsSubmitting;
  }

  $: if (connector.name === "clickhouse") {
    clickhouseShowSaveAnyway = showSaveAnyway;
  }

  // ClickHouse-specific property filtering and disabled state
  $: clickhouseProperties =
    connector.name === "clickhouse"
      ? clickhouseConnectorType === "rill-managed"
        ? (connector.sourceProperties ?? [])
        : (connector.configProperties ?? [])
      : [];

  $: clickhouseFilteredProperties = clickhouseProperties.filter(
    (property) =>
      !property.noPrompt &&
      property.key !== "managed" &&
      (connectionTab !== "dsn" ? property.key !== "dsn" : true),
  );

  $: clickhouseDsnProperties =
    connector.configProperties?.filter((property) => property.key === "dsn") ??
    [];

  $: clickhouseIsSubmitDisabled = (() => {
    if (connector.name !== "clickhouse") return false;
    if (clickhouseConnectorType === "rill-managed") {
      for (const property of clickhouseFilteredProperties) {
        if (property.required) {
          const key = String(property.key);
          const value = $paramsForm[key];
          if (isEmpty(value) || ($paramsErrors[key] as any)?.length)
            return true;
        }
      }
      return false;
    } else if (connectionTab === "dsn") {
      for (const property of clickhouseDsnProperties) {
        if (property.required) {
          const key = String(property.key);
          const value = $dsnForm[key];
          if (isEmpty(value) || ($dsnErrors[key] as any)?.length) return true;
        }
      }
      return false;
    } else {
      for (const property of clickhouseFilteredProperties) {
        if (property.required && property.key !== "managed") {
          const key = String(property.key);
          const value = $paramsForm[key];
          if (isEmpty(value) || ($paramsErrors[key] as any)?.length)
            return true;
        }
      }
      if (clickhouseConnectorType === "clickhouse-cloud") {
        if (!$paramsForm.ssl) return true;
      }
      return false;
    }
  })();

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
        <div class="h-full w-full flex flex-col">
          <div>
            <ConnectorTypeSelector
              bind:value={clickhouseConnectorType}
              options={CONNECTOR_TYPE_OPTIONS}
            />
            {#if clickhouseConnectorType === "rill-managed"}
              <div class="mt-4">
                <InformationalField
                  description="This option uses ClickHouse as an OLAP engine with Rill-managed infrastructure. No additional configuration is required - Rill will handle the setup and management of your ClickHouse instance."
                />
              </div>
            {/if}
          </div>

          {#if clickhouseConnectorType === "self-hosted" || clickhouseConnectorType === "clickhouse-cloud"}
            <Tabs bind:value={connectionTab} options={CONNECTION_TAB_OPTIONS}>
              <TabsContent value="parameters">
                <AddDataFormSection
                  id={paramsFormId}
                  enhance={paramsEnhance}
                  onSubmit={paramsSubmit}
                >
                  {#each clickhouseFilteredProperties as property (property.key)}
                    {@const propertyKey = property.key ?? ""}
                    {@const isPortField = propertyKey === "port"}
                    {@const isSSLField = propertyKey === "ssl"}

                    <div class="py-1.5 first:pt-0 last:pb-0">
                      {#if property.type === ConnectorDriverPropertyType.TYPE_STRING || property.type === ConnectorDriverPropertyType.TYPE_NUMBER}
                        <Input
                          id={propertyKey}
                          label={property.displayName}
                          placeholder={property.placeholder}
                          optional={property.required}
                          secret={property.secret}
                          hint={property.hint}
                          errors={normalizeErrors($paramsErrors[propertyKey])}
                          bind:value={$paramsFormStore[propertyKey]}
                          onInput={(_, e) => onStringInputChange(e)}
                          alwaysShowError
                          options={clickhouseConnectorType ===
                            "clickhouse-cloud" && isPortField
                            ? [
                                { value: "8443", label: "8443 (HTTPS)" },
                                {
                                  value: "9440",
                                  label: "9440 (Native Secure)",
                                },
                              ]
                            : undefined}
                        />
                      {:else if property.type === ConnectorDriverPropertyType.TYPE_BOOLEAN}
                        <Checkbox
                          id={propertyKey}
                          bind:checked={$paramsFormStore[propertyKey]}
                          label={property.displayName}
                          hint={property.hint}
                          optional={clickhouseConnectorType ===
                            "clickhouse-cloud" && isSSLField
                            ? false
                            : !property.required}
                          disabled={clickhouseConnectorType ===
                            "clickhouse-cloud" && isSSLField}
                        />
                      {:else if property.type === ConnectorDriverPropertyType.TYPE_INFORMATIONAL}
                        <InformationalField
                          description={property.description}
                          hint={property.hint}
                          href={property.docsUrl}
                        />
                      {/if}
                    </div>
                  {/each}
                </AddDataFormSection>
              </TabsContent>
              <TabsContent value="dsn">
                <AddDataFormSection
                  id={dsnFormId}
                  enhance={dsnEnhance}
                  onSubmit={dsnSubmit}
                >
                  {#each clickhouseDsnProperties as property (property.key)}
                    {@const propertyKey = property.key ?? ""}
                    <div class="py-1.0 first:pt-0 last:pb-0">
                      <Input
                        id={propertyKey}
                        label={property.displayName}
                        placeholder={property.placeholder}
                        secret={property.secret}
                        hint={property.hint}
                        errors={normalizeErrors($dsnErrors[propertyKey])}
                        bind:value={$dsnFormStore[propertyKey]}
                        alwaysShowError
                      />
                    </div>
                  {/each}
                </AddDataFormSection>
              </TabsContent>
            </Tabs>
          {:else}
            <AddDataFormSection
              id={paramsFormId}
              enhance={paramsEnhance}
              onSubmit={paramsSubmit}
            >
              {#each clickhouseFilteredProperties as property (property.key)}
                {@const propertyKey = property.key ?? ""}
                <div class="py-1.5 first:pt-0 last:pb-0">
                  {#if property.type === ConnectorDriverPropertyType.TYPE_STRING || property.type === ConnectorDriverPropertyType.TYPE_NUMBER}
                    <Input
                      id={propertyKey}
                      label={property.displayName}
                      placeholder={property.placeholder}
                      optional={!property.required}
                      secret={property.secret}
                      hint={property.hint}
                      errors={normalizeErrors($paramsErrors[propertyKey])}
                      bind:value={$paramsFormStore[propertyKey]}
                      onInput={(_, e) => onStringInputChange(e)}
                      alwaysShowError
                    />
                  {:else if property.type === ConnectorDriverPropertyType.TYPE_BOOLEAN}
                    <Checkbox
                      id={propertyKey}
                      bind:checked={$paramsFormStore[propertyKey]}
                      label={property.displayName}
                      hint={property.hint}
                      optional={!property.required}
                    />
                  {:else if property.type === ConnectorDriverPropertyType.TYPE_INFORMATIONAL}
                    <InformationalField
                      description={property.description}
                      hint={property.hint}
                      href={property.docsUrl}
                    />
                  {/if}
                </div>
              {/each}
            </AddDataFormSection>
          {/if}
        </div>
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
