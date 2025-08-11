<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import InformationalField from "@rilldata/web-common/components/forms/InformationalField.svelte";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import SubmissionError from "@rilldata/web-common/components/forms/SubmissionError.svelte";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import {
    ConnectorDriverPropertyType,
    type V1ConnectorDriver,
  } from "@rilldata/web-common/runtime-client";
  import type { ActionResult } from "@sveltejs/kit";
  import { createEventDispatcher } from "svelte";
  import {
    defaults,
    superForm,
    type SuperValidated,
  } from "sveltekit-superforms";
  import { yup } from "sveltekit-superforms/adapters";
  import { inferSourceName } from "../sourceUtils";
  import { humanReadableErrorMessage } from "../errors/errors";
  import {
    submitAddOLAPConnectorForm,
    submitAddSourceForm,
  } from "./submitAddDataForm";
  import type { AddDataFormType, ConnectorType } from "./types";
  import { dsnSchema, getYupSchema } from "./yupSchemas";
  import AddClickHouseForm from "./AddClickHouseForm.svelte";
  import Checkbox from "@rilldata/web-common/components/forms/Checkbox.svelte";
  import NeedHelpText from "./NeedHelpText.svelte";
  import Tabs from "@rilldata/web-common/components/forms/Tabs.svelte";
  import { TabsContent } from "@rilldata/web-common/components/tabs";
  import { isEmpty, normalizeErrors } from "./utils";
  import { CONNECTION_TAB_OPTIONS } from "./constants";
  import { getInitialFormValuesFromProperties } from "../sourceUtils";
  import { compileConnectorYAML } from "../../connectors/code-utils";
  import CopyIcon from "@rilldata/web-common/components/icons/CopyIcon.svelte";
  import Check from "@rilldata/web-common/components/icons/Check.svelte";

  const dispatch = createEventDispatcher();

  export let connector: V1ConnectorDriver;
  export let formType: AddDataFormType;
  export let onBack: () => void;
  export let onClose: () => void;

  const isSourceForm = formType === "source";
  const isConnectorForm = formType === "connector";

  let copied = false;
  let connectionTab: ConnectorType = "parameters";

  // Form 1: Individual parameters
  const paramsFormId = `add-data-${connector.name}-form`;
  const properties =
    (isSourceForm
      ? connector.sourceProperties
      : connector.configProperties?.filter(
          (property) => property.key !== "dsn",
        )) ?? [];

  // FIXME: APP-209
  const filteredParamsProperties = properties;
  const schema = yup(getYupSchema[connector.name as keyof typeof getYupSchema]);
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
    connector.configProperties?.some((property) => property.key === "dsn");
  const dsnFormId = `add-data-${connector.name}-dsn-form`;
  const dsnProperties =
    connector.configProperties?.filter((property) => property.key === "dsn") ??
    [];

  // FIXME: APP-209
  const filteredDsnProperties = dsnProperties;
  const dsnYupSchema = yup(dsnSchema);
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
  let clickhouseManaged: boolean;
  let clickhouseParamsForm;
  let clickhouseDsnForm;

  // TODO: move to utils.ts
  // Compute disabled state for the submit button
  $: isSubmitDisabled = (() => {
    if (connectionTab === "dsn") {
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
      for (const property of properties) {
        if (property.required) {
          const key = String(property.key);
          const value = $paramsForm[key];
          if (isEmpty(value) || $paramsErrors[key]?.length) return true;
        }
      }
      return false;
    }
  })();

  $: formId = connectionTab === "dsn" ? dsnFormId : paramsFormId;
  $: submitting = connectionTab === "dsn" ? $dsnSubmitting : $paramsSubmitting;

  // Reset errors when form is modified
  $: if (connectionTab === "dsn") {
    if ($dsnTainted) dsnError = null;
  } else {
    if ($paramsTainted) paramsError = null;
  }

  // Emit the submitting state to the parent
  $: dispatch("submitting", { submitting });

  // ClickHouse requires special handling because it supports both managed and self-managed modes:
  // - When managed=true: uses sourceProperties and minimal configuration
  // - When managed=false: uses configProperties and full connection details
  // - The form state is managed by the child AddClickHouseForm component
  // - Other connectors use a simpler single-form approach
  $: yamlPreview = (() => {
    if (connector.name === "clickhouse") {
      // Use the value of the child form state for clickhouse
      const values =
        connectionTab === "dsn" ? $clickhouseDsnForm : $clickhouseParamsForm;
      return compileConnectorYAML(
        connector,
        {
          ...values,
          managed: clickhouseManaged,
        },
        {
          fieldFilter: (property) => !property.noPrompt,
          orderedProperties:
            connectionTab === "dsn"
              ? filteredDsnProperties
              : filteredParamsProperties,
        },
      );
    } else {
      const values = connectionTab === "dsn" ? $dsnForm : $paramsForm;
      return compileConnectorYAML(connector, values, {
        fieldFilter: (property) => !property.noPrompt,
        orderedProperties:
          connectionTab === "dsn"
            ? filteredDsnProperties
            : filteredParamsProperties,
      });
    }
  })();

  function copyYamlPreview() {
    navigator.clipboard.writeText(yamlPreview);
    copied = true;
    setTimeout(() => {
      copied = false;
    }, 2_000);
  }

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
      if (formType === "source") {
        await submitAddSourceForm(queryClient, connector, values);
      } else {
        await submitAddOLAPConnectorForm(queryClient, connector, values);
      }
      onClose();
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
      } else {
        error = "Unknown error";
        details = undefined;
      }

      // Keep error state for each form
      if (connectionTab === "dsn") {
        dsnError = error;
        dsnErrorDetails = details;
      } else {
        paramsError = error;
        paramsErrorDetails = details;
      }
    }
  }
