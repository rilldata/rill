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
  import { slide } from "svelte/transition";
  import {
    defaults,
    superForm,
    type SuperValidated,
  } from "sveltekit-superforms";
  import { yup } from "sveltekit-superforms/adapters";
  import { ButtonGroup, SubButton } from "../../../components/button-group";
  import { inferSourceName } from "../sourceUtils";
  import { humanReadableErrorMessage } from "../errors/errors";
  import {
    submitAddOLAPConnectorForm,
    submitAddSourceForm,
  } from "./submitAddDataForm";
  import type { AddDataFormType } from "./types";
  import { dsnSchema, getYupSchema } from "./yupSchemas";
  import AddClickHouseForm from "./AddClickHouseForm.svelte";
  import Checkbox from "@rilldata/web-common/components/forms/Checkbox.svelte";
  import { getSpecDefaults } from "./utils";

  const FORM_TRANSITION_DURATION = 150;
  const dispatch = createEventDispatcher();

  export let connector: V1ConnectorDriver;
  export let formType: AddDataFormType;
  export let onBack: () => void;
  export let onClose: () => void;

  const isSourceForm = formType === "source";
  const isConnectorForm = formType === "connector";

  // Form 1: Individual parameters
  const paramsFormId = `add-data-${connector.name}-form`;
  const properties =
    (isSourceForm
      ? connector.sourceProperties
      : connector.configProperties?.filter(
          (property) => property.key !== "dsn",
        )) ?? [];
  const schema = yup(getYupSchema[connector.name as keyof typeof getYupSchema]);
  const initialDefaults = {
    ...defaults(schema),
    ...getSpecDefaults(connector.configProperties),
  };
  $: console.log("initialDefaults: ", initialDefaults);
  const {
    form: paramsForm,
    errors: paramsErrors,
    enhance: paramsEnhance,
    tainted: paramsTainted,
    submit: paramsSubmit,
    submitting: paramsSubmitting,
  } = superForm(initialDefaults, {
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

  function handleConnectionTypeChange(e: CustomEvent<any>): void {
    useDsn = e.detail === "dsn";
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
      if (useDsn) {
        dsnError = error;
        dsnErrorDetails = details;
      } else {
        paramsError = error;
        paramsErrorDetails = details;
      }
    }
  }

  $: console.log("$paramsForm: ", $paramsForm);
</script>

<div class="h-full w-full flex flex-col">
  <div class="pb-1 text-slate-500">
    Need help? Refer to our
    <a
      href={connector.docsUrl || "https://docs.rilldata.com/build/connect/"}
      rel="noreferrer noopener"
      target="_blank">docs</a
    >
    for more information.
  </div>

  <!-- ClickHouse has a special form that handles the managed flag -->
  {#if connector.name === "clickhouse"}
    <AddClickHouseForm
      {connector}
      {formType}
      {onBack}
      {onClose}
      on:submitting
    />
  {:else}
    <!-- For all other connectors, we show a connection method selector -->
    {#if hasDsnFormOption}
      <div class="py-3">
        <div class="text-sm font-medium mb-2">Connection method</div>
        <ButtonGroup
          selected={[useDsn ? "dsn" : "parameters"]}
          on:subbutton-click={handleConnectionTypeChange}
        >
          <SubButton value="parameters" ariaLabel="Enter parameters">
            <span class="px-2">Enter parameters</span>
          </SubButton>
          <SubButton value="dsn" ariaLabel="Use connection string">
            <span class="px-2">Enter connection string</span>
          </SubButton>
        </ButtonGroup>
      </div>
    {/if}

    <!-- If the user has selected to enter parameters, we show the parameters form -->
    {#if !useDsn}
      <!-- Form 1: Individual parameters -->
      {#if paramsError}
        <SubmissionError message={paramsError} details={paramsErrorDetails} />
      {/if}
      <form
        id={paramsFormId}
        class="pb-5 flex-grow overflow-y-auto"
        use:paramsEnhance
        on:submit|preventDefault={paramsSubmit}
        transition:slide={{ duration: FORM_TRANSITION_DURATION }}
      >
        {#each properties as property (property.key)}
          {@const propertyKey = property.key ?? ""}
          {@const label =
            property.displayName + (property.required ? "" : " (optional)")}
          <div class="py-1.5">
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
      <!-- If the user has selected to enter a connection string, we show the connection string form -->
    {:else}
      <!-- Form 2: DSN -->
      {#if dsnError}
        <SubmissionError message={dsnError} details={dsnErrorDetails} />
      {/if}
      <form
        id={dsnFormId}
        class="pb-5 flex-grow overflow-y-auto"
        use:dsnEnhance
        on:submit|preventDefault={dsnSubmit}
        transition:slide={{ duration: FORM_TRANSITION_DURATION }}
      >
        {#each dsnProperties as property (property.key)}
          {@const propertyKey = property.key ?? ""}
          <div class="py-1.5">
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
    {/if}

    <div class="flex items-center space-x-2 ml-auto">
      <Button onClick={onBack} type="secondary">Back</Button>
      <Button disabled={submitting} form={formId} submitForm type="primary">
        {#if isConnectorForm}
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
  {/if}
</div>
