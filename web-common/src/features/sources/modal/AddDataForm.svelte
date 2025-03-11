<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import InformationalField from "@rilldata/web-common/components/forms/InformationalField.svelte";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import SubmissionError from "@rilldata/web-common/components/forms/SubmissionError.svelte";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import {
    ConnectorDriverPropertyType,
    type RpcStatus,
    type V1ConnectorDriver,
  } from "@rilldata/web-common/runtime-client";
  import { slide } from "svelte/transition";
  import { defaults, superForm } from "sveltekit-superforms";
  import { yup } from "sveltekit-superforms/adapters";
  import { ButtonGroup, SubButton } from "../../../components/button-group";
  import { inferSourceName } from "../sourceUtils";
  import { humanReadableErrorMessage } from "./errors";
  import { submitAddDataForm } from "./submitAddDataForm";
  import type { AddDataFormType } from "./types";
  import { dsnSchema, getYupSchema } from "./yupSchemas";

  const FORM_TRANSITION_DURATION = 150;

  export let connector: V1ConnectorDriver;
  export let formType: AddDataFormType;
  export let onBack: () => void;
  export let onClose: () => void;

  let useDsn = false;

  $: formId = `add-data-${connector.name}-form`;
  $: dsnFormId = `add-data-${connector.name}-dsn-form`;

  $: isSourceForm = formType === "source";
  $: isConnectorForm = formType === "connector";

  $: hasDsnFormOption =
    isConnectorForm &&
    connector.configProperties?.some((property) => property.key === "dsn");

  // Form 1
  const schema = yup(getYupSchema[connector.name as keyof typeof getYupSchema]);
  const { form, errors, enhance, tainted, submit, submitting } = superForm(
    defaults(schema),
    {
      SPA: true,
      validators: schema,
      onUpdate: handleOnUpdate,
    },
  );
  let rpcError: RpcStatus | null = null;
  $: properties =
    (isSourceForm
      ? connector.sourceProperties
      : connector.configProperties?.filter(
          (property) => property.key !== "dsn",
        )) ?? [];

  // Form 2
  // SuperForms are not meant to have dynamic schemas, so we use a different form instance for the DSN form
  const dsnYupSchema = yup(dsnSchema);
  const {
    form: dsnForm,
    errors: dsnErrors,
    enhance: dsnEnhance,
    submit: dsnSubmit,
    submitting: dsnSubmitting,
  } = superForm(defaults(dsnYupSchema), {
    SPA: true,
    validators: dsnYupSchema,
    onUpdate: handleOnUpdate,
  });
  const dsnProperties =
    connector.configProperties?.filter((property) => property.key === "dsn") ??
    [];

  function handleConnectionTypeChange(e: CustomEvent<any>): void {
    useDsn = e.detail === "dsn";
  }

  function onStringInputChange(event: Event) {
    const target = event.target as HTMLInputElement;
    const { name, value } = target;

    if (name === "path") {
      if ($tainted?.name) return;
      const name = inferSourceName(connector, value);
      if (name)
        form.update(
          ($form) => {
            $form.name = name;
            return $form;
          },
          { taint: false },
        );
    }
  }

  async function handleOnUpdate({ form }) {
    if (!form.valid) return;
    const values = form.data;
    if (isSourceForm) {
      try {
        await submitAddDataForm(queryClient, formType, connector, values);
        onClose();
      } catch (e) {
        rpcError = e?.response?.data;
      }
      return;
    }

    // Connectors
    try {
      await submitAddDataForm(queryClient, formType, connector, values);
      onClose();
    } catch (e) {
      rpcError = e?.response?.data;
    }
  }
</script>

<div class="h-full w-full flex flex-col">
  <div class="pb-2 text-slate-500">
    Need help? Refer to our
    <a
      href="https://docs.rilldata.com/build/connect"
      rel="noreferrer noopener"
      target="_blank">docs</a
    > for more information.
  </div>

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

  <!-- Form 1 -->
  {#if !useDsn}
    <form
      id={formId}
      class="pb-5 flex-grow overflow-y-auto"
      use:enhance
      on:submit|preventDefault={submit}
      transition:slide={{ duration: FORM_TRANSITION_DURATION }}
    >
      {#if rpcError}
        <SubmissionError
          message={humanReadableErrorMessage(
            connector.name,
            rpcError.code,
            rpcError.message,
          )}
        />
      {/if}

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
              errors={$errors[propertyKey]}
              bind:value={$form[propertyKey]}
              onInput={(_, e) => onStringInputChange(e)}
              alwaysShowError
            />
          {:else if property.type === ConnectorDriverPropertyType.TYPE_BOOLEAN}
            <label for={property.key} class="flex items-center">
              <input
                id={propertyKey}
                type="checkbox"
                bind:checked={$form[propertyKey]}
                class="h-5 w-5"
              />
              <span class="ml-2 text-sm">{label}</span>
            </label>
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

  <!-- Form 2 -->
  {#if useDsn}
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
    <Button on:click={onBack} type="secondary">Back</Button>
    <Button
      disabled={useDsn ? $dsnSubmitting : $submitting}
      form={useDsn ? dsnFormId : formId}
      submitForm
      type="primary"
    >
      Add data
    </Button>
  </div>
</div>