</script>

<div class="add-data-layout flex flex-col h-full w-full md:flex-row">
  <!-- LEFT SIDE PANEL -->
  <div
    class="add-data-form-panel flex-1 flex flex-col min-w-0 md:pr-0 pr-0 relative"
  >
    <div
      class="flex flex-col flex-grow max-h-[552px] min-h-[552px] overflow-y-auto p-6"
    >
      {#if connector.name === "clickhouse"}
        <AddClickHouseForm
          {connector}
          {formType}
          {onClose}
          setError={(error, details) => {
            clickhouseError = error;
            clickhouseErrorDetails = details;
          }}
          bind:formId={clickhouseFormId}
          bind:submitting={clickhouseSubmitting}
          bind:isSubmitDisabled={clickhouseIsSubmitDisabled}
          bind:managed={clickhouseManaged}
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
              {#each filteredParamsProperties as property (property.key)}
                {@const propertyKey = property.key ?? ""}
                {@const label =
                  property.displayName +
                  (property.required ? "" : " (optional)")}
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
                      bind:value={$paramsForm[propertyKey]}
                      onInput={(_, e) => onStringInputChange(e)}
                      alwaysShowError
                    />
                  {:else if property.type === ConnectorDriverPropertyType.TYPE_BOOLEAN}
                    <Checkbox
                      id={propertyKey}
                      bind:checked={$paramsForm[propertyKey]}
                      {label}
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
            </form>
          </TabsContent>
          <TabsContent value="dsn">
            <form
              id={dsnFormId}
              class="pb-5 flex-grow overflow-y-auto"
              use:dsnEnhance
              on:submit|preventDefault={dsnSubmit}
            >
              {#each filteredDsnProperties as property (property.key)}
                {@const propertyKey = property.key ?? ""}
                <div class="py-1.5 first:pt-0 last:pb-0">
                  <Input
                    id={propertyKey}
                    label={property.displayName}
                    placeholder={property.placeholder}
                    secret={property.secret}
                    hint={property.hint}
                    errors={$dsnErrors[propertyKey]}
                    bind:value={$dsnForm[propertyKey]}
                    alwaysShowError
                  />
                </div>
              {/each}
            </form>
          </TabsContent>
        </Tabs>
      {:else}
        <form
          id={paramsFormId}
          class="pb-5 flex-grow overflow-y-auto"
          use:paramsEnhance
          on:submit|preventDefault={paramsSubmit}
        >
          {#each filteredParamsProperties as property (property.key)}
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
                  bind:value={$paramsForm[propertyKey]}
                  onInput={(_, e) => onStringInputChange(e)}
                  alwaysShowError
                />
              {:else if property.type === ConnectorDriverPropertyType.TYPE_BOOLEAN}
                <Checkbox
                  id={propertyKey}
                  bind:checked={$paramsForm[propertyKey]}
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
        </form>
      {/if}
    </div>

    <!-- LEFT FOOTER -->
    <div
      class="w-full bg-white border-t border-gray-200 p-6 flex justify-between gap-2"
    >
      <Button onClick={onBack} type="secondary">Back</Button>

      <Button
        disabled={connector.name === "clickhouse"
          ? clickhouseSubmitting || clickhouseIsSubmitDisabled
          : submitting || isSubmitDisabled}
        form={connector.name === "clickhouse" ? clickhouseFormId : formId}
        submitForm
        type="primary"
      >
        {#if connector.name === "clickhouse"}
          {#if clickhouseManaged}
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
          {#if submitting}
            Testing connection...
          {:else}
            Connect
          {/if}
        {:else}
          Add data
        {/if}
      </Button>
    </div>
  </div>

  <!-- RIGHT SIDE PANEL -->
  <div
    class="add-data-side-panel flex flex-col gap-6 p-6 bg-[#FAFAFA] w-full max-w-full border-l-0 border-t mt-6 pl-0 pt-6 md:w-96 md:min-w-[320px] md:max-w-[400px] md:border-l md:border-t-0 md:mt-0 md:pl-6"
  >
    {#if dsnError || paramsError || clickhouseError}
      <SubmissionError
        message={clickhouseError ??
          (connectionTab === "dsn" ? dsnError : paramsError) ??
          ""}
        details={clickhouseErrorDetails ??
          (connectionTab === "dsn" ? dsnErrorDetails : paramsErrorDetails) ??
          ""}
      />
    {/if}

    <div>
      <div class="text-sm leading-none font-medium mb-4">Connector preview</div>
      <div class="relative">
        <button
          class="absolute top-2 right-2 p-1 rounded"
          type="button"
          aria-label="Copy YAML"
          on:click={copyYamlPreview}
        >
          {#if copied}
            <Check size="16px" />
          {:else}
            <CopyIcon size="16px" />
          {/if}
        </button>
        <pre
          class="bg-muted p-3 rounded text-xs border border-gray-200 overflow-x-auto">{yamlPreview}</pre>
      </div>
    </div>

    <NeedHelpText {connector} />
  </div>
</div>
