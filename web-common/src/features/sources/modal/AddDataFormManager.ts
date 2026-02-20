import type { SuperValidated } from "sveltekit-superforms";
import type { Writable } from "svelte/store";
import type { V1ConnectorDriver } from "@rilldata/web-common/runtime-client";
import type { AddDataFormType } from "./types";
import { getValidationSchemaForConnector } from "./FormValidation";
import { inferModelNameFromSQL, inferSourceName } from "../sourceUtils";
import {
  submitAddConnectorForm,
  submitAddSourceForm,
} from "./submitAddDataForm";
import { normalizeConnectorError } from "./utils";
import {
  getConnectorSchema,
  getFormHeight,
  hasExplorerStep as hasExplorerStepSchema,
  isMultiStepConnector as isMultiStepConnectorSchema,
} from "./connector-schemas";
import {
  connectorStepStore,
  setConnectorConfig,
  setConnectorInstanceName,
  setStep,
  type ConnectorStepState,
} from "./connectorStepStore";
import { get } from "svelte/store";
import { compileConnectorYAML } from "../../connectors/code-utils";
import type { ActionResult } from "@sveltejs/kit";
import type { QueryClient } from "@tanstack/query-core";
import {
  filterSchemaValuesForSubmit,
  findRadioEnumKey,
  getSchemaFieldMetaList,
  getSchemaSecretKeys,
  getSchemaStringKeys,
} from "../../templates/schema-utils";
import type { ButtonLabels } from "../../templates/schemas/types";
import { processFileContent } from "../../templates/file-encoding";
import { generateTemplate } from "./generate-template";
import { runtime } from "../../../runtime-client/runtime-store";

type FormData = Record<string, unknown>;
// Use unknown to be compatible with superforms' complex ValidationErrors type
type ValidationErrors = Record<string, unknown>;

type SuperFormUpdateOptions = {
  taint?: boolean;
};

type FormStore = Writable<FormData> & {
  update: (
    updater: (value: FormData) => FormData,
    options?: SuperFormUpdateOptions,
  ) => void;
};

type ErrorsStore = Writable<ValidationErrors> & {
  set: (errors: ValidationErrors) => void;
  update: (updater: (errors: ValidationErrors) => ValidationErrors) => void;
};

const BUTTON_LABELS = {
  public: { idle: "Continue", submitting: "Continuing..." },
  connector: { idle: "Test and Connect", submitting: "Testing connection..." },
  source: { idle: "Import Data", submitting: "Importing data..." },
};

export class AddDataFormManager {
  formHeight: string;
  formId: string;

  // Form stores (passed in from caller)
  private formStore: FormStore;
  private errorsStore: ErrorsStore;
  private connector: V1ConnectorDriver;
  private formType: AddDataFormType;
  private schemaName: string;

  // Centralized error normalization for this manager
  private normalizeError(e: unknown): { message: string; details?: string } {
    return normalizeConnectorError(this.connector.name ?? "", e);
  }

  private getSelectedAuthMethod?: () => string | undefined;
  // Keep only fields that belong to a given schema step. Prevents source-step
  // values (e.g., URI/model) from leaking into connector state when we persist.
  private filterValuesForStep(
    values: Record<string, unknown>,
    step: "connector" | "source" | "explorer",
  ): Record<string, unknown> {
    const schema = getConnectorSchema(this.schemaName);
    if (!schema?.properties) return values;
    return filterSchemaValuesForSubmit(schema, values, { step });
  }

  constructor(args: {
    connector: V1ConnectorDriver;
    formType: AddDataFormType;
    formStore: FormStore;
    errorsStore: ErrorsStore;
    getSelectedAuthMethod?: () => string | undefined;
    schemaName?: string; // Override connector.name for schema/validation lookup
  }) {
    const {
      connector,
      formType,
      formStore,
      errorsStore,
      getSelectedAuthMethod,
      schemaName,
    } = args;
    this.connector = connector;
    this.formType = formType;
    this.formStore = formStore;
    this.errorsStore = errorsStore;
    this.getSelectedAuthMethod = getSelectedAuthMethod;

    // Use schemaName if provided, otherwise fall back to connector.name
    this.schemaName = schemaName ?? connector.name ?? "";
    const effectiveSchemaName = this.schemaName;

    // IDs
    this.formId = `add-data-${effectiveSchemaName}-form`;

    const schema = getConnectorSchema(effectiveSchemaName);

    // Layout height (derived from schema metadata)
    this.formHeight = getFormHeight(schema);
  }

