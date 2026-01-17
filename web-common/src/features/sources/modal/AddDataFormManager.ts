import { superForm, defaults } from "sveltekit-superforms";
import type { SuperValidated } from "sveltekit-superforms";
import * as yupLib from "yup";
import {
  yup as yupAdapter,
  type Infer as YupInfer,
  type InferIn as YupInferIn,
} from "sveltekit-superforms/adapters";
import type { V1ConnectorDriver } from "@rilldata/web-common/runtime-client";
import type { AddDataFormType } from "./types";
import { getValidationSchemaForConnector } from "./FormValidation";
import { inferSourceName } from "../sourceUtils";
import {
  submitAddConnectorForm,
  submitAddSourceForm,
} from "./submitAddDataForm";
import {
  normalizeConnectorError,
  applyClickHouseCloudRequirements,
  isEmpty,
} from "./utils";
import {
  FORM_HEIGHT_DEFAULT,
  FORM_HEIGHT_TALL,
  MULTI_STEP_CONNECTORS,
  TALL_FORM_CONNECTORS,
} from "./constants";
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
import type { ConnectorDriverProperty } from "@rilldata/web-common/runtime-client";
import type { ClickHouseConnectorType } from "./constants";
import type { ActionResult } from "@sveltejs/kit";
import { getConnectorSchema } from "./connector-schemas";
import {
  filterSchemaValuesForSubmit,
  findRadioEnumKey,
  getRequiredFieldsForValues,
  getSchemaFieldMetaList,
  getSchemaInitialValues,
  getSchemaSecretKeys,
  getSchemaStringKeys,
  isVisibleForValues,
  type SchemaFieldMeta,
} from "../../templates/schema-utils";

const dsnSchema = yupLib.object({
  dsn: yupLib.string().required("DSN is required"),
});

export type ClickhouseUiState = {
  properties: Array<ConnectorDriverProperty | SchemaFieldMeta>;
  filteredProperties: Array<ConnectorDriverProperty | SchemaFieldMeta>;
  dsnProperties: Array<ConnectorDriverProperty | SchemaFieldMeta>;
  isSubmitDisabled: boolean;
  formId: string;
  submitting: boolean;
  enforcedConnectionTab?: "parameters" | "dsn";
  shouldClearErrors?: boolean;
};

type SuperFormStore = {
  update: (
    updater: (value: Record<string, unknown>) => Record<string, unknown>,
    options?: any,
  ) => void;
};

type ClickhouseStateArgs = {
  connectorType: ClickHouseConnectorType;
  connectionTab: "parameters" | "dsn";
  paramsFormValues: Record<string, unknown>;
  dsnFormValues: Record<string, unknown>;
  paramsErrors: Record<string, unknown>;
  dsnErrors: Record<string, unknown>;
  paramsForm: SuperFormStore;
  dsnForm: SuperFormStore;
  paramsSubmitting: boolean;
  dsnSubmitting: boolean;
};

// Minimal onUpdate event type carrying Superforms's validated form
type SuperFormUpdateEvent = {
  form: SuperValidated<Record<string, unknown>, any, Record<string, unknown>>;
};

const BUTTON_LABELS = {
  public: { idle: "Continue", submitting: "Continuing..." },
  connector: { idle: "Test and Connect", submitting: "Testing connection..." },
  source: { idle: "Import Data", submitting: "Importing data..." },
};

export class AddDataFormManager {
  formHeight: string;
  paramsFormId: string;
  dsnFormId: string;
  hasDsnFormOption: boolean;
  hasOnlyDsn: boolean;
  properties: Array<ConnectorDriverProperty | SchemaFieldMeta>;
  filteredParamsProperties: Array<ConnectorDriverProperty | SchemaFieldMeta>;
  dsnProperties: Array<ConnectorDriverProperty | SchemaFieldMeta>;
  filteredDsnProperties: Array<ConnectorDriverProperty | SchemaFieldMeta>;

