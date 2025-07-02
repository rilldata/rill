<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import InformationalField from "@rilldata/web-common/components/forms/InformationalField.svelte";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
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
  import type { AddDataFormType } from "./types";
  import { dsnSchema, getYupSchema } from "./yupSchemas";
  import { ExternalLinkIcon } from "lucide-svelte";
  import {
    CONNECTOR_TYPE_OPTIONS,
    type ClickHouseConnectorType,
  } from "../../connectors/olap/constants";
  import Checkbox from "@rilldata/web-common/components/forms/Checkbox.svelte";
  import Tabs from "@rilldata/web-common/components/forms/Tabs.svelte";
  import { TabsContent } from "@rilldata/web-common/components/tabs";

  const dispatch = createEventDispatcher();

  export let connector: V1ConnectorDriver;
  export let formType: AddDataFormType;
  export let onBack: () => void;
  export let onClose: () => void;

  const isSourceForm = formType === "source";
  const isConnectorForm = formType === "connector";
  const isClickHouse = connector.name === "clickhouse";

  let connectorType: ClickHouseConnectorType = "self-managed";

  function getSpecDefaults(properties) {
    const defaults = {};
    (properties ?? []).forEach((property) => {
      if (property.default !== undefined) {
        let value = property.default;
        // Convert to correct type
        if (property.type === ConnectorDriverPropertyType.TYPE_BOOLEAN) {
          value = value === "true";
        }
        defaults[property.key] = value;
      }
    });
    return defaults;
  }

  // Form 1: Individual parameters
  const paramsFormId = `add-data-${connector.name}-form`;
  const schema = yup(getYupSchema[connector.name as keyof typeof getYupSchema]);
  const {
    form: paramsForm,
    errors: paramsErrors,
    enhance: paramsEnhance,
    tainted: paramsTainted,
    submit: paramsSubmit,
    submitting: paramsSubmitting,
  } = superForm(defaults(schema), {
    SPA: true,
    validators: schema,
    onUpdate: handleOnUpdate,
    resetForm: false,
  });
  let paramsError: string | null = null;
  let paramsErrorDetails: string | undefined = undefined;

  // Form 2: DSN
  // SuperForms are not meant to have dynamic schemas, so we use a different form instance for the DSN form
  let useDsn = false;
  let connectionTab: "parameters" | "dsn" = useDsn ? "dsn" : "parameters";
  $: useDsn = connectionTab === "dsn";
  const hasDsnFormOption =
    isConnectorForm &&
    connector.configProperties?.some((property) => property.key === "dsn");
  const dsnFormId = `add-data-${connector.name}-dsn-form`;
  const dsnProperties =
    connector.configProperties?.filter((property) => property.key === "dsn") ??
    [];
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

  // Active form
  $: formId = useDsn ? dsnFormId : paramsFormId;
  $: submitting = useDsn ? $dsnSubmitting : $paramsSubmitting;

  // Reset errors when form is modified
  $: if (useDsn) {
    if ($dsnTainted) dsnError = null;
  } else {
    if ($paramsTainted) paramsError = null;
  }

  const filteredProperties =
    (isSourceForm
      ? connector.sourceProperties
      : connector.configProperties?.filter(
          (property) =>
            !(isClickHouse && !useDsn && property.key === "dsn") &&
            // NOTE: To hide the "managed" property from the form.
            !property.noPrompt,
        )) ?? [];

  // Emit the submitting state to the parent
  $: dispatch("submitting", { submitting });

  // Update params form whenever the data connector changes, unless the form is tainted
  $: if (connector) {
    const specDefaults = getSpecDefaults(connector.configProperties);
    paramsForm.update(($form) => ({
      ...$form,
      ...specDefaults,
    }));
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
    let values = event.form.data;

    // Ensure managed property is set for ClickHouse
    if (isClickHouse) {
      if (connectorType === "rill-managed") {
        Object.keys(values).forEach((key) => {
          if (key !== "managed")
            delete (values as Record<string, unknown>)[key];
        });
        (values as Record<string, unknown>).managed = true;
      } else {
        (values as Record<string, unknown>).managed = false;
      }
    }

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
      if (useDsn) {
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
  <div class="add-data-form-panel flex-1 flex flex-col min-w-0 md:pr-0 pr-0">
    <div
      class="p-6 flex flex-col flex-grow max-h-[552px] min-h-[552px] overflow-y-auto"
    >
      {#if !isClickHouse}
        <!-- Non-ClickHouse Form -->
        <form
          id={paramsFormId}
          use:paramsEnhance
          on:submit|preventDefault={paramsSubmit}
        >
          {#each filteredProperties as property (property.key)}
            {@const propertyKey = property.key ?? ""}
            {@const label =
              property.displayName + (property.required ? "" : " (optional)")}
            <div class="py-1.5 first:pt-0 last:pb-0">
              {#if property.type === ConnectorDriverPropertyType.TYPE_STRING || property.type === ConnectorDriverPropertyType.TYPE_NUMBER}
                <Input
                  id={propertyKey}
                  label={property.displayName}
                  placeholder={property.placeholder}
                  optional={!property.required}
                  secret={property.secret}
                  hint={property.hint}
                  errors={$paramsErrors[propertyKey]}
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
      {:else}
        <!-- ClickHouse Form -->
        <div class="pb-3">
          <Select
            id="connector-type"
            options={CONNECTOR_TYPE_OPTIONS}
            bind:value={connectorType}
            label="Connector type"
          />
          {#if connectorType === "self-managed"}
            <div class="text-xs text-muted-foreground mt-2">
              Connect to your own ClickHouse instance (Cloud or self-hosted)
            </div>
          {/if}
        </div>

        {#if connectorType === "rill-managed"}
          <InformationalField
            description="This option uses ClickHouse as an OLAP engine with Rill-managed infrastructure. No additional configuration is required - Rill will handle the setup and management of your ClickHouse instance."
          />
          <form
            id={paramsFormId}
            use:paramsEnhance
            on:submit|preventDefault={paramsSubmit}
          >
            <input type="hidden" name="managed" value="true" />
          </form>
        {:else}
          <!-- Self-managed ClickHouse Form -->
          {#if hasDsnFormOption}
            <div class="pb-3">
              <Tabs
                value={connectionTab}
                options={[
                  { value: "parameters", label: "Enter parameters" },
                  { value: "dsn", label: "Enter connection string" },
                ]}
                on:change={(event) => (connectionTab = event.detail)}
              >
                <TabsContent value="parameters">
                  <!-- Parameters Form -->
                  <form
                    id={paramsFormId}
                    use:paramsEnhance
                    on:submit|preventDefault={paramsSubmit}
                  >
                    {#each filteredProperties as property (property.key)}
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
                            errors={$paramsErrors[propertyKey]}
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
                  <!-- DSN Form -->
                  <form
                    id={dsnFormId}
                    use:dsnEnhance
                    on:submit|preventDefault={dsnSubmit}
                  >
                    {#each dsnProperties as property (property.key)}
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
            </div>
          {/if}
        {/if}
      {/if}
    </div>
    <div
      class="flex items-center justify-between space-x-2 px-6 py-4 border-t border-gray-200"
    >
      <Button onClick={onBack} type="secondary">Back</Button>
      <Button disabled={submitting} form={formId} submitForm type="primary">
        {#if isConnectorForm}
          {#if isClickHouse && connectorType === "rill-managed"}
            {submitting ? "Connecting..." : "Connect"}
          {:else}
            {submitting ? "Testing connection..." : "Test and Connect"}
          {/if}
        {:else}
          Add data
        {/if}
      </Button>
    </div>
  </div>

  <div
    class="add-data-side-panel flex flex-col gap-6 p-6 bg-[#FAFAFA] w-full max-w-full border-l-0 border-t mt-6 pl-0 pt-6 md:w-96 md:min-w-[320px] md:max-w-[400px] md:border-l md:border-t-0 md:mt-0 md:pl-6"
  >
    {#if dsnError || paramsError}
      <SubmissionError
        message={(useDsn ? dsnError : paramsError) ?? ""}
        details={(useDsn ? dsnErrorDetails : paramsErrorDetails) ?? ""}
      />
    {/if}
    <div>
      <div class="text-sm leading-none font-medium mb-4">Help</div>
      <div
        class="text-sm leading-normal font-medium text-muted-foreground mb-2"
      >
        Need help connecting to {connector.displayName}? Check out our
        documentation for detailed instructions.
      </div>
      <span class="flex flex-row items-center gap-2 group">
        <a
          href={connector.docsUrl || "https://docs.rilldata.com/build/connect/"}
          rel="noreferrer noopener"
          target="_blank"
          class="text-sm leading-normal text-primary-500 hover:text-primary-600 font-medium group-hover:underline break-all"
        >
          How to connect to {connector.displayName}
        </a>
        <ExternalLinkIcon size="16px" color="#6366F1" />
      </span>
    </div>
  </div>
</div>