  get isSourceForm(): boolean {
    return this.formType === "source";
  }

  get isConnectorForm(): boolean {
    return this.formType === "connector";
  }

  get isMultiStepConnector(): boolean {
    const schema = getConnectorSchema(this.schemaName);
    return isMultiStepConnectorSchema(schema);
  }

  get hasExplorerStep(): boolean {
    const schema = getConnectorSchema(this.schemaName);
    return hasExplorerStepSchema(schema);
  }

  /**
   * Determines whether the "Save Anyway" button should be shown for the current submission.
   */
  private shouldShowSaveAnywayButton(args: {
    isConnectorForm: boolean;
    event?:
      | {
          result?: Extract<ActionResult, { type: "success" | "failure" }>;
        }
      | undefined;
    stepState: ConnectorStepState | undefined;
    selectedAuthMethod?: string;
  }): boolean {
    const { isConnectorForm, event, stepState, selectedAuthMethod } = args;

    // Only show for connector forms (not sources)
    if (!isConnectorForm) return false;

    // Need a submission result to show the button
    if (!event?.result) return false;

    // Multi-step connectors: don't show on source/explorer step (final step)
    if (stepState?.step === "source" || stepState?.step === "explorer")
      return false;

    // Public auth bypasses connection test, so no "Save Anyway" needed
    if (stepState?.step === "connector" && selectedAuthMethod === "public")
      return false;

    return true;
  }

  handleSkip(): void {
    const stepState = get(connectorStepStore) as ConnectorStepState;
    if (!this.isMultiStepConnector || stepState.step !== "connector") return;
    setConnectorConfig({});
    setConnectorInstanceName(null);
    setStep("source");
  }

  handleBack(onBack: () => void): void {
    const stepState = get(connectorStepStore) as ConnectorStepState;
    if (this.isMultiStepConnector && stepState.step === "source") {
      setStep("connector");
    } else if (this.hasExplorerStep && stepState.step === "explorer") {
      setStep("connector");
    } else {
      onBack();
    }
  }

  getPrimaryButtonLabel(args: {
    isConnectorForm: boolean;
    step: "connector" | "source" | string;
    submitting: boolean;
    schemaButtonLabels?: ButtonLabels | null;
    selectedAuthMethod?: string;
  }): string {
    const {
      isConnectorForm,
      step,
      submitting,
      schemaButtonLabels,
      selectedAuthMethod,
    } = args;
    const isStepFlowConnector =
      this.isMultiStepConnector || this.hasExplorerStep;
    const isOnConnectorStep = step === "connector";
    const isOnSourceOrExplorerStep = step === "source" || step === "explorer";

    // Schema-provided labels override defaults (e.g. rill-managed ClickHouse)
    if (schemaButtonLabels && isOnConnectorStep) {
      return submitting ? schemaButtonLabels.loading : schemaButtonLabels.idle;
    }

    if (isConnectorForm) {
      // Step 1 of multi-step: "Test and Connect" or "Continue" for public auth
      if (isStepFlowConnector && isOnConnectorStep) {
        const labels =
          selectedAuthMethod === "public"
            ? BUTTON_LABELS.public
            : BUTTON_LABELS.connector;
        return submitting ? labels.submitting : labels.idle;
      }
      // Step 2 of multi-step: "Import Data"
      if (isStepFlowConnector && isOnSourceOrExplorerStep) {
        return submitting
          ? BUTTON_LABELS.source.submitting
          : BUTTON_LABELS.source.idle;
      }
      // Single-step connector form
      return submitting
        ? BUTTON_LABELS.connector.submitting
        : BUTTON_LABELS.connector.idle;
    }

    // Source-only form (no connector step)
    return "Test and Add data";
  }

