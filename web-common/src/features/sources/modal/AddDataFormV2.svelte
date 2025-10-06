<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import { get as getStore, type Readable, readable } from "svelte/store";
  import type { V1ConnectorDriver } from "@rilldata/web-common/runtime-client";
  import type { AddDataFormType, ConnectorType } from "./types";
  import type { ClickHouseConnectorType } from "./constants";

  import { FormStateManager } from "./form-state-manager";
  import { FormErrorManager } from "./form-error-manager";
  import { getConnectorHandler } from "./connector-handlers";
  import {
    createConnectorForms,
    createClickHouseForms,
    getCurrentForm,
    getCurrentFormId,
    isSubmitDisabled,
  } from "./form-factory";
  import FormRendererV2 from "./FormRendererV2.svelte";
  import LeftFooter from "./LeftFooter.svelte";
  import RightSidePanel from "./RightSidePanel.svelte";

  // Import utilities
  import { inferSourceName } from "../sourceUtils";
  import { humanReadableErrorMessage } from "../errors/errors";
  import "./connector-handler-registry"; // Register handlers

  const dispatch = createEventDispatcher();

  export let connector: V1ConnectorDriver;
  export let formType: AddDataFormType;
  export let onBack: () => void;
  export let onClose: () => void;

  const stateManager = new FormStateManager(connector, formType);
  const errorManager = new FormErrorManager();
  const connectorHandler = getConnectorHandler(connector);

  // Get the store from the state manager
  const stateStore = stateManager.state;

  // Subscribe to the store and destructure the state
  let state: any;
  let connectionTab: ConnectorType = "parameters";
  let clickhouseConnectorType: ClickHouseConnectorType = "self-hosted";
  let submitting = false;
  let copied = false;
  $: state = $stateStore;
  $: connectionTab = state.connectionTab;
  $: clickhouseConnectorType = state.clickhouseConnectorType;
  $: submitting = state.submitting;
  $: copied = state.copied;

  // Form states
  let forms: {
    paramsForm: any;
    dsnForm: any;
  } | null = null;

  let clickhouseForms: {
    paramsForm: any;
    dsnForm: any;
  } | null = null;

  // YAML preview state
  let yamlPreview = "";
  const emptyReadable = readable<any>({});
  let activeFormStore: Readable<any> = emptyReadable;

  // Initialize forms during component initialization (required by superForm)
  initializeForms();

  function initializeForms() {
    const onSubmit = async (values: Record<string, unknown>) => {
      try {
        await connectorHandler.handleSubmit(connector, formType, values);
        onClose();
      } catch (error) {
        const formId = getCurrentFormId(connector, connectionTab);
        errorManager.setApiError(
          formId,
          error,
          connector.name || "",
          (connectorName: string, code: string, message: string) =>
            humanReadableErrorMessage(connectorName, Number(code), message),
        );
      }
    };

    if (stateManager.isClickhouseConnector) {
      clickhouseForms = createClickHouseForms(connector, formType, onSubmit);
    } else {
      forms = createConnectorForms(connector, formType, onSubmit);
    }
  }

  // Handle string input changes (for name inference)
  function onStringInputChange(event: Event) {
    const target = event.target as HTMLInputElement;
    const { name, value } = target;

    if (name === "path") {
      const currentForm = getCurrentForm(
        connector,
        connectionTab,
        forms || { paramsForm: null, dsnForm: null },
      );
      if (!currentForm || currentForm.tainted?.name) return;

      const nameVal = inferSourceName(connector, value);
      if (nameVal) {
        (currentForm.form as any).update(
          ($form: any) => {
            $form.name = nameVal;
            return $form;
          },
          { taint: false },
        );
      }
    }
  }

  // Handle ClickHouse string input changes
  function onClickHouseStringInputChange(event: Event) {
    const target = event.target as HTMLInputElement;
    const { name, value } = target;

    if (name === "path" && clickhouseForms) {
      if (clickhouseForms.paramsForm.tainted?.name) return;

      const nameVal = inferSourceName(connector, value);
      if (nameVal) {
        (clickhouseForms.paramsForm.form as any).update(
          ($form: any) => {
            $form.name = nameVal;
            return $form;
          },
          { taint: false },
        );
      }
    }
  }

  // Track active form store and update YAML preview reactively
  $: activeFormStore =
    stateManager.isClickhouseConnector && clickhouseForms
      ? connectionTab === "dsn"
        ? (clickhouseForms.dsnForm.form as unknown as Readable<any>)
        : (clickhouseForms.paramsForm.form as unknown as Readable<any>)
      : forms
        ? (getCurrentForm(connector, connectionTab, forms)
            ?.form as unknown as Readable<any>) || emptyReadable
        : emptyReadable;

  $: yamlPreview = connectorHandler.getYamlPreview(
    connector,
    formType,
    $activeFormStore as any,
  );

  // Copy YAML preview
  function copyYamlPreview() {
    navigator.clipboard.writeText(yamlPreview);
    stateManager.setCopied(true);
    setTimeout(() => {
      stateManager.setCopied(false);
    }, 2000);
  }

  // Emit submitting state to parent
  $: dispatch("submitting", { submitting });

  // Computed properties for footer
  $: footerDisabled =
    submitting ||
    isSubmitDisabled(
      connector,
      connectionTab,
      forms || { paramsForm: null, dsnForm: null },
    );
  $: footerLoading = submitting;
  $: footerLoadingCopy = stateManager.getLoadingCopyText();
  $: footerFormId = getCurrentFormId(connector, connectionTab);
  $: footerSubmitButtonText = stateManager.getSubmitButtonText();
