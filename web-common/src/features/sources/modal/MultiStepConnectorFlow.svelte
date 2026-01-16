<script lang="ts">
  import type { V1ConnectorDriver } from "@rilldata/web-common/runtime-client";

  import AddDataFormSection from "./AddDataFormSection.svelte";
  import JSONSchemaFormRenderer from "../../templates/JSONSchemaFormRenderer.svelte";
  import { getInitialFormValuesFromProperties } from "../sourceUtils";
  import { connectorStepStore, setAuthMethod } from "./connectorStepStore";
  import {
    findRadioEnumKey,
    getRadioEnumOptions,
    getSchemaInitialValues,
    isStepMatch,
  } from "../../templates/schema-utils";
  import { getConnectorSchema } from "./connector-schemas";
  import { isMultiStepConnectorDisabled } from "./utils";
  import type { AddDataFormManager } from "./AddDataFormManager";
  import type { MultiStepFormSchema } from "../../templates/schemas/types";
  import type { ConnectorStepState } from "./connectorStepStore";

  export let connector: V1ConnectorDriver;
  export let formManager: AddDataFormManager;
  export let paramsForm: any;
  export let paramsErrors: any;
  export let paramsEnhance: any;
  export let paramsSubmit: (
    event: Event & {
      currentTarget: EventTarget & HTMLFormElement;
    },
  ) => unknown;
  export let paramsFormId: string;
  export let onStringInputChange: (e: Event) => void;
  export let handleFileUpload: (file: File) => Promise<string>;
  export let submitting: boolean;

  // Outputs bound by parent
  export let activeAuthMethod: string | null = null;
  export let primaryButtonLabel = "";
  export let primaryLoadingCopy = "";
  export let isSubmitDisabled = true;
  export let formId = paramsFormId;
  export let shouldShowSkipLink = false;

  const selectedAuthMethodStore = {
    subscribe: (run: (value: string) => void) =>
      connectorStepStore.subscribe((state) =>
        run(state.selectedAuthMethod ?? ""),
      ),
    set: (method: string) => setAuthMethod(method || null),
  };

  $: stepState = $connectorStepStore as ConnectorStepState;
  let activeSchema: MultiStepFormSchema | null = null;
  let activeAuthInfo: ReturnType<typeof getRadioEnumOptions> | null = null;
  let selectedAuthMethod = "";

  $: selectedAuthMethod = $selectedAuthMethodStore;

  // Initialize source step values from stored connector config.
  $: if (stepState.step === "source" && stepState.connectorConfig) {
    const schema = connector.name
      ? getConnectorSchema(connector.name) || null
      : null;
    const initialValues = schema
      ? getSchemaInitialValues(schema, { step: "source" })
      : getInitialFormValuesFromProperties(connector.sourceProperties ?? []);
    const combinedValues = {
      ...stepState.connectorConfig,
      ...initialValues,
    };
    paramsForm.update(() => combinedValues, { taint: false });
  }

  // Restore defaults (and persisted auth) when returning to connector step.
  // Also drop any source-step fields so previous model inputs can't resurface.
  $: if (stepState.step === "connector") {
    const schema = connector.name
      ? getConnectorSchema(connector.name) || null
      : null;
    paramsForm.update(
      ($current) => {
        const base = schema
          ? getSchemaInitialValues(schema, { step: "connector" })
          : getInitialFormValuesFromProperties(
              connector.configProperties ?? [],
            );
        if (activeSchema) {
          const authKey = findRadioEnumKey(activeSchema);
          const persisted = stepState.selectedAuthMethod;
          if (authKey && persisted) {
            base[authKey] = persisted;
          }
        }
        // Drop any source-step fields so a previous source submission (e.g., GCS
        // URI and model name) doesn't leak into a fresh connector step.
        const filteredCurrent =
          schema?.properties && schema.properties
            ? Object.fromEntries(
                Object.entries($current).filter(([key]) =>
                  isStepMatch(schema, key, "connector"),
                ),
              )
            : $current;
        return { ...base, ...filteredCurrent };
      },
      { taint: false },
    );
  }

  // Determine schema and auth options for the connector.
  $: activeSchema = connector.name
    ? getConnectorSchema(connector.name) || null
    : null;
  $: activeAuthInfo = activeSchema ? getRadioEnumOptions(activeSchema) : null;

  // Ensure we always have a valid auth method selection for the active schema.
  $: if (activeSchema && activeAuthInfo) {
    const options = activeAuthInfo.options ?? [];
    const fallback = activeAuthInfo.defaultValue || options[0]?.value || null;
    const authKey = activeAuthInfo.key || findRadioEnumKey(activeSchema);
    const hasValidSelection = options.some(
      (option) => option.value === stepState.selectedAuthMethod,
    );
    if (!hasValidSelection) {
      if (fallback !== stepState.selectedAuthMethod) {
        setAuthMethod(fallback ?? null);
        if (fallback && authKey) {
          paramsForm.update(($form) => {
            if ($form[authKey] !== fallback) $form[authKey] = fallback;
            return $form;
          });
        }
      }
    }
  } else if (stepState.selectedAuthMethod) {
    setAuthMethod(null);
  }

  // Keep auth method store aligned with the form selection.
  $: if (activeSchema) {
    const authKey = findRadioEnumKey(activeSchema);
    if (authKey) {
      const currentValue = $paramsForm?.[authKey] as string | undefined;
      const normalized = currentValue ? String(currentValue) : null;
      if (normalized !== (stepState.selectedAuthMethod ?? null)) {
        setAuthMethod(normalized);
      }
    }
  }

  // Active auth method for UI (button labels/loading).
  $: activeAuthMethod = (() => {
    if (!(activeSchema && paramsForm)) return selectedAuthMethod;
    const authKey = findRadioEnumKey(activeSchema);
    if (authKey && $paramsForm?.[authKey] != null) {
      return String($paramsForm[authKey]);
    }
    return selectedAuthMethod;
  })();

  // CTA and disable state for multi-step connectors.
  $: isSubmitDisabled = isMultiStepConnectorDisabled(
    activeSchema,
    $paramsForm,
    $paramsErrors,
    stepState.step,
  );
  $: primaryButtonLabel = formManager.getPrimaryButtonLabel({
    isConnectorForm: formManager.isConnectorForm,
    step: stepState.step,
    submitting,
    selectedAuthMethod: activeAuthMethod ?? selectedAuthMethod,
  });
  $: primaryLoadingCopy =
    stepState.step === "source"
      ? "Importing data..."
      : activeAuthMethod === "public"
        ? "Continuing..."
        : "Testing connection...";
  $: formId = paramsFormId;
  $: shouldShowSkipLink = stepState.step === "connector";
</script>

<AddDataFormSection
  id={paramsFormId}
  enhance={paramsEnhance}
  onSubmit={paramsSubmit}
>
  <JSONSchemaFormRenderer
    schema={activeSchema}
    step={stepState.step}
    form={paramsForm}
    errors={$paramsErrors}
    {onStringInputChange}
    {handleFileUpload}
  />
</AddDataFormSection>