  makeOnUpdate(args: {
    onClose: () => void;
    queryClient: QueryClient;
    getSelectedAuthMethod?: () => string | undefined;
    setParamsError: (message: string | null, details?: string) => void;
    setShowSaveAnyway?: (value: boolean) => void;
  }) {
    const {
      onClose,
      queryClient,
      getSelectedAuthMethod,
      setParamsError,
      setShowSaveAnyway,
    } = args;
    const connector = this.connector;
    const schema = getConnectorSchema(this.schemaName);
    const isMultiStep = isMultiStepConnectorSchema(schema);
    const isExplorer = hasExplorerStepSchema(schema);
    const isStepFlowConnector = isMultiStep || isExplorer;
    const isConnectorForm = this.formType === "connector";

    return async (event: {
      form: SuperValidated<FormData, string, FormData>;
      result?: Extract<ActionResult, { type: "success" | "failure" }>;
      cancel?: () => void;
    }) => {
      const values = event.form.data;
      const stepState = get(connectorStepStore) as ConnectorStepState;
      const isOnSourceOrExplorerStep =
        stepState.step === "source" || stepState.step === "explorer";
      const isOnConnectorStep = stepState.step === "connector";

      // Resolve the auth method from form values or the parent component's state
      const authKey = schema ? findRadioEnumKey(schema) : null;
      const selectedAuthMethod =
        (authKey && values?.[authKey] != null
          ? String(values[authKey])
          : undefined) ||
        getSelectedAuthMethod?.() ||
        "";
      const isPublicAuth = selectedAuthMethod === "public";

      // Filter form values to only include fields for the current step
      const stepForFilter =
        isStepFlowConnector && isOnSourceOrExplorerStep
          ? stepState.step
          : this.formType === "source"
            ? "source"
            : "connector";
      const submitValues = schema
        ? filterSchemaValuesForSubmit(schema, values, { step: stepForFilter })
        : values;

      // Fast-path: public auth skips validation/test and advances directly
      if (isMultiStep && isOnConnectorStep && isPublicAuth) {
        const connectorValues = this.filterValuesForStep(values, "connector");
        setConnectorConfig(connectorValues);
        setStep("source");
        return;
      }

      // --- Validation ---
      if (isStepFlowConnector && isOnSourceOrExplorerStep) {
        // Source/explorer step uses its own validation schema (not superforms)
        const sourceValidator = getValidationSchemaForConnector(
          connector.name as string,
          "source",
          stepState.step,
        );
        const result = await sourceValidator.validate(values);
        if (!result.success) {
          const fieldErrors: Record<string, string[]> = {};
          for (const issue of result.issues ?? []) {
            const key =
              issue.path?.[0] != null ? String(issue.path[0]) : "_errors";
            if (!fieldErrors[key]) fieldErrors[key] = [];
            fieldErrors[key].push(issue.message);
          }
          this.errorsStore.set(fieldErrors);
          event.cancel?.();
          return;
        }
        this.errorsStore.set({});
      } else if (!event.form.valid) {
        return;
      }

      // Show "Save Anyway" when a connector test fails
      if (
        typeof setShowSaveAnyway === "function" &&
        this.shouldShowSaveAnywayButton({
          isConnectorForm,
          event,
          stepState,
          selectedAuthMethod,
        })
      ) {
        setShowSaveAnyway(true);
      }

      // --- Submission ---
      try {
        if (isStepFlowConnector && isOnSourceOrExplorerStep) {
          // Step 2: submit the source/model and close
          await submitAddSourceForm(
            queryClient,
            connector,
            submitValues,
            stepState.connectorInstanceName ?? undefined,
          );
          onClose();
        } else if (isStepFlowConnector && isOnConnectorStep) {
          // Step 1: test connector, persist config, then advance to step 2
          await this.submitConnectorStepAndAdvance({
            queryClient,
            values,
            submitValues,
            isPublicAuth,
            isMultiStep,
          });
        } else if (this.formType === "source") {
          // Single-step source form
          await submitAddSourceForm(queryClient, connector, submitValues);
          onClose();
        } else {
          // Single-step connector form
          await submitAddConnectorForm(
            queryClient,
            connector,
            submitValues,
            false,
          );
          onClose();
        }
      } catch (e) {
        const { message, details } = this.normalizeError(e);
        setParamsError(message, details);
      }
    };
  }

