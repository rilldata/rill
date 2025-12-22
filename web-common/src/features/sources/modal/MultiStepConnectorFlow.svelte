<script lang="ts">
  import type { V1ConnectorDriver } from "@rilldata/web-common/runtime-client";

  import AddDataFormSection from "./AddDataFormSection.svelte";
  import JSONSchemaFormRenderer from "../../templates/JSONSchemaFormRenderer.svelte";
  import { getInitialFormValuesFromProperties } from "../sourceUtils";
  import { connectorStepStore, setAuthMethod } from "./connectorStepStore";
  import {
    findRadioEnumKey,
    getRadioEnumOptions,
  } from "../../templates/schema-utils";
  import { getConnectorSchema } from "./connector-schemas";
  import { isMultiStepConnectorDisabled } from "./utils";
  import type { AddDataFormManager } from "./AddDataFormManager";
  import type { MultiStepFormSchema } from "../../templates/schemas/types";
  import type { ConnectorStepState } from "./connectorStepStore";
  import type { ConnectorDriverProperty } from "@rilldata/web-common/runtime-client";

  export let connector: V1ConnectorDriver;
  export let formManager: AddDataFormManager;
  // export let properties: ConnectorDriverProperty[];
  // export let filteredParamsProperties: ConnectorDriverProperty[];
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
  let previousAuthMethod: string | null = null;

  $: selectedAuthMethod = $selectedAuthMethodStore;

  // Initialize previousAuthMethod on first load
  $: if (previousAuthMethod === null && selectedAuthMethod) {
    previousAuthMethod = selectedAuthMethod;
  }

  // Initialize source step values from stored connector config.
  $: if (stepState.step === "source" && stepState.connectorConfig) {
    const sourceProperties = connector.sourceProperties ?? [];
    const initialValues = getInitialFormValuesFromProperties(sourceProperties);
    const combinedValues = {
      ...stepState.connectorConfig,
      ...initialValues,
    };
    paramsForm.update(() => combinedValues, { taint: false });
  }

  // Restore defaults (and persisted auth) when returning to connector step.
  // For schema-based connectors, JSONSchemaFormRenderer handles defaults
  $: if (stepState.step === "connector" && !activeSchema) {
    paramsForm.update(
      ($current) => {
        const base = getInitialFormValuesFromProperties(
          connector.configProperties ?? [],
        );
        return { ...base, ...$current };
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
  $: if (activeSchema && activeAuthInfo && stepState.step === "connector") {
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
            if ($form[authKey] !== fallback) {
              $form[authKey] = fallback;
            }
            return $form;
          }, { taint: false });
        }
      }
    }
  } else if (stepState.selectedAuthMethod && !activeAuthInfo) {
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

  // Clear form when auth method changes (e.g., switching from parameters to DSN).
  $: if (activeSchema && selectedAuthMethod !== previousAuthMethod && previousAuthMethod !== null && previousAuthMethod !== "") {
    const authKey = findRadioEnumKey(activeSchema);
    if (authKey) {
      // Get default values for the new auth method
      const defaults: Record<string, any> = {};

      // Set auth_method to the new value
      defaults[authKey] = selectedAuthMethod;

      // Add default values from schema for fields visible in this auth method
      if (activeSchema.properties) {
        for (const [key, prop] of Object.entries(activeSchema.properties)) {
          if (key === authKey) continue; // Already set

          // Check if this field is visible for the current auth method
          const visibleIf = prop["x-visible-if"];
          if (visibleIf && authKey in visibleIf) {
            const expectedValue = visibleIf[authKey];
            const matches = Array.isArray(expectedValue)
              ? expectedValue.includes(selectedAuthMethod)
              : expectedValue === selectedAuthMethod;

            if (matches && prop.default !== undefined) {
              defaults[key] = prop.default;
            }
          }
        }
      }

      // Update form with cleared values
      paramsForm.update(() => defaults, { taint: false });

      // Clear any form errors
      paramsErrors.update(() => ({}));
    }
  }

  // Track previous auth method for comparison
  $: if (selectedAuthMethod) {
    previousAuthMethod = selectedAuthMethod;
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
  $: currentMode = $paramsForm?.mode as string | undefined;
  $: primaryButtonLabel = formManager.getPrimaryButtonLabel({
    isConnectorForm: formManager.isConnectorForm,
    step: stepState.step,
    submitting,
    selectedAuthMethod: activeAuthMethod ?? selectedAuthMethod,
    mode: currentMode,
  });
  $: primaryLoadingCopy =
    stepState.step === "source"
      ? "Importing data..."
      : activeAuthMethod === "public"
        ? "Continuing..."
        : currentMode === "read"
          ? "Adding connector..."
          : "Testing connection...";
  $: formId = paramsFormId;
  $: shouldShowSkipLink = stepState.step === "connector";
</script>

{#if stepState.step === "connector"}
  <AddDataFormSection
    id={paramsFormId}
    enhance={paramsEnhance}
    onSubmit={paramsSubmit}
  >
    {#if activeSchema}
      <JSONSchemaFormRenderer
        schema={activeSchema}
        step="connector"
        form={paramsForm}
        errors={$paramsErrors}
        {onStringInputChange}
        {handleFileUpload}
      />
    {/if}
  </AddDataFormSection>
{:else}
  <AddDataFormSection
    id={paramsFormId}
    enhance={paramsEnhance}
    onSubmit={paramsSubmit}
  >
    {#if activeSchema}
      <JSONSchemaFormRenderer
        schema={activeSchema}
        step="source"
        form={paramsForm}
        errors={$paramsErrors}
        {onStringInputChange}
        {handleFileUpload}
      />
    {/if}
  </AddDataFormSection>
{/if}
