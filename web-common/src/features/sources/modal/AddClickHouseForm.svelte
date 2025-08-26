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
  import { submitAddConnectorForm } from "./submitAddDataForm";
  import type { ConnectorType } from "./types";
  import { dsnSchema, getYupSchema } from "./yupSchemas";
  import Checkbox from "@rilldata/web-common/components/forms/Checkbox.svelte";
  import { isEmpty, normalizeErrors } from "./utils";
  import {
    CONNECTOR_TYPE_OPTIONS,
    CONNECTION_TAB_OPTIONS,
    type ClickHouseConnectorType,
  } from "./constants";
  import ConnectorTypeSelector from "@rilldata/web-common/components/forms/ConnectorTypeSelector.svelte";
  import { getInitialFormValuesFromProperties } from "../sourceUtils";

  export let connector: V1ConnectorDriver;
  export let formId: string;
  export let submitting: boolean;
  export let isSubmitDisabled: boolean;
  export let connectorType: ClickHouseConnectorType = "self-managed";
  export let onClose: () => void;
  export let setError: (
    error: string | null,
    details?: string,
  ) => void = () => {};
  export let connectionTab: ConnectorType = "parameters";
  export { paramsForm, dsnForm };

  const dispatch = createEventDispatcher();

  // ClickHouse schema includes the 'managed' property for backend compatibility
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

  $: submitting =
    connectorType === "clickhouse-cloud" || connectionTab === "parameters"
      ? $paramsSubmitting
      : $dsnSubmitting;
  $: formId =
    connectorType === "clickhouse-cloud" || connectionTab === "parameters"
      ? paramsFormId
      : dsnFormId;

  // Reset connectionTab if switching to Rill-managed or ClickHouse Cloud
  $: if (
    connectorType === "rill-managed" ||
    connectorType === "clickhouse-cloud"
  ) {
    connectionTab = "parameters";
  }

  // Reset errors when form is modified
  $: if (
    connectorType === "clickhouse-cloud" ||
    connectionTab === "parameters"
  ) {
    if ($paramsTainted) paramsError = null;
  } else if (connectionTab === "dsn") {
    if ($dsnTainted) dsnError = null;
  }

  // Emit the submitting state to the parent
  $: dispatch("submitting", { submitting });

  let prevConnectorType = connectorType;
  $: {
    // Switching to Rill-managed: set managed=true and clear other properties
    if (
      connectorType === "rill-managed" &&
      Object.keys($paramsForm).length > 1
    ) {
      paramsForm.update(() => ({ managed: true }), { taint: false });
      resetError();
    }
    // Switching to self-managed: restore defaults and set managed=false
    else if (
      prevConnectorType === "rill-managed" &&
      connectorType === "self-managed"
    ) {
      paramsForm.update(() => ({ ...initialFormValues, managed: false }), {
        taint: false,
      });
    }
    // Switching to ClickHouse Cloud: set specific defaults
    else if (
      prevConnectorType !== "clickhouse-cloud" &&
      connectorType === "clickhouse-cloud"
    ) {
      paramsForm.update(
        () => ({
          ...initialFormValues,
          managed: false,
          port: "8443",
          ssl: true,
        }),
        {
          taint: false,
        },
      );
    }
    prevConnectorType = connectorType;
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
    cancel: () => void;
    result: Extract<
      import("@sveltejs/kit").ActionResult,
      { type: "success" | "failure" }
    >;
  }) {
    if (!event.form.valid) return;
    const values = { ...event.form.data };

    // Ensure ClickHouse Cloud specific requirements are met
    if (connectorType === "clickhouse-cloud") {
      (values as any).ssl = true;
      (values as any).port = "8443";
    }

    try {
      await submitAddConnectorForm(queryClient, connector, values);
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
      if (
        connectorType === "clickhouse-cloud" ||
        connectionTab === "parameters"
      ) {
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

  $: properties = (() => {
    if (connectorType === "rill-managed") {
      return connector.sourceProperties ?? [];
    } else if (connectorType === "clickhouse-cloud") {
      // ClickHouse Cloud: only show specific properties
      return (connector.configProperties ?? []).filter((p) =>
        ["host", "port", "username", "password", "database"].includes(
          p.key ?? "",
        ),
      );
    } else {
      // Self-managed: show all config properties except dsn
      return (connector.configProperties ?? []).filter((p) =>
        connectionTab !== "dsn" ? p.key !== "dsn" : true,
      );
    }
  })();

  $: filteredProperties = properties.filter(
    (property) => !property.noPrompt && property.key !== "managed",
  );

  // TODO: move to utils.ts
  // Compute disabled state for the submit button
  // Refer to `runtime/drivers/clickhouse/clickhouse.go` for the required properties
  $: isSubmitDisabled = (() => {
    if (connectorType === "rill-managed") {
      // Rill-managed form: check all required properties including 'managed'
      for (const property of filteredProperties) {
        if (property.required) {
          const key = String(property.key);
          const value = $paramsForm[key];
          if (isEmpty(value) || $paramsErrors[key]?.length) return true;
        }
      }
      return false;
    } else if (connectionTab === "dsn") {
      // Self-managed or Cloud DSN form
      for (const property of dsnProperties) {
        if (property.required) {
          const key = String(property.key);
          const value = $dsnForm[key];
          if (isEmpty(value) || $dsnErrors[key]?.length) return true;
        }
      }
      return false;
    } else {
      // Parameters form: check required properties based on connector type
      for (const property of filteredProperties) {
        if (property.required && property.key !== "managed") {
          const key = String(property.key);
          const value = $paramsForm[key];
          if (isEmpty(value) || $paramsErrors[key]?.length) return true;
        }
      }

      // For ClickHouse Cloud, ensure SSL is enabled
      if (connectorType === "clickhouse-cloud") {
        if (!$paramsForm.ssl) return true;
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
      bind:value={connectorType}
      options={CONNECTOR_TYPE_OPTIONS}
    />
    {#if connectorType === "rill-managed"}
      <div class="mt-4">
        <InformationalField
          description="This option uses ClickHouse as an OLAP engine with Rill-managed infrastructure. No additional configuration is required - Rill will handle the setup and management of your ClickHouse instance."
        />
      </div>
    {:else if connectorType === "clickhouse-cloud"}
      <div class="my-4">
        <InformationalField
          description="Connect to your ClickHouse Cloud instance. SSL is automatically enabled and port is set to 8443 (HTTPS). You'll need your host, username, password, and database from the ClickHouse Cloud console."
        />
      </div>
    {/if}
  </div>

  {#if connectorType === "self-managed"}
    <Tabs bind:value={connectionTab} options={CONNECTION_TAB_OPTIONS}>
      <TabsContent value="parameters">
        <form
          id={paramsFormId}
          class="flex-grow overflow-y-auto"
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
          class="flex-grow overflow-y-auto"
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
        </form>
      </TabsContent>
    </Tabs>
  {:else if connectorType === "clickhouse-cloud"}
    <!-- ClickHouse Cloud: show parameters form directly without tabs -->
    <form
      id={paramsFormId}
      class="pb-5 flex-grow overflow-y-auto"
      use:paramsEnhance
      on:submit|preventDefault={paramsSubmit}
    >
      {#each filteredProperties as property (property.key)}
        {@const propertyKey = property.key ?? ""}
        {@const isPortField = propertyKey === "port"}
        {@const isSSLField = propertyKey === "ssl"}

        <!-- Skip SSL field for ClickHouse Cloud since it's always enabled -->
        {#if !isSSLField}
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
                disabled={isPortField}
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

            <!-- Show info about fixed values for ClickHouse Cloud -->
            {#if isPortField}
              <div class="mt-1 text-xs text-gray-600">
                Port is fixed to 8443 for ClickHouse Cloud (HTTPS)
              </div>
            {/if}
          </div>
        {/if}
      {/each}
    </form>
  {:else}
    <!-- Only managed form -->
    <form
      id={paramsFormId}
      class="flex-grow overflow-y-auto"
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