  /**
   * Submit the connector step: test the connection (or skip for public auth),
   * persist connector config, then advance to the source/explorer step.
   */
  private async submitConnectorStepAndAdvance(args: {
    queryClient: QueryClient;
    values: FormData;
    submitValues: FormData;
    isPublicAuth: boolean;
    isMultiStep: boolean;
  }) {
    const { queryClient, values, submitValues, isPublicAuth, isMultiStep } =
      args;
    const nextStep = isMultiStep ? "source" : "explorer";

    if (isPublicAuth) {
      // Public auth skips the connection test
      const connectorValues = this.filterValuesForStep(values, "connector");
      setConnectorConfig(connectorValues);
      setConnectorInstanceName(null);
      setStep(nextStep);
      return;
    }

    // Test the connection, then persist config and advance
    const connectorInstanceName = await submitAddConnectorForm(
      queryClient,
      this.connector,
      submitValues,
      false,
    );
    const connectorValues = this.filterValuesForStep(submitValues, "connector");
    setConnectorConfig(connectorValues);
    setConnectorInstanceName(connectorInstanceName);
    setStep(nextStep);
  }

  onStringInputChange = (
    event: Event,
    taintedFields?: Record<string, boolean> | null,
  ) => {
    const target = event.target as HTMLInputElement;
    const { name, value } = target;
    const key = name || target.id;

    // Clear stale field-level errors as soon as the user edits the input.
    const clearFieldError = (store: ErrorsStore) => {
      if (!store?.update || !key) return;
      store.update(($errors) => {
        if (!$errors || !Object.prototype.hasOwnProperty.call($errors, key)) {
          return $errors;
        }
        const next = { ...$errors };
        delete next[key];
        return next;
      });
    };
    clearFieldError(this.errorsStore);
    if (name === "path" || name === "sql") {
      const nameTainted =
        taintedFields && typeof taintedFields === "object"
          ? Boolean(taintedFields?.name)
          : false;
      if (nameTainted) return;
      const inferred =
        name === "sql"
          ? inferModelNameFromSQL(value)
          : inferSourceName(this.connector, value);
      if (inferred) {
        const formStore = this.formStore;
        formStore.update(
          ($form) => {
            $form.name = inferred;
            return $form;
          },
          { taint: false },
        );
      }
    }
  };

  async handleFileUpload(file: File, fieldKey?: string): Promise<string> {
    const content = await file.text();

    if (fieldKey) {
      const schema = getConnectorSchema(this.schemaName);
      const field = schema?.properties?.[fieldKey];
      if (field?.["x-file-encoding"]) {
        const result = processFileContent(content, field);

        if (Object.keys(result.extractedValues).length > 0) {
          this.formStore.update(
            ($form) => {
              for (const [key, value] of Object.entries(
                result.extractedValues,
              )) {
                $form[key] = value;
              }
              return $form;
            },
            { taint: false },
          );
        }

        return result.encodedContent;
      }
    }

    return content;
  }

