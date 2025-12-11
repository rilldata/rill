<script lang="ts">
  import InformationalField from "@rilldata/web-common/components/forms/InformationalField.svelte";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import {
    ConnectorDriverPropertyType,
    type V1ConnectorDriver,
  } from "@rilldata/web-common/runtime-client";
  import {
    defaults,
    superForm,
    type SuperValidated,
  } from "sveltekit-superforms";
  import { yup } from "sveltekit-superforms/adapters";
  import Tabs from "@rilldata/web-common/components/forms/Tabs.svelte";
  import { TabsContent } from "@rilldata/web-common/components/tabs";
  import { humanReadableErrorMessage } from "../errors/errors";
  import { submitAddConnectorForm } from "./submitAddDataForm";
  import type { ConnectorType } from "./types";
  import { dsnSchema, getYupSchema } from "./yupSchemas";
  import Checkbox from "@rilldata/web-common/components/forms/Checkbox.svelte";
  import { isEmpty, normalizeErrors } from "./utils";
  import { CONNECTION_TAB_OPTIONS } from "./constants";
  import { getInitialFormValuesFromProperties } from "../sourceUtils";

  // Types
  interface ApiError {
    response?: {
      data?: {
        message?: string;
        code?: string;
      };
    };
    message?: string;
    details?: string;
  }

  interface FormError {
    message: string | null;
    details?: string;
  }

  type FormErrors = Record<ConnectorType, FormError>;

  // Props
  export let connector: V1ConnectorDriver;
  export let onClose: () => void;
  export let setError: (
    error: string | null,
    details?: string,
  ) => void = () => {};

  // Exported state
  export let formId: string = "";
  export let isSubmitting: boolean = false;
  export let isSubmitDisabled: boolean = false;
  export let connectionTab: ConnectorType = "parameters";
  export let showSaveAnyway: boolean = false;

  export { paramsForm, dsnForm, handleSaveAnyway };

  // Constants
  const CONNECTOR_NAME = "starrocks";
  const paramsFormId = `add-${CONNECTOR_NAME}-data-${connector.name}-form`;
  const dsnFormId = `add-${CONNECTOR_NAME}-data-${connector.name}-dsn-form`;

  // Derived properties
  $: dsnProperties =
    connector.configProperties?.filter((p) => p.key === "dsn") ?? [];
  $: hasDsnProperty = dsnProperties.length > 0;
  $: nonDsnProperties = (connector.configProperties ?? []).filter(
    (p) => p.key !== "dsn",
  );
  $: filteredProperties = nonDsnProperties.filter((p) => !p.noPrompt);

  // Parameters form setup
  const starrocksSchema = yup(getYupSchema[CONNECTOR_NAME]);
  const initialFormValues = getInitialFormValuesFromProperties(
    connector.configProperties ?? [],
  );

  const {
    form: paramsForm,
    errors: paramsErrors,
    enhance: paramsEnhance,
    tainted: paramsTainted,
    submit: paramsSubmit,
    submitting: paramsSubmitting,
  } = superForm(initialFormValues, {
    SPA: true,
    validators: starrocksSchema,
    onUpdate: handleOnUpdate,
    resetForm: false,
  });

  // DSN form setup
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

  // Error state (unified)
  let formErrors: FormErrors = {
    parameters: { message: null },
    dsn: { message: null },
  };

  // Reactive bindings
  $: isSubmitting =
    connectionTab === "parameters" ? $paramsSubmitting : $dsnSubmitting;
  $: formId = connectionTab === "parameters" ? paramsFormId : dsnFormId;

  // Clear errors when form is modified
  $: if (connectionTab === "parameters" && $paramsTainted) {
    formErrors.parameters = { message: null };
  }
  $: if (connectionTab === "dsn" && $dsnTainted) {
    formErrors.dsn = { message: null };
  }

  // Clear opposite tab's errors when switching
  $: {
    const oppositeTab = connectionTab === "parameters" ? "dsn" : "parameters";
    formErrors[oppositeTab] = { message: null };
  }

  // Compute disabled state
  $: isSubmitDisabled = (() => {
    const properties =
      connectionTab === "dsn" ? dsnProperties : filteredProperties;
    const form = connectionTab === "dsn" ? $dsnForm : $paramsForm;
    const errors = connectionTab === "dsn" ? $dsnErrors : $paramsErrors;

    for (const property of properties) {
      if (property.required) {
        const key = String(property.key);
        if (isEmpty(form[key]) || errors[key]?._errors?.length) {
          return true;
        }
      }
    }
    return false;
  })();

  // Error extraction utility
  function extractErrorInfo(e: unknown): { error: string; details?: string } {
    const apiError = e as ApiError;
    const originalMessage =
      e instanceof Error
        ? e.message
        : (apiError?.response?.data?.message ?? apiError?.message);

    if (!originalMessage) {
      return { error: "Unknown error" };
    }

    const errorCode = apiError?.response?.data?.code;
    const numericCode = typeof errorCode === "string" ? undefined : errorCode;
    const humanReadableResult = humanReadableErrorMessage(
      connector.name,
      numericCode,
      originalMessage,
    );
    const humanReadable =
      typeof humanReadableResult === "string"
        ? humanReadableResult
        : originalMessage;

    const details =
      humanReadable !== originalMessage
        ? originalMessage
        : apiError?.details !== originalMessage
          ? apiError?.details
          : undefined;

    return { error: humanReadable, details };
  }

  function setCurrentError(error: string, details?: string) {
    formErrors[connectionTab] = { message: error, details };
    setError(error, details);
  }

  function clearAllErrors() {
    formErrors = { parameters: { message: null }, dsn: { message: null } };
    setError(null, undefined);
  }

  function handleOnUpdate({
    form,
  }: {
    form: SuperValidated<Record<string, unknown>>;
  }) {
    showSaveAnyway = true;
    if (form.valid) {
      handleSubmit(form.data);
    }
  }

  async function handleSubmit(values: Record<string, unknown>) {
    clearAllErrors();

    try {
      await submitAddConnectorForm(queryClient, connector, values, false);
      onClose();
    } catch (e) {
      const { error, details } = extractErrorInfo(e);
      setCurrentError(error, details);
    }
  }

  async function handleSaveAnyway() {
    showSaveAnyway = false;
    const values = connectionTab === "dsn" ? $dsnForm : $paramsForm;

    try {
      await submitAddConnectorForm(queryClient, connector, values, true);
      onClose();
    } catch (e) {
      showSaveAnyway = true;
      const { error, details } = extractErrorInfo(e);
      setCurrentError(error, details);
    }
  }