</script>

<div class="add-data-layout flex flex-col h-full w-full md:flex-row">
  <!-- LEFT SIDE PANEL -->
  <div
    class="add-data-form-panel flex-1 flex flex-col min-w-0 md:pr-0 pr-0 relative"
  >
    <div
      class="flex flex-col flex-grow {[
        'clickhouse',
        'snowflake',
        'salesforce',
      ].includes(connector.name ?? '')
        ? 'max-h-[38.5rem] min-h-[38.5rem]'
        : 'max-h-[34.5rem] min-h-[34.5rem]'} overflow-y-auto p-6"
    >
      {#if forms || clickhouseForms}
        <FormRendererV2
          {connector}
          {formType}
          bind:connectionTab
          {connectorHandler}
          {forms}
          {onStringInputChange}
          bind:clickhouseConnectorType
          {clickhouseForms}
          {onClickHouseStringInputChange}
        />
      {/if}
    </div>

    <LeftFooter
      {onBack}
      disabled={footerDisabled}
      loading={footerLoading}
      loadingCopy={footerLoadingCopy}
      formId={footerFormId}
      submitButtonText={footerSubmitButtonText}
    />
  </div>

  <RightSidePanel
    {connector}
    isSourceForm={stateManager.isSourceForm}
    dsnError={errorManager.getError(
      connectorHandler.getFormId(connector, "dsn"),
    )?.message || null}
    paramsError={errorManager.getError(
      connectorHandler.getFormId(connector, "params"),
    )?.message || null}
    clickhouseError={stateManager.isClickhouseConnector
      ? errorManager.getError(connectorHandler.getFormId(connector, "params"))
          ?.message || null
      : null}
    dsnErrorDetails={errorManager.getError(
      connectorHandler.getFormId(connector, "dsn"),
    )?.details || undefined}
    paramsErrorDetails={errorManager.getError(
      connectorHandler.getFormId(connector, "params"),
    )?.details || undefined}
    clickhouseErrorDetails={stateManager.isClickhouseConnector
      ? errorManager.getError(connectorHandler.getFormId(connector, "params"))
          ?.details || undefined
      : undefined}
    hasOnlyDsn={stateManager.hasOnlyDsn}
    {connectionTab}
    {yamlPreview}
    {copied}
    {copyYamlPreview}
  />
</div>