  /**
   * Compute YAML preview for the current form state.
   * HTTPS connector step uses synchronous client-side YAML generation;
   * all other paths call the GenerateTemplate RPC.
   */
  async computeYamlPreview(ctx: {
    stepState: ConnectorStepState | undefined;
    isMultiStepConnector: boolean;
    isConnectorForm: boolean;
    formValues: Record<string, unknown>;
  }): Promise<string> {
    const connector = this.connector;
    const { stepState, isMultiStepConnector, isConnectorForm, formValues } =
      ctx;

    const schema = getConnectorSchema(this.schemaName);
    const instanceId = get(runtime).instanceId;

    // HTTPS connector step: use synchronous client-side YAML generation
    // (backend doesn't support headers array yet)
    const getHttpsConnectorYamlPreview = (
      values: Record<string, unknown>,
    ): string => {
      const schemaConnectorFields = schema
        ? getSchemaFieldMetaList(schema, { step: "connector" })
        : [];
      const schemaConnectorSecretKeys = schema
        ? getSchemaSecretKeys(schema, { step: "connector" })
        : undefined;
      const schemaConnectorStringKeys = schema
        ? getSchemaStringKeys(schema, { step: "connector" })
        : undefined;

      const filteredValues = schema
        ? filterSchemaValuesForSubmit(schema, values, { step: "connector" })
        : values;
      return compileConnectorYAML(connector, filteredValues, {
        fieldFilter: (property) => {
          if ("internal" in property && property.internal) return false;
          return !("noPrompt" in property && property.noPrompt);
        },
        orderedProperties: schemaConnectorFields,
        secretKeys: schemaConnectorSecretKeys,
        stringKeys: schemaConnectorStringKeys,
        schema: schema ?? undefined,
      });
    };

    // For HTTPS connector step, use client-side preview
    if (connector.name === "https" && isConnectorForm) {
      return getHttpsConnectorYamlPreview(formValues);
    }
    if (
      connector.name === "https" &&
      isMultiStepConnector &&
      stepState?.step === "connector"
    ) {
      return getHttpsConnectorYamlPreview(formValues);
    }

    // All other paths: use the GenerateTemplate RPC
    const isOnConnectorStep = !stepState || stepState.step === "connector";
    const isOnSourceOrExplorerStep =
      stepState?.step === "source" || stepState?.step === "explorer";

    if (isMultiStepConnector && isOnConnectorStep) {
      // Step 1 of multi-step: preview the connector YAML
      const filteredValues = schema
        ? filterSchemaValuesForSubmit(schema, formValues, {
            step: "connector",
          })
        : formValues;
      const response = await generateTemplate(instanceId, {
        resourceType: "connector",
        driver: connector.name as string,
        properties: filteredValues,
      });
      return response.blob ?? "";
    }

    if (isMultiStepConnector && isOnSourceOrExplorerStep) {
      // Step 2 of multi-step: preview the model/source YAML
      const combinedValues = {
        ...(stepState?.connectorConfig || {}),
        ...formValues,
      } as Record<string, unknown>;

      // Filter out connector-step properties
      let sourceValues = combinedValues;
      if (schema) {
        const connectorPropertyKeys = new Set(
          getSchemaFieldMetaList(schema, { step: "connector" }).map(
            (field) => field.key,
          ),
        );
        sourceValues = Object.fromEntries(
          Object.entries(combinedValues).filter(
            ([key]) => !connectorPropertyKeys.has(key),
          ),
        );
      }

      const response = await generateTemplate(instanceId, {
        resourceType: "model",
        driver: connector.name as string,
        properties: sourceValues,
        connectorName:
          stepState?.connectorInstanceName || (connector.name as string),
      });
      return response.blob ?? "";
    }

    if (isConnectorForm) {
      // Single-step connector
      const filteredValues = schema
        ? filterSchemaValuesForSubmit(schema, formValues, {
            step: "connector",
          })
        : formValues;
      const response = await generateTemplate(instanceId, {
        resourceType: "connector",
        driver: connector.name as string,
        properties: filteredValues,
      });
      return response.blob ?? "";
    }

    // Single-step source form
    const response = await generateTemplate(instanceId, {
      resourceType: "model",
      driver: connector.name as string,
      properties: formValues,
    });
    return response.blob ?? "";
  }

  /**
   * Save connector anyway, returning a result object for the caller to handle.
   * Schema conditionals handle connector-specific requirements (e.g., SSL).
   */
  async saveConnectorAnyway(args: {
    queryClient: QueryClient;
    values: FormData;
  }): Promise<{ ok: true } | { ok: false; message: string; details?: string }> {
    const { queryClient, values } = args;
    const schema = getConnectorSchema(this.schemaName);
    const processedValues = schema
      ? filterSchemaValuesForSubmit(schema, values, { step: "connector" })
      : values;
    try {
      await submitAddConnectorForm(
        queryClient,
        this.connector,
        processedValues,
        true,
      );
      return { ok: true } as const;
    } catch (e) {
      const { message, details } = this.normalizeError(e);
      return { ok: false, message, details } as const;
    }
  }
}
