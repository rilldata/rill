<script lang="ts">
  import InformationalField from "@rilldata/web-common/components/forms/InformationalField.svelte";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import {
    ConnectorDriverPropertyType,
    type V1ConnectorDriver,
  } from "@rilldata/web-common/runtime-client";
  import { createEventDispatcher } from "svelte";
  import {
    defaults,
    superForm,
    type SuperValidated,
  } from "sveltekit-superforms";
  import { yup } from "sveltekit-superforms/adapters";
  import Tabs from "@rilldata/web-common/components/forms/Tabs.svelte";
  import { TabsContent } from "@rilldata/web-common/components/tabs";
  import { inferSourceName } from "../sourceUtils";
  import { humanReadableErrorMessage } from "../errors/errors";
  import {
    submitAddOLAPConnectorForm,
    submitAddSourceForm,
  } from "./submitAddDataForm";
  import type { AddDataFormType, ConnectorType } from "./types";
  import { dsnSchema, getYupSchema } from "./yupSchemas";
  import Checkbox from "@rilldata/web-common/components/forms/Checkbox.svelte";
  import { isEmpty, normalizeErrors } from "./utils";
  import { CONNECTOR_TYPE_OPTIONS, CONNECTION_TAB_OPTIONS } from "./constants";
  import ConnectorTypeSelector from "@rilldata/web-common/components/forms/ConnectorTypeSelector.svelte";
  import { getInitialFormValuesFromProperties } from "../sourceUtils";

  export let connector: V1ConnectorDriver;
  export let formType: AddDataFormType;
  export let formId: string;
  export let submitting: boolean;
  export let isSubmitDisabled: boolean;
  export let managed: boolean;
  export let onClose: () => void;
  export let setError: (
    error: string | null,
    details?: string,
  ) => void = () => {};
  export let connectionTab: ConnectorType = "parameters";
  export { paramsForm, dsnForm };

  const dispatch = createEventDispatcher();

  // Always include 'managed' in the schema for ClickHouse
  const clickhouseSchema = yup(getYupSchema["clickhouse"]);
  const initialFormValues = getInitialFormValuesFromProperties(
    connector.configProperties ?? [],
  );
  const paramsFormId = `add-clickhouse-data-${connector.name}-form`;
  const {
    form: paramsForm,
    errors: paramsErrors,
    enhance: paramsEnhance,
    tainted: paramsTainted,
    submit: paramsSubmit,
    submitting: paramsSubmitting,
  } = superForm(initialFormValues, {
    SPA: true,
    validators: clickhouseSchema,
    onUpdate: handleOnUpdate,
    resetForm: false,
  });
  let paramsError: string | null = null;
  let paramsErrorDetails: string | undefined = undefined;

  const dsnFormId = `add-clickhouse-data-${connector.name}-dsn-form`;
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

  $: managed = $paramsForm.managed;
  $: submitting = connectionTab === "dsn" ? $dsnSubmitting : $paramsSubmitting;
  $: formId = connectionTab === "dsn" ? dsnFormId : paramsFormId;

  // Reset connectionTab if switching to Rill-managed
  $: if ($paramsForm.managed) {
    connectionTab = "parameters";
  }

  // Reset errors when form is modified
  $: if (connectionTab === "dsn") {
    if ($dsnTainted) dsnError = null;
  } else {
    if ($paramsTainted) paramsError = null;
  }

  // Emit the submitting state to the parent
  $: dispatch("submitting", { submitting });

  let prevManaged = $paramsForm.managed;
  $: {
    // Switching to managed: strip all but managed
    if ($paramsForm.managed && Object.keys($paramsForm).length > 1) {
      paramsForm.update(() => ({ managed: true }), { taint: false });
      resetError();
    }
    // Switching to self-managed: restore defaults
    else if (prevManaged && !$paramsForm.managed) {
      paramsForm.update(() => ({ ...initialFormValues, managed: false }), {
        taint: false,
      });
    }
    prevManaged = $paramsForm.managed;
  }

  function onStringInputChange(event: Event) {
    const target = event.target as HTMLInputElement;
    const { name, value } = target;
    if (name === "path") {
      if ($paramsTainted?.name) return;
      const nameVal = inferSourceName(connector, value);
      if (nameVal)
        paramsForm.update(
          ($form) => {
            $form.name = nameVal;
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
    result: Extract<
      import("@sveltejs/kit").ActionResult,
      { type: "success" | "failure" }
    >;
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
      if (connectionTab === "dsn") {
        dsnError = error;
        dsnErrorDetails = details;
        setError(dsnError, dsnErrorDetails);
      } else {
        paramsError = error;
        paramsErrorDetails = details;
        setError(paramsError, paramsErrorDetails);
      }
    }
  }

  $: properties = $paramsForm.managed
    ? (connector.sourceProperties ?? [])
    : (connector.configProperties?.filter((p) =>
        connectionTab !== "dsn" ? p.key !== "dsn" : true,
      ) ?? []);
  $: filteredProperties = properties.filter(
    (property) => !property.noPrompt && property.key !== "managed",
  );

  // TODO: move to utils.ts
  // Compute disabled state for the submit button
  // Refer to `runtime/drivers/clickhouse/clickhouse.go` for the required
  // Account for the managed property and the dsn property can be either true or false
  $: isSubmitDisabled = (() => {
    if ($paramsForm.managed) {
      // Managed form: only check required properties where property.key === 'managed' or property.key is not 'managed'
      for (const property of filteredProperties) {
        if (
          property.required &&
          (property.key === "managed" || property.key !== "managed")
        ) {
          const key = String(property.key);
          const value = $paramsForm[key];
          if (isEmpty(value) || $paramsErrors[key]?.length) return true;
        }
      }
      return false;
    } else if (connectionTab === "dsn") {
      // Self-managed DSN form
      for (const property of dsnProperties) {
        if (property.required) {
          const key = String(property.key);
          const value = $dsnForm[key];
          if (isEmpty(value) || $dsnErrors[key]?.length) return true;
        }
      }
      return false;
    } else {
      // Self-managed parameters form: only check required properties where property.key !== 'managed'
      for (const property of filteredProperties) {
        if (property.required && property.key !== "managed") {
          const key = String(property.key);
          const value = $paramsForm[key];
          if (isEmpty(value) || $paramsErrors[key]?.length) return true;
        }
      }
      return false;
    }
  })();

  function resetError() {
    paramsError = null;
    paramsErrorDetails = undefined;
    dsnError = null;
    dsnErrorDetails = undefined;
    setError(null, undefined);
  }
</script>

<div class="h-full w-full flex flex-col">
  <div>
    <ConnectorTypeSelector
      bind:value={$paramsForm.managed}
      options={CONNECTOR_TYPE_OPTIONS}
    />
    {#if $paramsForm.managed}
      <InformationalField
        description="This option uses ClickHouse as an OLAP engine with Rill-managed infrastructure. No additional configuration is required - Rill will handle the setup and management of your ClickHouse instance."
      />
    {/if}
  </div>

  {#if !$paramsForm.managed}
    <Tabs bind:value={connectionTab} options={CONNECTION_TAB_OPTIONS}>
      <TabsContent value="parameters">
        <form
          id={paramsFormId}
          class="pb-5 flex-grow overflow-y-auto"
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
      </TabsContent>
      <TabsContent value="dsn">
        <form
          id={dsnFormId}
          class="pb-5 flex-grow overflow-y-auto"
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
                errors={normalizeErrors($dsnErrors[propertyKey])}
                bind:value={$dsnForm[propertyKey]}
                alwaysShowError
              />
            </div>
          {/each}
        </form>
      </TabsContent>
    </Tabs>
  {:else}
    <!-- Only managed form -->
    <form
      id={paramsFormId}
      class="pb-5 flex-grow overflow-y-auto"
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
