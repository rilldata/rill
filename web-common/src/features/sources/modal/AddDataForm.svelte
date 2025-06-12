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
  import { humanReadableErrorMessage } from "./errors";
  import {
    submitAddOLAPConnectorForm,
    submitAddSourceForm,
  } from "./submitAddDataForm";
  import type { AddDataFormType } from "./types";
  import { dsnSchema, getYupSchema } from "./yupSchemas";
  import yaml from "js-yaml";
  import { ExternalLinkIcon } from "lucide-svelte";

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

  // Generate YAML preview from form state
  $: yamlPreview = (() => {
    let values = useDsn ? $dsnForm : $paramsForm;
    let props = useDsn ? dsnProperties : properties;
    let out = {};
    for (const property of props) {
      const key = property.key;
      if (!key) continue;
      let value = values[key];
      if (property.secret && value) {
        value = "********";
      }
      if (value !== undefined && value !== null && value !== "") {
        out[key] = value;
      }
    }
    const title = `# ${connector.displayName} Connector\n\nConfiguration`;
    if (Object.keys(out).length === 0) return title;
    return `${title}\n${yaml.dump(out, { lineWidth: 80 })}`;
  })();

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

      // Handle different error types
      if (e instanceof Error) {
        error = e.message;
      } else if (e?.response?.data) {
        error = humanReadableErrorMessage(
          connector.name,
          e.response.data.code,
          e.response.data.message,
        );
      } else {
        error = "Unknown error";
      }

      // Keep error state for each form
      if (useDsn) {
        dsnError = error;
      } else {
        paramsError = error;
      }
    }
  }
</script>

<div class="add-data-layout">
  <div class="add-data-form-panel">
    <div class="p-6 flex flex-col flex-grow">
      {#if hasDsnFormOption}
        <div class="pb-3">
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

      {#if !useDsn}
        <!-- Form 1: Individual parameters -->
        <form
          id={paramsFormId}
          class="flex-grow overflow-y-auto"
          use:paramsEnhance
          on:submit|preventDefault={paramsSubmit}
          transition:slide={{ duration: FORM_TRANSITION_DURATION }}
        >
          {#if paramsError}
            <SubmissionError message={paramsError} />
          {/if}

          {#each properties as property (property.key)}
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
                <label for={property.key} class="flex items-center">
                  <input
                    id={propertyKey}
                    type="checkbox"
                    bind:checked={$paramsForm[propertyKey]}
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
      {:else}
        <!-- Form 2: DSN -->
        <form
          id={dsnFormId}
          class="flex-grow overflow-y-auto"
          use:dsnEnhance
          on:submit|preventDefault={dsnSubmit}
          transition:slide={{ duration: FORM_TRANSITION_DURATION }}
        >
          {#if dsnError}
            <SubmissionError message={dsnError} />
          {/if}

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
    </div>
    <div
      class="flex items-center justify-between space-x-2 px-6 py-4 border-t border-gray-200"
    >
      <Button onClick={onBack} type="secondary">Back</Button>
      <Button disabled={submitting} form={formId} submitForm type="primary">
        {#if isConnectorForm}
          {#if submitting}
            Testing connection...
          {:else}
            Test and Connect
          {/if}
        {:else}
          Add data
        {/if}
      </Button>
    </div>
  </div>

  <div class="add-data-side-panel">
    <div>
      <div class="text-sm leading-none font-medium mb-4">
        Connection preview
      </div>
      <pre>{yamlPreview}</pre>
    </div>
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
          class="text-sm leading-normal text-primary-500 hover:text-primary-600 font-medium group-hover:underline"
        >
          How to connect to {connector.displayName}
        </a>
        <ExternalLinkIcon size="16px" color="#6366F1" />
      </span>
    </div>
  </div>
</div>

<style lang="postcss">
  .add-data-layout {
    @apply flex flex-row h-full w-full;
  }
  .add-data-form-panel {
    @apply flex-1 flex flex-col min-w-0;
  }
  .add-data-side-panel {
    @apply w-96 min-w-[320px] max-w-[400px] border-l border-gray-200 pl-6 flex flex-col gap-6 p-6;
    /* FIXME: bg-sidebar-background */
    @apply bg-[#FAFAFA];
  }
  .add-data-side-panel pre {
    @apply p-4 rounded-md text-xs border border-gray-200 font-medium;
    @apply whitespace-pre-wrap overflow-x-visible;
    /* FIXME: bg-base-muted */
    @apply bg-[#F4F4F5];
  }
  .add-data-side-panel a {
    @apply text-primary-500 font-medium break-all;
  }
  @media (max-width: 900px) {
    .add-data-layout {
      @apply flex-col;
    }
    .add-data-side-panel {
      @apply w-full max-w-full border-l-0 border-t mt-6 pl-0 pt-6;
    }
    .add-data-form-panel {
      @apply pr-0;
    }
  }
</style>