</script>

<div class="h-full w-full flex flex-col">
  {#if hasDsnProperty}
    <Tabs bind:value={connectionTab} options={CONNECTION_TAB_OPTIONS}>
      <TabsContent value="parameters">
        <form
          id={paramsFormId}
          class="flex-grow overflow-y-auto"
          method="POST"
          use:paramsEnhance
          on:submit|preventDefault={paramsSubmit}
        >
          {#each filteredProperties as property (property.key)}
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
      </TabsContent>

      <TabsContent value="dsn">
        <form
          id={dsnFormId}
          class="flex-grow overflow-y-auto"
          method="POST"
          use:dsnEnhance
          on:submit|preventDefault={dsnSubmit}
        >
          {#each dsnProperties as property (property.key)}
            {@const propertyKey = property.key ?? ""}
            <div class="py-1.0 first:pt-0 last:pb-0">
              <Input
                id={propertyKey}
                label={property.displayName}
                placeholder={property.placeholder}
                secret={property.secret}
                hint={property.hint}
                errors={normalizeErrors($dsnErrors[propertyKey])}
                bind:value={$dsnForm[propertyKey]}
                alwaysShowError
              />
            </div>
          {/each}
          <InformationalField>
            Refer to the
            <a
              href="https://docs.starrocks.io/"
              target="_blank"
              rel="noopener noreferrer"
              class="underline"
            >
              StarRocks documentation
            </a>
            for DSN format.
          </InformationalField>
        </form>
      </TabsContent>
    </Tabs>
  {:else}
    <form
      id={paramsFormId}
      class="flex-grow overflow-y-auto"
      method="POST"
      use:paramsEnhance
      on:submit|preventDefault={paramsSubmit}
    >
      {#each filteredProperties as property (property.key)}
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
