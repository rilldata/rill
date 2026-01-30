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
  isExplorerConnector as isExplorerConnectorSchema,
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
import { compileSourceYAML, prepareSourceFormData } from "../sourceUtils";
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
  paramsFormId: string;

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
    this.paramsFormId = `add-data-${effectiveSchemaName}-form`;

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

  get isExplorerConnector(): boolean {
    const schema = getConnectorSchema(this.schemaName);
    return isExplorerConnectorSchema(schema);
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
    } else if (this.isExplorerConnector && stepState.step === "explorer") {
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
      this.isMultiStepConnector || this.isExplorerConnector;

    // Use schema-provided button labels when available (e.g., rill-managed ClickHouse)
    if (schemaButtonLabels && step === "connector") {
      return submitting ? schemaButtonLabels.loading : schemaButtonLabels.idle;
    }

    if (isConnectorForm) {
      if (isStepFlowConnector && step === "connector") {
        if (selectedAuthMethod === "public") {
          return submitting
            ? BUTTON_LABELS.public.submitting
            : BUTTON_LABELS.public.idle;
        }
        return submitting
          ? BUTTON_LABELS.connector.submitting
          : BUTTON_LABELS.connector.idle;
      }
      if (isStepFlowConnector && (step === "source" || step === "explorer")) {
        return submitting
          ? BUTTON_LABELS.source.submitting
          : BUTTON_LABELS.source.idle;
      }
      return submitting
        ? BUTTON_LABELS.connector.submitting
        : BUTTON_LABELS.connector.idle;
    }

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
    const isExplorer = isExplorerConnectorSchema(schema);
    const isStepFlowConnector = isMultiStep || isExplorer;
    const isConnectorForm = this.formType === "connector";

    return async (event: {
      form: SuperValidated<FormData, string, FormData>;
      result?: Extract<ActionResult, { type: "success" | "failure" }>;
      cancel?: () => void;
    }) => {
      const values = event.form.data;
      const stepState = get(connectorStepStore) as ConnectorStepState;
      const stepForFilter =
        isStepFlowConnector &&
        (stepState.step === "source" || stepState.step === "explorer")
          ? stepState.step
          : this.formType === "source"
            ? "source"
            : "connector";
      const filteredValues = schema
        ? filterSchemaValuesForSubmit(schema, values, {
            step: stepForFilter,
          })
        : values;
      const submitValues = filteredValues;
      const authKey = schema ? findRadioEnumKey(schema) : null;
      const selectedAuthMethod =
        (authKey && values && values[authKey] != null
          ? String(values[authKey])
          : undefined) ||
        getSelectedAuthMethod?.() ||
        "";
      // Fast-path: public auth skips validation/test and goes straight to source step.
      if (
        isMultiStep &&
        stepState.step === "connector" &&
        selectedAuthMethod === "public"
      ) {
        const connectorValues = this.filterValuesForStep(values, "connector");
        setConnectorConfig(connectorValues);
        setStep("source");
        return;
      }

      if (
        isStepFlowConnector &&
        (stepState.step === "source" || stepState.step === "explorer")
      ) {
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
          const errorsStore = this.errorsStore;
          errorsStore.set(fieldErrors);
          event.cancel?.();
          return;
        }
        const errorsStore = this.errorsStore;
        errorsStore.set({});
      } else if (!event.form.valid) {
        return;
      }

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

      try {
        if (
          isStepFlowConnector &&
          (stepState.step === "source" || stepState.step === "explorer")
        ) {
          const connectorInstanceName =
            stepState.connectorInstanceName ?? undefined;
          await submitAddSourceForm(
            queryClient,
            connector,
            submitValues,
            connectorInstanceName,
          );
          onClose();
        } else if (isStepFlowConnector && stepState.step === "connector") {
          // For public auth, skip Test & Connect and go straight to the next step.
          if (selectedAuthMethod === "public") {
            const connectorValues = this.filterValuesForStep(
              values,
              "connector",
            );
            setConnectorConfig(connectorValues);
            setConnectorInstanceName(null);
            if (isMultiStep) {
              setStep("source");
            } else if (isExplorer) {
              setStep("explorer");
            }
            return;
          }
          const connectorInstanceName = await submitAddConnectorForm(
            queryClient,
            connector,
            submitValues,
            false,
          );
          const connectorValues = this.filterValuesForStep(
            submitValues,
            "connector",
          );
          setConnectorConfig(connectorValues);
          setConnectorInstanceName(connectorInstanceName);
          if (isMultiStep) {
            setStep("source");
          } else if (isExplorer) {
            setStep("explorer");
          }
          return;
        } else if (this.formType === "source") {
          await submitAddSourceForm(queryClient, connector, submitValues);
          onClose();
        } else {
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

  async handleFileUpload(file: File): Promise<string> {
    const content = await file.text();
    try {
      const parsed = JSON.parse(content);
      const sanitized = JSON.stringify(parsed);
      if (this.connector.name === "bigquery" && parsed.project_id) {
        const formStore = this.formStore;
        formStore.update(
          ($form) => {
            $form.project_id = parsed.project_id;
            return $form;
          },
          { taint: false },
        );
      }
      return sanitized;
    } catch (error: unknown) {
      if (error instanceof SyntaxError) {
        throw new Error(`Invalid JSON file: ${error.message}`);
      }
      const message =
        error && typeof error === "object" && "message" in error
          ? String((error as { message: unknown }).message)
          : "Unknown error";
      throw new Error(`Failed to read file: ${message}`);
    }
  }

  /**
   * Compute YAML preview for the current form state.
   * Schema conditionals handle connector-specific requirements (e.g., managed flag, SSL).
   */
  /**
   * Compute YAML preview for the current form state.
   * Schema conditionals handle connector-specific requirements.
   */
  computeYamlPreview(ctx: {
    stepState: ConnectorStepState | undefined;
    isMultiStepConnector: boolean;
    isConnectorForm: boolean;
    paramsFormValues: Record<string, unknown>;
  }): string {
    const connector = this.connector;
    const {
      stepState,
      isMultiStepConnector,
      isConnectorForm,
      paramsFormValues,
    } = ctx;

    const schema = getConnectorSchema(this.schemaName);
    const schemaConnectorFields = schema
      ? getSchemaFieldMetaList(schema, { step: "connector" })
      : null;
    const schemaConnectorSecretKeys = schema
      ? getSchemaSecretKeys(schema, { step: "connector" })
      : undefined;
    const schemaConnectorStringKeys = schema
      ? getSchemaStringKeys(schema, { step: "connector" })
      : undefined;

    const connectorPropertiesForPreview = schemaConnectorFields ?? [];

    const getConnectorYamlPreview = (values: Record<string, unknown>) => {
      const filteredValues = schema
        ? filterSchemaValuesForSubmit(schema, values, { step: "connector" })
        : values;
      return compileConnectorYAML(connector, filteredValues, {
        fieldFilter: (property) => {
          if ("internal" in property && property.internal) return false;
          return !("noPrompt" in property && property.noPrompt);
        },
        orderedProperties: connectorPropertiesForPreview,
        secretKeys: schemaConnectorSecretKeys,
        stringKeys: schemaConnectorStringKeys,
      });
    };

    const getSourceYamlPreview = (values: Record<string, unknown>) => {
      // For multi-step connectors in step 2, filter out connector properties
      let filteredValues = values;
      if (
        (isMultiStepConnector && stepState?.step === "source") ||
        stepState?.step === "explorer"
      ) {
        const connectorPropertyKeys = new Set(
          schema
            ? getSchemaFieldMetaList(schema, { step: "connector" })
                .filter((field) => !field.internal)
                .map((field) => field.key)
            : [],
        );
        filteredValues = Object.fromEntries(
          Object.entries(values).filter(
            ([key]) => !connectorPropertyKeys.has(key),
          ),
        );
      }

      const [rewrittenConnector, rewrittenFormValues] = prepareSourceFormData(
        connector,
        filteredValues,
        {
          connectorInstanceName: stepState?.connectorInstanceName || undefined,
        },
      );
      const isExplorerStep = stepState?.step === "explorer";
      const isRewrittenToDuckDb = rewrittenConnector.name === "duckdb";
      const rewrittenSchema = getConnectorSchema(rewrittenConnector.name ?? "");
      const sourceStep = isExplorerStep ? "explorer" : "source";
      const rewrittenSecretKeys = rewrittenSchema
        ? getSchemaSecretKeys(rewrittenSchema, { step: sourceStep })
        : undefined;
      const rewrittenStringKeys = rewrittenSchema
        ? getSchemaStringKeys(rewrittenSchema, { step: sourceStep })
        : undefined;
      if (isRewrittenToDuckDb || isExplorerStep) {
        return compileSourceYAML(rewrittenConnector, rewrittenFormValues, {
          secretKeys: rewrittenSecretKeys,
          stringKeys: rewrittenStringKeys,
          originalDriverName: connector.name || undefined,
        });
      }
      return getConnectorYamlPreview(rewrittenFormValues);
    };

    // Multi-step connectors (S3, GCS, Azure)
    if (isMultiStepConnector) {
      if (stepState?.step === "connector") {
        return getConnectorYamlPreview(paramsFormValues);
      } else {
        const combinedValues = {
          ...(stepState?.connectorConfig || {}),
          ...paramsFormValues,
        } as Record<string, unknown>;
        return getSourceYamlPreview(combinedValues);
      }
    }

    if (isConnectorForm) return getConnectorYamlPreview(paramsFormValues);
    return getSourceYamlPreview(paramsFormValues);
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
