<script lang="ts">
  import type { ActionResult } from "@sveltejs/kit";
  import type { V1ConnectorDriver } from "@rilldata/web-common/runtime-client";

  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import JSONSchemaFormRenderer from "../../templates/JSONSchemaFormRenderer.svelte";
  import { getInitialFormValuesFromProperties } from "../sourceUtils";
  import {
    connectorStepStore,
    setAuthMethod,
    type ConnectorStepState,
  } from "./connectorStepStore";
  import {
    findRadioEnumKey,
    getRadioEnumOptions,
  } from "../../templates/schema-utils";
  import { getConnectorSchema } from "./connector-schemas";
  import { isMultiStepConnectorDisabled } from "./utils";
  import { AddDataFormManager } from "./AddDataFormManager";
  import type { MultiStepFormSchema } from "../../templates/schemas/types";

  export let connector: V1ConnectorDriver;
  export let onClose: () => void;
  export let onBack: () => void;

  // Outputs bound by parent
  export let isSubmitDisabled = true;
  export let primaryButtonLabel = "";
  export let primaryLoadingCopy = "";
  export let formId = "";
  export let yamlPreview = "";
  export let yamlPreviewTitle = "Connector preview";
  export let showSaveAnyway = false;
  export let saveAnywayLoading = false;
  export let saveAnywayHandler: () => Promise<void> = async () => {};
  export let paramsError: string | null = null;
  export let paramsErrorDetails: string | undefined = undefined;
  export let isSubmitting = false;
  export let shouldShowSkipLink = false;
  export let handleBack: () => void = () => onBack();
  export let handleSkip: () => void = () => {};

  let handleOnUpdateFn: <
    T extends Record<string, unknown>,
    M = any,
    In extends Record<string, unknown> = T,
  >(event: {
    form: import("sveltekit-superforms").SuperValidated<T, M, In>;
    formEl: HTMLFormElement;
    cancel: () => void;
    result: Extract<ActionResult, { type: "success" | "failure" }>;
  }) => Promise<void> = async (_event) => {};
  const forwardOnUpdate = (e: any) => handleOnUpdateFn(e);
  let handleOnUpdate = handleOnUpdateFn;

  const formManager = new AddDataFormManager({
    connector,
    formType: "connector",
    onParamsUpdate: forwardOnUpdate,
    onDsnUpdate: (_e: any) => {},
    getSelectedAuthMethod: () => activeAuthMethod ?? undefined,
  });

  const paramsFormId = formManager.paramsFormId;
  const {
    form: paramsForm,
    errors: paramsErrors,
    enhance: paramsEnhance,
    tainted: paramsTainted,
    submit: paramsSubmit,
    submitting: paramsSubmitting,
  } = formManager.params;

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
  let activeAuthMethod: string | null = null;

  $: selectedAuthMethod = $selectedAuthMethodStore;

  // Initialize (and clear) source step values whenever we enter the source step.
  $: if (stepState.step === "source") {
    const sourceProperties = connector.sourceProperties ?? [];
    const initialValues = getInitialFormValuesFromProperties(sourceProperties);
    const combinedValues = {
      ...initialValues,
      ...(stepState.connectorConfig ?? {}),
    };
    paramsForm.update(() => combinedValues, { taint: false });
  }

  // Restore defaults (and persisted auth) when returning to connector step.
  $: if (stepState.step === "connector") {
    paramsForm.update(
      ($current) => {
        const base = getInitialFormValuesFromProperties(
          connector.configProperties ?? [],
        );
        if (activeSchema) {
          const authKey = findRadioEnumKey(activeSchema);
          const persisted = stepState.selectedAuthMethod;
          if (authKey && persisted) {
            base[authKey] = persisted;
          }
        }
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

  // Clear Save Anyway whenever we leave the connector step so it can't bleed into the model step.
  $: if (stepState.step === "source" && showSaveAnyway) {
    showSaveAnyway = false;
  }

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
    submitting: isSubmitting,
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
  $: handleBack = () => formManager.handleBack(onBack);
  $: handleSkip = () => formManager.handleSkip();

  // YAML preview lives here for multi-step connectors.
  $: yamlPreview = formManager.computeYamlPreview({
    connectionTab: "parameters",
    onlyDsn: formManager.hasOnlyDsn,
    filteredParamsProperties: formManager.filteredParamsProperties,
    filteredDsnProperties: formManager.filteredDsnProperties,
    stepState,
    isMultiStepConnector: true,
    isConnectorForm: true,
    paramsFormValues: $paramsForm,
    dsnFormValues: {},
  });
  $: yamlPreviewTitle =
    stepState.step === "connector" ? "Connector preview" : "Model preview";
  // Submission wiring
  $: isSubmitting = $paramsSubmitting;
  $: handleOnUpdate = formManager.makeOnUpdate({
    onClose,
    queryClient,
    getConnectionTab: () => "parameters",
    getSelectedAuthMethod: () => activeAuthMethod || undefined,
    setParamsError: (message: string | null, details?: string) => {
      paramsError = message;
      paramsErrorDetails = details;
    },
    setDsnError: (_message: string | null, _details?: string) => {},
    setShowSaveAnyway: (value: boolean) => {
      showSaveAnyway = value;
    },
  });
  $: handleOnUpdateFn = handleOnUpdate;

  // Reset errors when form is modified
  $: if ($paramsTainted) {
    paramsError = null;
    paramsErrorDetails = undefined;
    // Clear field-level errors so a corrected input can resubmit.
    (paramsErrors as any)?.set?.({});
  }

  // Save Anyway handler (parent renders button, uses this handler)
  async function handleSaveAnyway() {
    saveAnywayLoading = true;
    const result = await formManager.saveConnectorAnyway({
      queryClient,
      values: $paramsForm,
    });
    if (result.ok) {
      onClose();
    } else {
      paramsError = result.message;
      paramsErrorDetails = result.details;
    }
    saveAnywayLoading = false;
  }

  $: saveAnywayHandler = handleSaveAnyway;
  const onStringInputChange = (event: Event) =>
    formManager.onStringInputChange(
      event,
      $paramsTainted as Record<string, boolean> | null | undefined,
    );
  const handleFileUpload = (file: File) => formManager.handleFileUpload(file);
</script>

<JSONSchemaFormRenderer
  formId={paramsFormId}
  enhance={paramsEnhance}
  onSubmit={paramsSubmit}
  className="pb-5 flex-grow overflow-y-auto"
  schema={activeSchema}
  step={stepState.step}
  form={paramsForm}
  errors={$paramsErrors}
  {onStringInputChange}
  {handleFileUpload}
/>