  // superforms instances
  params: ReturnType<typeof superForm>;
  dsn: ReturnType<typeof superForm>;
  private connector: V1ConnectorDriver;
  private formType: AddDataFormType;
  private clickhouseInitialValues: Record<string, unknown>;

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
    const schema = getConnectorSchema(this.connector.name ?? "");
    if (!schema?.properties) return values;
    return filterSchemaValuesForSubmit(schema, values, { step });
  }

  constructor(args: {
    connector: V1ConnectorDriver;
    formType: AddDataFormType;
    onParamsUpdate: (event: SuperFormUpdateEvent) => void;
    onDsnUpdate: (event: SuperFormUpdateEvent) => void;
    getSelectedAuthMethod?: () => string | undefined;
  }) {
    const {
      connector,
      formType,
      onParamsUpdate,
      onDsnUpdate,
      getSelectedAuthMethod,
    } = args;
    this.connector = connector;
    this.formType = formType;
    this.getSelectedAuthMethod = getSelectedAuthMethod;

    // Layout height
    this.formHeight = TALL_FORM_CONNECTORS.has(connector.name ?? "")
      ? FORM_HEIGHT_TALL
      : FORM_HEIGHT_DEFAULT;

    // IDs
    this.paramsFormId = `add-data-${connector.name}-form`;
    this.dsnFormId = `add-data-${connector.name}-dsn-form`;

    const isSourceForm = formType === "source";
    const schema = getConnectorSchema(connector.name ?? "");
    const schemaStep = isSourceForm ? "source" : "connector";
    const schemaFields = schema
      ? getSchemaFieldMetaList(schema, { step: schemaStep })
      : [];

    // Base properties
    this.properties = schemaFields;

    // Filter properties based on connector type
    this.filteredParamsProperties = this.properties;

    // DSN properties
    this.dsnProperties = schemaFields.filter((field) => field.key === "dsn");
    this.filteredDsnProperties = this.dsnProperties;

    // DSN flags
    this.hasDsnFormOption = false;
    this.hasOnlyDsn = false;

    // Superforms: params
    const paramsAdapter = getValidationSchemaForConnector(
      connector.name as string,
      formType,
    );
    type ParamsOut = Record<string, unknown>;
    type ParamsIn = Record<string, unknown>;
    const initialFormValues = schema
      ? getSchemaInitialValues(schema, { step: schemaStep })
      : {};
    const paramsDefaults = defaults<ParamsOut, any, ParamsIn>(
      initialFormValues as Partial<ParamsOut>,
      paramsAdapter,
    );
    this.params = superForm<ParamsOut, any, ParamsIn>(paramsDefaults, {
      SPA: true,
      validators: paramsAdapter,
      onUpdate: onParamsUpdate,
      resetForm: false,
      validationMethod: "onsubmit",
    });

    // Superforms: dsn
    const dsnAdapter = yupAdapter(dsnSchema);
    type DsnOut = YupInfer<typeof dsnSchema, "yup">;
    type DsnIn = YupInferIn<typeof dsnSchema, "yup">;
    this.dsn = superForm<DsnOut, any, DsnIn>(defaults(dsnAdapter), {
      SPA: true,
      validators: dsnAdapter,
      onUpdate: onDsnUpdate,
      resetForm: false,
      validationMethod: "onsubmit",
    });

    // ClickHouse-specific defaults
    this.clickhouseInitialValues =
      connector.name === "clickhouse" && schema
        ? getSchemaInitialValues(schema, { step: "connector" })
        : {};
  }

  get isSourceForm(): boolean {
    return this.formType === "source";
  }

  get isConnectorForm(): boolean {
    return this.formType === "connector";
  }

  get isMultiStepConnector(): boolean {
    return MULTI_STEP_CONNECTORS.includes(this.connector.name ?? "");
  }

  get isExplorerConnector(): boolean {
    return Boolean(this.connector.implementsWarehouse);
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

  getActiveFormId(args: {
    connectionTab: "parameters" | "dsn";
    onlyDsn: boolean;
  }): string {
    const { connectionTab, onlyDsn } = args;
    return onlyDsn || connectionTab === "dsn"
      ? this.dsnFormId
      : this.paramsFormId;
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
    clickhouseConnectorType?: ClickHouseConnectorType;
    clickhouseSubmitting?: boolean;
    selectedAuthMethod?: string;
  }): string {
    const {
      isConnectorForm,
      step,
      submitting,
      clickhouseConnectorType,
      clickhouseSubmitting,
      selectedAuthMethod,
    } = args;
    const isClickhouse = this.connector.name === "clickhouse";
    const isStepFlowConnector = this.isMultiStepConnector || this.isExplorerConnector;

    if (isClickhouse) {
      if (clickhouseConnectorType === "rill-managed") {
        return clickhouseSubmitting ? "Connecting..." : "Connect";
      }
      return clickhouseSubmitting
        ? "Testing connection..."
        : "Test and Connect";
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

  computeClickhouseState(args: ClickhouseStateArgs): ClickhouseUiState | null {
    if (this.connector.name !== "clickhouse") return null;
    const {
      connectorType,
      connectionTab,
      paramsFormValues,
      dsnFormValues,
      paramsErrors,
      dsnErrors,
      paramsForm,
      paramsSubmitting,
      dsnSubmitting,
    } = args;
    const schema = getConnectorSchema(this.connector.name ?? "");

    // Keep connector_type in sync on the params form
    if (paramsFormValues?.connector_type !== connectorType) {
      paramsForm.update(
        ($form: any) => ({
          ...$form,
          connector_type: connectorType,
        }),
        { taint: false } as any,
      );
    }

    const enforcedConnectionTab =
      connectorType === "rill-managed" ? ("parameters" as const) : undefined;
    const activeConnectionTab = enforcedConnectionTab ?? connectionTab;

    const requiredFields = schema
      ? getRequiredFieldsForValues(schema, paramsFormValues ?? {}, "connector")
      : new Set<string>();
    const isSubmitDisabled = Array.from(requiredFields).some((key) => {
      if (schema && !isVisibleForValues(schema, key, paramsFormValues ?? {})) {
        return false;
      }
      const errorsForField =
        activeConnectionTab === "dsn"
          ? (dsnErrors?.[key] as any)
          : (paramsErrors?.[key] as any);
      const value =
        activeConnectionTab === "dsn"
          ? dsnFormValues?.[key]
          : paramsFormValues?.[key];
      return isEmpty(value) || Boolean(errorsForField?.length);
    });

    const submitting =
      activeConnectionTab === "dsn" ? dsnSubmitting : paramsSubmitting;
    const formId =
      activeConnectionTab === "dsn" ? this.dsnFormId : this.paramsFormId;

    return {
      properties: this.properties,
      filteredProperties: this.filteredParamsProperties,
      dsnProperties: this.dsnProperties,
      isSubmitDisabled,
      formId,
      submitting,
      enforcedConnectionTab,
      shouldClearErrors: connectorType === "rill-managed",
    };
  }

  getClickhouseDefaults(
    connectorType: ClickHouseConnectorType,
  ): Record<string, unknown> | null {
    if (this.connector.name !== "clickhouse") return null;
    const baseDefaults = { ...this.clickhouseInitialValues };
    delete (baseDefaults as Record<string, unknown>).connector_type;

    if (connectorType === "clickhouse-cloud") {
      return {
        ...baseDefaults,
        managed: false,
        port: "8443",
        ssl: true,
        connector_type: "clickhouse-cloud",
        connection_mode: "parameters",
      };
    }

    if (connectorType === "rill-managed") {
      return {
        ...baseDefaults,
        managed: true,
        connector_type: "rill-managed",
        connection_mode: "parameters",
      };
    }

    return {
      ...baseDefaults,
      managed: false,
      connector_type: "self-hosted",
      connection_mode: "parameters",
    };
  }

  makeOnUpdate(args: {
    onClose: () => void;
    queryClient: any;
    getConnectionTab: () => "parameters" | "dsn";
    getSelectedAuthMethod?: () => string | undefined;
    setParamsError: (message: string | null, details?: string) => void;
    setDsnError: (message: string | null, details?: string) => void;
    setShowSaveAnyway?: (value: boolean) => void;
  }) {
    const {
      onClose,
      queryClient,
      getConnectionTab,
      getSelectedAuthMethod,
      setParamsError,
      setDsnError,
      setShowSaveAnyway,
    } = args;
    const connector = this.connector;
    const isMultiStepConnector = MULTI_STEP_CONNECTORS.includes(
      connector.name ?? "",
    );
    const isExplorerConnector = this.isExplorerConnector;
    const isStepFlowConnector = isMultiStepConnector || isExplorerConnector;
    const isConnectorForm = this.formType === "connector";

    return async (event: {
      form: SuperValidated<
        Record<string, unknown>,
        any,
        Record<string, unknown>
      >;
      result?: Extract<ActionResult, { type: "success" | "failure" }>;
      cancel?: () => void;
    }) => {
      const values = event.form.data;
      const connectionTab = getConnectionTab();
      const schema = getConnectorSchema(this.connector.name ?? "");
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
      const submitValues =
        this.connector.name === "clickhouse"
          ? this.filterClickhouseValues(filteredValues, connectionTab)
          : filteredValues;
      const authKey = schema ? findRadioEnumKey(schema) : null;
      const selectedAuthMethod =
        (authKey && values && values[authKey] != null
          ? String(values[authKey])
          : undefined) ||
        getSelectedAuthMethod?.() ||
        "";
      // Fast-path: public auth skips validation/test and goes straight to source step.
      if (
        isMultiStepConnector &&
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
          (this.params.errors as any).set(fieldErrors);
          event.cancel?.();
          return;
        }
        (this.params.errors as any).set({});
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
            if (isMultiStepConnector) {
              setStep("source");
            } else if (isExplorerConnector) {
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
          if (isMultiStepConnector) {
            setStep("source");
          } else if (isExplorerConnector) {
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
        if (isConnectorForm && (this.hasOnlyDsn || connectionTab === "dsn")) {
          setDsnError(message, details);
        } else {
          setParamsError(message, details);
        }
      } finally {
        // no-op: saveAnyway handled in Svelte
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
    const clearFieldError = (store: any) => {
      if (!store?.update || !key) return;
      store.update(($errors: Record<string, unknown>) => {
        if (!$errors || !Object.prototype.hasOwnProperty.call($errors, key)) {
          return $errors;
        }
        const next = { ...$errors };
        delete next[key];
        return next;
      });
    };
    clearFieldError(this.params.errors);
    clearFieldError(this.dsn.errors);
    if (name === "path") {
      const nameTainted =
        taintedFields && typeof taintedFields === "object"
          ? Boolean(taintedFields?.name)
          : false;
      if (nameTainted) return;
      const inferred = inferSourceName(this.connector, value);
      if (inferred)
        (this.params.form as any).update(
          ($form: any) => {
            $form.name = inferred;
            return $form;
          },
          { taint: false } as any,
        );
    }
  };

  async handleFileUpload(file: File): Promise<string> {
    const content = await file.text();
    try {
      const parsed = JSON.parse(content);
      const sanitized = JSON.stringify(parsed);
      if (this.connector.name === "bigquery" && parsed.project_id) {
        (this.params.form as any).update(
          ($form: any) => {
            $form.project_id = parsed.project_id;
            return $form;
          },
          { taint: false } as any,
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
   */
  computeYamlPreview(ctx: {
    connectionTab: "parameters" | "dsn";
    onlyDsn: boolean;
    filteredParamsProperties: Array<ConnectorDriverProperty | SchemaFieldMeta>;
    filteredDsnProperties: Array<ConnectorDriverProperty | SchemaFieldMeta>;
    stepState: ConnectorStepState | undefined;
    isMultiStepConnector: boolean;
    isConnectorForm: boolean;
    paramsFormValues: Record<string, unknown>;
    dsnFormValues: Record<string, unknown>;
    clickhouseConnectorType?: ClickHouseConnectorType;
    clickhouseParamsValues?: Record<string, unknown>;
    clickhouseDsnValues?: Record<string, unknown>;
  }): string {
    const connector = this.connector;
    const {
      connectionTab,
      onlyDsn,
      filteredParamsProperties,
      filteredDsnProperties,
      stepState,
      isMultiStepConnector,
      isConnectorForm,
      paramsFormValues,
      dsnFormValues,
      clickhouseConnectorType,
      clickhouseParamsValues,
      clickhouseDsnValues,
    } = ctx;

    const schema = getConnectorSchema(connector.name ?? "");
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
      const orderedProperties =
        onlyDsn || connectionTab === "dsn"
          ? filteredDsnProperties
          : connectorPropertiesForPreview;
      return compileConnectorYAML(connector, filteredValues, {
        fieldFilter: (property) => {
          if (onlyDsn || connectionTab === "dsn") return true;
          if ("internal" in property && property.internal) return false;
          return !("noPrompt" in property && property.noPrompt);
        },
        orderedProperties,
        secretKeys: schemaConnectorSecretKeys,
        stringKeys: schemaConnectorStringKeys,
      });
    };

    const getClickHouseYamlPreview = (
      values: Record<string, unknown>,
      chType: ClickHouseConnectorType | undefined,
    ) => {
      const filteredValues = schema
        ? filterSchemaValuesForSubmit(schema, values, { step: "connector" })
        : values;
      // Convert to managed boolean and apply CH Cloud requirements for preview
      const managed = chType === "rill-managed";
      const previewValues = {
        ...filteredValues,
        managed,
      } as Record<string, unknown>;
      const finalValues = applyClickHouseCloudRequirements(
        connector.name,
        chType as ClickHouseConnectorType,
        previewValues,
      );
      return compileConnectorYAML(connector, finalValues, {
        fieldFilter: (property) => {
          if (onlyDsn || connectionTab === "dsn") return true;
          if ("internal" in property && property.internal) return false;
          return !("noPrompt" in property && property.noPrompt);
        },
        orderedProperties:
          connectionTab === "dsn"
            ? filteredDsnProperties
            : filteredParamsProperties,
        secretKeys: schemaConnectorSecretKeys,
        stringKeys: schemaConnectorStringKeys,
      });
    };

    const getSourceYamlPreview = (values: Record<string, unknown>) => {
      // For multi-step connectors in step 2, filter out connector properties
      let filteredValues = values;
      if (isMultiStepConnector && stepState?.step === "source") {
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
      const isRewrittenToDuckDb = rewrittenConnector.name === "duckdb";
      const rewrittenSchema = getConnectorSchema(rewrittenConnector.name ?? "");
      const rewrittenSecretKeys = rewrittenSchema
        ? getSchemaSecretKeys(rewrittenSchema, { step: "source" })
        : undefined;
      const rewrittenStringKeys = rewrittenSchema
        ? getSchemaStringKeys(rewrittenSchema, { step: "source" })
        : undefined;
      if (isRewrittenToDuckDb) {
        return compileSourceYAML(rewrittenConnector, rewrittenFormValues, {
          secretKeys: rewrittenSecretKeys,
          stringKeys: rewrittenStringKeys,
        });
      }
      return getConnectorYamlPreview(rewrittenFormValues);
    };

    // ClickHouse special-case
    if (connector.name === "clickhouse") {
      const values =
        connectionTab === "dsn"
          ? clickhouseDsnValues || {}
          : clickhouseParamsValues || {};
      return getClickHouseYamlPreview(values, clickhouseConnectorType);
    }

    // Multi-step connectors
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

    const currentValues =
      onlyDsn || connectionTab === "dsn" ? dsnFormValues : paramsFormValues;
    if (isConnectorForm) return getConnectorYamlPreview(currentValues);
    return getSourceYamlPreview(currentValues);
  }

  /**
   * Save connector anyway (non-ClickHouse), returning a result object for the caller to handle.
   */
  async saveConnectorAnyway(args: {
    queryClient: any;
    values: Record<string, unknown>;
    clickhouseConnectorType?: ClickHouseConnectorType;
    connectionTab?: "parameters" | "dsn";
  }): Promise<{ ok: true } | { ok: false; message: string; details?: string }> {
    const { queryClient, values, clickhouseConnectorType, connectionTab } =
      args;
    const tab = connectionTab ?? "parameters";
    const schema = getConnectorSchema(this.connector.name ?? "");
    const prunedValues = schema
      ? filterSchemaValuesForSubmit(schema, values, { step: "connector" })
      : values;
    const filteredValues = this.filterClickhouseValues(prunedValues, tab);
    const processedValues = applyClickHouseCloudRequirements(
      this.connector.name,
      (clickhouseConnectorType as ClickHouseConnectorType) ||
        ("self-hosted" as ClickHouseConnectorType),
      filteredValues,
    );
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

  private filterClickhouseValues(
    values: Record<string, unknown>,
    connectionTab: "parameters" | "dsn",
  ): Record<string, unknown> {
    if (this.connector.name !== "clickhouse") return values;
    if (connectionTab !== "dsn") {
      if (!("dsn" in values)) return values;
      // eslint-disable-next-line @typescript-eslint/no-unused-vars
      const { dsn: _unused, ...rest } = values;
      return rest;
    }

    const allowed = new Set(["dsn", "managed"]);
    return Object.fromEntries(
      Object.entries(values).filter(([key]) => allowed.has(key)),
    );
  }
}
