<script lang="ts">
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
  import {
    inferSourceName,
    prepareSourceFormData,
    compileSourceYAML,
  } from "../sourceUtils";
  import { humanReadableErrorMessage } from "../errors/errors";
  import {
    submitAddConnectorForm,
    submitAddSourceForm,
  } from "./submitAddDataForm";
  import type { AddDataFormType, ConnectorType } from "./types";
  import { dsnSchema, getYupSchema } from "./yupSchemas";
  import { isEmpty } from "./utils";
  import { type ClickHouseConnectorType } from "./constants";
  import { getInitialFormValuesFromProperties } from "../sourceUtils";
  import { compileConnectorYAML } from "../../connectors/code-utils";
  import LeftFooter from "./LeftFooter.svelte";
  import RightSidePanel from "./RightSidePanel.svelte";
  import FormRenderer from "./FormRenderer.svelte";

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

  // Filter properties based on connector type
  const filteredParamsProperties = (() => {
    // FIXME: https://linear.app/rilldata/issue/APP-408/support-ducklake-in-the-ui
    if (connector.name === "duckdb") {
      return properties.filter(
        (property) => property.key !== "attach" && property.key !== "mode",
      );
    }
    // For other connectors, filter out noPrompt properties
    return properties.filter((property) => !property.noPrompt);
  })();

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
  const hasDsnFormOption =
    isConnectorForm &&
    connector.configProperties?.some((property) => property.key === "dsn") &&
    connector.configProperties?.some((property) => property.key !== "dsn");
  const dsnFormId = `add-data-${connector.name}-dsn-form`;
  const dsnProperties =
    connector.configProperties?.filter((property) => property.key === "dsn") ??
    [];

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

  // ClickHouse-specific variables
  let clickhouseConnectorType: ClickHouseConnectorType;

  // Initialize ClickHouse connector type
  $: if (typeof clickhouseConnectorType === "undefined") {
    clickhouseConnectorType = "self-hosted";
  }

  // Reset connectionTab if switching to Rill-managed
  $: if (clickhouseConnectorType === "rill-managed") {
    connectionTab = "parameters";
  }

  // ClickHouse-specific forms and state
  const clickhouseSchema = yup(getYupSchema["clickhouse"]);
  const clickhouseInitialFormValues = getInitialFormValuesFromProperties(
    connector.configProperties ?? [],
  );
  const clickhouseParamsFormId = `add-clickhouse-data-${connector.name}-form`;
  const {
    form: clickhouseParamsForm,
    errors: clickhouseParamsErrors,
    enhance: clickhouseParamsEnhance,
    tainted: clickhouseParamsTainted,
    submit: clickhouseParamsSubmit,
    submitting: clickhouseParamsSubmitting,
  } = superForm(clickhouseInitialFormValues, {
    SPA: true,
    validators: clickhouseSchema,
    onUpdate: handleClickHouseOnUpdate,
    resetForm: false,
  });

  const clickhouseDsnFormId = `add-clickhouse-data-${connector.name}-dsn-form`;
  const clickhouseDsnProperties =
    connector.configProperties?.filter((property) => property.key === "dsn") ??
    [];
  const clickhouseDsnYupSchema = yup(dsnSchema);
  const {
    form: clickhouseDsnForm,
    errors: clickhouseDsnErrors,
    enhance: clickhouseDsnEnhance,
    tainted: clickhouseDsnTainted,
    submit: clickhouseDsnSubmit,
    submitting: clickhouseDsnSubmitting,
  } = superForm(defaults(clickhouseDsnYupSchema), {
    SPA: true,
    validators: clickhouseDsnYupSchema,
    onUpdate: handleClickHouseOnUpdate,
    resetForm: false,
  });

  // ClickHouse-specific computed properties
  $: clickhouseProperties = (() => {
    if (clickhouseConnectorType === "rill-managed") {
      return connector.sourceProperties ?? [];
    } else if (clickhouseConnectorType === "clickhouse-cloud") {
      // ClickHouse Cloud: show all config properties except dsn (same as self-hosted)
      return (connector.configProperties ?? []).filter((p) =>
        connectionTab !== "dsn" ? p.key !== "dsn" : true,
      );
    } else {
      // Self-managed: show all config properties except dsn
      return (connector.configProperties ?? []).filter((p) =>
        connectionTab !== "dsn" ? p.key !== "dsn" : true,
      );
    }
  })();

  $: clickhouseFilteredProperties = clickhouseProperties.filter(
    (property) => !property.noPrompt && property.key !== "managed",
  );

  function onClickHouseStringInputChange(event: Event) {
    const target = event.target as HTMLInputElement;
    const { name, value } = target;
    if (name === "path") {
      if ($clickhouseParamsTainted?.name) return;
      const nameVal = inferSourceName(connector, value);
      if (nameVal)
        clickhouseParamsForm.update(
          ($form) => {
            $form.name = nameVal;
            return $form;
          },
          { taint: false },
        );
    }
  }

  async function handleClickHouseOnUpdate<
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
    // Only apply these when using parameters tab, not DSN tab
    if (
      clickhouseConnectorType === "clickhouse-cloud" &&
      connectionTab === "parameters"
    ) {
      (values as any).ssl = true;
      (values as any).port = "8443";
    }

    try {
      if (formType === "source") {
        await submitAddSourceForm(queryClient, connector, values);
      } else {
        await submitAddConnectorForm(queryClient, connector, values);
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
      } else if (e?.message) {
        error = e.message;
        details = undefined;
      } else {
        error = "Unknown error";
        details = undefined;
      }
      if (connectionTab === "parameters") {
        clickhouseError = error;
        clickhouseErrorDetails = details;
      } else if (connectionTab === "dsn") {
        clickhouseError = error;
        clickhouseErrorDetails = details;
      }
    }
  }

  // Helper function to check if connector only has DSN (no tabs)
  function hasOnlyDsn(): boolean {
    return (
      isConnectorForm &&
      Boolean(
        connector.configProperties?.some((property) => property.key === "dsn"),
      ) &&
      !Boolean(
        connector.configProperties?.some((property) => property.key !== "dsn"),
      )
    );
  }

  // Compute disabled state for the submit button
  $: isSubmitDisabled = (() => {
    if (hasOnlyDsn() || connectionTab === "dsn") {
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

  $: formId = (() => {
    if (hasOnlyDsn() || connectionTab === "dsn") {
      return dsnFormId;
    } else {
      return paramsFormId;
    }
  })();

  $: submitting = (() => {
    if (hasOnlyDsn() || connectionTab === "dsn") {
      return $dsnSubmitting;
    } else {
      return $paramsSubmitting;
    }
  })();

  // Reset errors when form is modified
  $: (() => {
    if (hasOnlyDsn() || connectionTab === "dsn") {
      if ($dsnTainted) dsnError = null;
    } else {
      if ($paramsTainted) paramsError = null;
    }
  })();

  // Clear errors when switching tabs
  $: (() => {
    if (hasDsnFormOption) {
      if (connectionTab === "dsn") {
        paramsError = null;
        paramsErrorDetails = undefined;
      } else {
        dsnError = null;
        dsnErrorDetails = undefined;
      }
    }
  })();

  // Emit the submitting state to the parent
  $: dispatch("submitting", { submitting });

  // Computed properties for LeftFooter
  $: footerDisabled = (() => {
    if (connector.name === "clickhouse") {
      // ClickHouse-specific logic
      if (clickhouseConnectorType === "rill-managed") {
        return $clickhouseParamsSubmitting;
      } else {
        return connectionTab === "parameters"
          ? $clickhouseParamsSubmitting
          : $clickhouseDsnSubmitting;
      }
    } else {
      return submitting || isSubmitDisabled;
    }
  })();

  $: footerLoading = (() => {
    if (connector.name === "clickhouse") {
      return connectionTab === "parameters"
        ? $clickhouseParamsSubmitting
        : $clickhouseDsnSubmitting;
    } else {
      return submitting;
    }
  })();

  $: footerLoadingCopy = (() => {
    if (connector.name === "clickhouse") {
      return "Connecting...";
    } else {
      return "Testing connection...";
    }
  })();

  $: footerFormId = (() => {
    if (connector.name === "clickhouse") {
      return connectionTab === "parameters"
        ? clickhouseParamsFormId
        : clickhouseDsnFormId;
    } else {
      return formId;
    }
  })();

  $: footerSubmitButtonText = (() => {
    if (connector.name === "clickhouse") {
      if (clickhouseConnectorType === "rill-managed") {
        if ($clickhouseParamsSubmitting) {
          return "Connecting...";
        } else {
          return "Connect";
        }
      } else if (
        connectionTab === "parameters"
          ? $clickhouseParamsSubmitting
          : $clickhouseDsnSubmitting
      ) {
        return "Testing connection...";
      } else {
        return "Test and Connect";
      }
    } else if (isConnectorForm) {
      if (submitting) {
        return "Testing connection...";
      } else {
        return "Test and Connect";
      }
    } else {
      return "Test and Add data";
    }
  })();

  function getConnectorYamlPreview(values: Record<string, unknown>) {
    return compileConnectorYAML(connector, values, {
      fieldFilter: (property) => {
        // When in DSN mode, don't filter out noPrompt properties
        // because the DSN field itself might have noPrompt: true
        if (hasOnlyDsn() || connectionTab === "dsn") {
          return true; // Show all DSN properties
        }
        return !property.noPrompt;
      },
      orderedProperties:
        hasOnlyDsn() || connectionTab === "dsn"
          ? filteredDsnProperties
          : filteredParamsProperties,
    });
  }

  function getSourceYamlPreview(values: Record<string, unknown>) {
    const [rewrittenConnector, rewrittenFormValues] = prepareSourceFormData(
      connector,
      values,
    );

    // Check if the connector was rewritten to DuckDB
    const isRewrittenToDuckDb = rewrittenConnector.name === "duckdb";

    if (isRewrittenToDuckDb) {
      return compileSourceYAML(rewrittenConnector, rewrittenFormValues);
    } else {
      return getConnectorYamlPreview(rewrittenFormValues);
    }
  }

  function getClickHouseYamlPreview(
    values: Record<string, unknown>,
    connectorType: ClickHouseConnectorType,
  ) {
    // For rill-managed, create a simple YAML with just the managed field
    if (connectorType === "rill-managed") {
      return `type: clickhouse
managed: true`;
    }

    // Ensure ClickHouse Cloud specific requirements are met in preview
    const previewValues = { ...values } as Record<string, unknown>;
    if (connectorType === "clickhouse-cloud") {
      previewValues.ssl = true;
      previewValues.port = "8443";
    }

    return compileConnectorYAML(connector, previewValues, {
      fieldFilter: (property) => {
        // When in DSN mode, don't filter out noPrompt properties
        // because the DSN field itself might have noPrompt: true
        if (connectionTab === "dsn") {
          return true; // Show all DSN properties
        }
        return !property.noPrompt;
      },
      orderedProperties:
        connectionTab === "dsn"
          ? clickhouseDsnProperties
          : clickhouseFilteredProperties,
    });
  }

  // ClickHouse-specific YAML preview
  $: clickhouseYamlPreview = (() => {
    if (connector.name === "clickhouse") {
      // For rill-managed, use minimal values since no form fields are shown
      if (clickhouseConnectorType === "rill-managed") {
        const minimalValues = {
          // Include any required fields that might be needed
          ...getInitialFormValuesFromProperties(
            connector.sourceProperties ?? [],
          ),
        };
        return getClickHouseYamlPreview(minimalValues, clickhouseConnectorType);
      }

      // For self-hosted and clickhouse-cloud, use actual form values
      const values =
        connectionTab === "dsn" ? $clickhouseDsnForm : $clickhouseParamsForm;
      return getClickHouseYamlPreview(values, clickhouseConnectorType);
    }
    return "";
  })();

  // General YAML preview
  $: generalYamlPreview = (() => {
    if (connector.name === "clickhouse") {
      return clickhouseYamlPreview;
    }

    const values =
      hasOnlyDsn() || connectionTab === "dsn" ? $dsnForm : $paramsForm;

    if (isConnectorForm) {
      // Connector form
      return getConnectorYamlPreview(values);
    } else {
      // Source form
      return getSourceYamlPreview(values);
    }
  })();

  $: yamlPreview = generalYamlPreview;

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
        await submitAddConnectorForm(queryClient, connector, values);
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
      } else if (e?.message) {
        error = e.message;
        details = undefined;
      } else {
        error = "Unknown error";
        details = undefined;
      }

      // Keep error state for each form - match the display logic
      if (hasOnlyDsn() || connectionTab === "dsn") {
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
      class="flex flex-col flex-grow {[
        'clickhouse',
        'snowflake',
        'salesforce',
      ].includes(connector.name ?? '')
        ? 'max-h-[38.5rem] min-h-[38.5rem]'
        : 'max-h-[34.5rem] min-h-[34.5rem]'} overflow-y-auto p-6"
    >
      <FormRenderer
        {connector}
        {formType}
        bind:connectionTab
        hasDsnFormOption={Boolean(hasDsnFormOption)}
        {paramsFormId}
        {dsnFormId}
        {paramsForm}
        {dsnForm}
        {paramsErrors}
        {dsnErrors}
        {paramsEnhance}
        {dsnEnhance}
        {paramsSubmit}
        {dsnSubmit}
        {filteredParamsProperties}
        {filteredDsnProperties}
        {onStringInputChange}
        {clickhouseConnectorType}
        {clickhouseParamsForm}
        {clickhouseDsnForm}
        {clickhouseParamsErrors}
        {clickhouseDsnErrors}
        {clickhouseParamsEnhance}
        {clickhouseDsnEnhance}
        {clickhouseParamsSubmit}
        {clickhouseDsnSubmit}
        {clickhouseParamsFormId}
        {clickhouseDsnFormId}
        {clickhouseDsnProperties}
        {clickhouseFilteredProperties}
        {onClickHouseStringInputChange}
      />
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
    {isSourceForm}
    {dsnError}
    {paramsError}
    {clickhouseError}
    {dsnErrorDetails}
    {paramsErrorDetails}
    {clickhouseErrorDetails}
    hasOnlyDsn={hasOnlyDsn()}
    {connectionTab}
    {yamlPreview}
    {copied}
    {copyYamlPreview}
  />
</div>
