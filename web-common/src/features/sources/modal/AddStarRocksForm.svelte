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
  import { humanReadableErrorMessage } from "../errors/errors";
  import { submitAddConnectorForm } from "./submitAddDataForm";
  import type { ConnectorType } from "./types";
  import { dsnSchema, getYupSchema } from "./yupSchemas";
  import Checkbox from "@rilldata/web-common/components/forms/Checkbox.svelte";
  import { isEmpty, normalizeErrors } from "./utils";
  import { CONNECTION_TAB_OPTIONS } from "./constants";
  import { getInitialFormValuesFromProperties } from "../sourceUtils";

  export let connector: V1ConnectorDriver;
  export let formId: string;
  export let isSubmitting: boolean;
  export let isSubmitDisabled: boolean;
  export let connectionTab: ConnectorType = "parameters";
  export let showSaveAnyway: boolean = false;
  export let onClose: () => void;
  export let setError: (
    error: string | null,
    details?: string,
  ) => void = () => {};

  export { paramsForm, dsnForm };
  export { handleSaveAnyway };

  const starrocksSchema = yup(getYupSchema["starrocks"]);
  const initialFormValues = getInitialFormValuesFromProperties(
    connector.configProperties ?? [],
  );
  const paramsFormId = `add-starrocks-data-${connector.name}-form`;
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
  let paramsError: string | null = null;
  let paramsErrorDetails: string | undefined = undefined;

  const dsnFormId = `add-starrocks-data-${connector.name}-dsn-form`;
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

  $: submitting =
    connectionTab === "parameters" ? $paramsSubmitting : $dsnSubmitting;
  $: isSubmitting = submitting;
  $: formId = connectionTab === "parameters" ? paramsFormId : dsnFormId;

  // Reset errors when form is modified
  $: if (connectionTab === "parameters") {
    if ($paramsTainted) paramsError = null;
  } else if (connectionTab === "dsn") {
    if ($dsnTainted) dsnError = null;
  }

  // Clear errors when switching tabs
  $: if (connectionTab === "dsn") {
    paramsError = null;
    paramsErrorDetails = undefined;
  } else {
    dsnError = null;
    dsnErrorDetails = undefined;
  }

  function handleOnUpdate({ form }: { form: SuperValidated<any> }) {
    // Show Save Anyway button as soon as form submission starts
    showSaveAnyway = true;

    if (form.valid) {
      handleSubmit(form.data);
    }
  }

  async function handleSubmit(values: any) {
    // Clear previous errors
    paramsError = null;
    paramsErrorDetails = undefined;
    dsnError = null;
    dsnErrorDetails = undefined;
    setError(null, undefined);

    try {
      await submitAddConnectorForm(
        queryClient,
        connector,
        values,
        false, // Normal submission, not saveAnyway
      );
      onClose();
    } catch (e) {
      let error: string;
      let details: string | undefined = undefined;
      let originalMessage: string | undefined;

      // Extract the original message from various error formats
      if (e instanceof Error) {
        originalMessage = e.message;
      } else if (e?.response?.data?.message) {
        originalMessage = e.response.data.message;
      } else if (e?.message) {
        originalMessage = e.message;
      }

      // Apply human-readable error transformation
      if (originalMessage) {
        const humanReadable = humanReadableErrorMessage(
          connector.name,
          e?.response?.data?.code,
          originalMessage,
        );
        error = humanReadable;
        // Show original message as details only if it differs from the human-readable version
        details =
          humanReadable !== originalMessage ? originalMessage : undefined;
        // Also include e.details if available and different
        if (!details && e?.details && e.details !== originalMessage) {
          details = e.details;
        }
      } else {
        error = "Unknown error";
        details = undefined;
      }
      if (connectionTab === "parameters") {
        paramsError = error;
        paramsErrorDetails = details;
        setError(paramsError, paramsErrorDetails);
      } else if (connectionTab === "dsn") {
        dsnError = error;
        dsnErrorDetails = details;
        setError(dsnError, dsnErrorDetails);
      }
    }
  }

  $: properties = (connector.configProperties ?? []).filter((p) =>
    connectionTab !== "dsn" ? p.key !== "dsn" : true,
  );

  $: filteredProperties = properties.filter((property) => !property.noPrompt);

  // Compute disabled state for the submit button
  $: isSubmitDisabled = (() => {
    if (connectionTab === "dsn") {
      // DSN form
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
      for (const property of filteredProperties) {
        if (property.required) {
          const key = String(property.key);
          const value = $paramsForm[key];
          if (isEmpty(value) || $paramsErrors[key]?.length) return true;
        }
      }
      return false;
    }
  })();

  async function handleSaveAnyway() {
    showSaveAnyway = false;
    const values = connectionTab === "dsn" ? $dsnForm : $paramsForm;

    try {
      await submitAddConnectorForm(queryClient, connector, values, true);
      onClose();
    } catch (e) {
      showSaveAnyway = true;
      let error: string;
      let details: string | undefined = undefined;
      let originalMessage: string | undefined;

      // Extract the original message from various error formats
      if (e instanceof Error) {
        originalMessage = e.message;
      } else if (e?.response?.data?.message) {
        originalMessage = e.response.data.message;
      } else if (e?.message) {
        originalMessage = e.message;
      }

      // Apply human-readable error transformation
      if (originalMessage) {
        const humanReadable = humanReadableErrorMessage(
          connector.name,
          e?.response?.data?.code,
          originalMessage,
        );
        error = humanReadable;
        details =
          humanReadable !== originalMessage ? originalMessage : undefined;
        if (!details && e?.details && e.details !== originalMessage) {
          details = e.details;
        }
      } else {
        error = "Unknown error";
        details = undefined;
      }
      if (connectionTab === "parameters") {
        paramsError = error;
        paramsErrorDetails = details;
        setError(paramsError, paramsErrorDetails);
      } else if (connectionTab === "dsn") {
        dsnError = error;
        dsnErrorDetails = details;
        setError(dsnError, dsnErrorDetails);
      }
    }
  }

  $: hasDsnProperty = properties.some((p) => p.key === "dsn");
</script>

{#if hasDsnProperty}
  <Tabs
    options={CONNECTION_TAB_OPTIONS}
    bind:activeTab={connectionTab}
    class="mb-3"
  />
{/if}

{#if connectionTab === "parameters"}
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
{:else}
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
{/if}
