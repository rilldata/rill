import { superForm, defaults } from "sveltekit-superforms";
import type { SuperValidated } from "sveltekit-superforms";
import {
  yup,
  type Infer as YupInfer,
  type InferIn as YupInferIn,
} from "sveltekit-superforms/adapters";
import type { V1ConnectorDriver } from "@rilldata/web-common/runtime-client";
import type { AddDataFormType } from "./types";
import { getValidationSchemaForConnector, dsnSchema } from "./FormValidation";
import {
  getInitialFormValuesFromProperties,
  inferSourceName,
} from "../sourceUtils";
import {
  submitAddConnectorForm,
  submitAddSourceForm,
} from "./submitAddDataForm";
import { normalizeConnectorError } from "./utils";
import {
  FORM_HEIGHT_DEFAULT,
  FORM_HEIGHT_TALL,
  MULTI_STEP_CONNECTORS,
  TALL_FORM_CONNECTORS,
} from "./constants";
import {
  connectorStepStore,
  setConnectorConfig,
  setStep,
} from "./connectorStepStore";
import { get } from "svelte/store";
import { compileConnectorYAML } from "../../connectors/code-utils";
import { compileSourceYAML, prepareSourceFormData } from "../sourceUtils";
import type { ConnectorDriverProperty } from "@rilldata/web-common/runtime-client";
import type { ClickHouseConnectorType } from "./constants";
import { applyClickHouseCloudRequirements } from "./utils";
import type { ActionResult } from "@sveltejs/kit";

// Minimal onUpdate event type carrying Superforms's validated form
type SuperFormUpdateEvent = {
  form: SuperValidated<Record<string, unknown>, any, Record<string, unknown>>;
};

// Shape of the step store for multi-step connectors
type ConnectorStepState = {
  step: "connector" | "source";
  connectorConfig: Record<string, unknown> | null;
};

export class AddDataFormManager {
  formHeight: string;
  paramsFormId: string;
  dsnFormId: string;
  hasDsnFormOption: boolean;
  hasOnlyDsn: boolean;
  properties: ConnectorDriverProperty[];
  filteredParamsProperties: ConnectorDriverProperty[];
  dsnProperties: ConnectorDriverProperty[];
  filteredDsnProperties: ConnectorDriverProperty[];

  // superforms instances
  params: ReturnType<typeof superForm>;
  dsn: ReturnType<typeof superForm>;
  private connector: V1ConnectorDriver;
  private formType: AddDataFormType;

  // Centralized error normalization for this manager
  private normalizeError(e: unknown): { message: string; details?: string } {
    return normalizeConnectorError(this.connector.name ?? "", e);
  }

  constructor(args: {
    connector: V1ConnectorDriver;
    formType: AddDataFormType;
    onParamsUpdate: (event: SuperFormUpdateEvent) => void;
    onDsnUpdate: (event: SuperFormUpdateEvent) => void;
  }) {
    const { connector, formType, onParamsUpdate, onDsnUpdate } = args;
    this.connector = connector;
    this.formType = formType;

    // Layout height
    this.formHeight = TALL_FORM_CONNECTORS.has(connector.name ?? "")
      ? FORM_HEIGHT_TALL
      : FORM_HEIGHT_DEFAULT;

    // IDs
    this.paramsFormId = `add-data-${connector.name}-form`;
    this.dsnFormId = `add-data-${connector.name}-dsn-form`;

    const isSourceForm = formType === "source";
    const isConnectorForm = formType === "connector";

    // Base properties
    this.properties =
      (isSourceForm
        ? connector.sourceProperties
        : connector.configProperties?.filter((p) => p.key !== "dsn")) ?? [];

    // Filter properties based on connector type
    this.filteredParamsProperties = (() => {
      if (connector.name === "duckdb") {
        return this.properties.filter(
          (p) => p.key !== "attach" && p.key !== "mode",
        );
      }
      return this.properties.filter((p) => !p.noPrompt);
    })();

    // DSN properties
    this.dsnProperties =
      connector.configProperties?.filter((p) => p.key === "dsn") ?? [];
    this.filteredDsnProperties = this.dsnProperties;

    // DSN flags
    this.hasDsnFormOption = !!(
      isConnectorForm &&
      connector.configProperties?.some((p) => p.key === "dsn") &&
      connector.configProperties?.some((p) => p.key !== "dsn")
    );
    this.hasOnlyDsn = !!(
      isConnectorForm &&
      connector.configProperties?.some((p) => p.key === "dsn") &&
      !connector.configProperties?.some((p) => p.key !== "dsn")
    );

    // Superforms: params
    const paramsSchemaDef = getValidationSchemaForConnector(
      connector.name as string,
    );
    const paramsAdapter = yup(paramsSchemaDef);
    type ParamsOut = YupInfer<typeof paramsSchemaDef, "yup">;
    type ParamsIn = YupInferIn<typeof paramsSchemaDef, "yup">;
    const initialFormValues = getInitialFormValuesFromProperties(
      this.properties,
    );
    const paramsDefaults = defaults<ParamsOut, any, ParamsIn>(
      initialFormValues as Partial<ParamsOut>,
      paramsAdapter,
    );
    this.params = superForm<ParamsOut, any, ParamsIn>(paramsDefaults, {
      SPA: true,
      validators: paramsAdapter,
      onUpdate: onParamsUpdate,
      resetForm: false,
    });

    // Superforms: dsn
    const dsnAdapter = yup(dsnSchema);
    type DsnOut = YupInfer<typeof dsnSchema, "yup">;
    type DsnIn = YupInferIn<typeof dsnSchema, "yup">;
    this.dsn = superForm<DsnOut, any, DsnIn>(defaults(dsnAdapter), {
      SPA: true,
      validators: dsnAdapter,
      onUpdate: onDsnUpdate,
      resetForm: false,
    });
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
    setConnectorConfig(get(this.params.form) as Record<string, unknown>);
    setStep("source");
  }

  handleBack(onBack: () => void): void {
    const stepState = get(connectorStepStore) as ConnectorStepState;
    if (this.isMultiStepConnector && stepState.step === "source") {
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
  }): string {
    const {
      isConnectorForm,
      step,
      submitting,
      clickhouseConnectorType,
      clickhouseSubmitting,
    } = args;
    const isClickhouse = this.connector.name === "clickhouse";

    if (isClickhouse) {
      if (clickhouseConnectorType === "rill-managed") {
        return clickhouseSubmitting ? "Connecting..." : "Connect";
      }
      return clickhouseSubmitting
        ? "Testing connection..."
        : "Test and Connect";
    }

    if (isConnectorForm) {
      if (this.isMultiStepConnector && step === "connector") {
        return submitting ? "Testing connection..." : "Test and Connect";
      }
      if (this.isMultiStepConnector && step === "source") {
        return submitting ? "Creating model..." : "Test and Add data";
      }
      return submitting ? "Testing connection..." : "Test and Connect";
    }

    return "Test and Add data";
  }

  makeOnUpdate(args: {
    onClose: () => void;
    queryClient: any;
    getConnectionTab: () => "parameters" | "dsn";
    setParamsError: (message: string | null, details?: string) => void;
    setDsnError: (message: string | null, details?: string) => void;
    setShowSaveAnyway?: (value: boolean) => void;
  }) {
    const {
      onClose,
      queryClient,
      getConnectionTab,
      setParamsError,
      setDsnError,
      setShowSaveAnyway,
    } = args;
    const connector = this.connector;
    const isMultiStepConnector = MULTI_STEP_CONNECTORS.includes(
      connector.name ?? "",
    );
    const isConnectorForm = this.formType === "connector";

    return async (event: {
      form: SuperValidated<
        Record<string, unknown>,
        any,
        Record<string, unknown>
      >;
      result?: Extract<ActionResult, { type: "success" | "failure" }>;
    }) => {
      // For non-ClickHouse connectors, expose Save Anyway when a submission starts
      if (
        isConnectorForm &&
        connector.name !== "clickhouse" &&
        typeof setShowSaveAnyway === "function" &&
        event?.result
      ) {
        setShowSaveAnyway(true);
      }

      if (!event.form.valid) return;

      const values = event.form.data;

      try {
        const stepState = get(connectorStepStore) as ConnectorStepState;
        if (isMultiStepConnector && stepState.step === "source") {
          await submitAddSourceForm(queryClient, connector, values);
          onClose();
        } else if (isMultiStepConnector && stepState.step === "connector") {
          await submitAddConnectorForm(queryClient, connector, values, false);
          setConnectorConfig(values);
          setStep("source");
          return;
        } else if (this.formType === "source") {
          await submitAddSourceForm(queryClient, connector, values);
          onClose();
        } else {
          await submitAddConnectorForm(queryClient, connector, values, false);
          onClose();
        }
      } catch (e) {
        const { message, details } = this.normalizeError(e);
        const connectionTab = getConnectionTab();
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
    filteredParamsProperties: ConnectorDriverProperty[];
    filteredDsnProperties: ConnectorDriverProperty[];
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

    const getConnectorYamlPreview = (values: Record<string, unknown>) => {
      return compileConnectorYAML(connector, values, {
        fieldFilter: (property) => {
          if (onlyDsn || connectionTab === "dsn") return true;
          return !property.noPrompt;
        },
        orderedProperties:
          onlyDsn || connectionTab === "dsn"
            ? filteredDsnProperties
            : filteredParamsProperties,
      });
    };

    const getClickHouseYamlPreview = (
      values: Record<string, unknown>,
      chType: ClickHouseConnectorType | undefined,
    ) => {
      // Convert to managed boolean and apply CH Cloud requirements for preview
      const managed = chType === "rill-managed";
      const previewValues = { ...values, managed } as Record<string, unknown>;
      const finalValues = applyClickHouseCloudRequirements(
        connector.name,
        chType as ClickHouseConnectorType,
        previewValues,
      );
      return compileConnectorYAML(connector, finalValues, {
        fieldFilter: (property) => {
          if (onlyDsn || connectionTab === "dsn") return true;
          return !property.noPrompt;
        },
        orderedProperties:
          connectionTab === "dsn"
            ? filteredDsnProperties
            : filteredParamsProperties,
      });
    };

    const getSourceYamlPreview = (values: Record<string, unknown>) => {
      // For multi-step connectors in step 2, filter out connector properties
      let filteredValues = values;
      if (isMultiStepConnector && stepState?.step === "source") {
        const connectorPropertyKeys = new Set(
          connector.configProperties?.map((p) => p.key).filter(Boolean) || [],
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
      );
      const isRewrittenToDuckDb = rewrittenConnector.name === "duckdb";
      if (isRewrittenToDuckDb) {
        return compileSourceYAML(rewrittenConnector, rewrittenFormValues);
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
  }): Promise<{ ok: true } | { ok: false; message: string; details?: string }> {
    const { queryClient, values, clickhouseConnectorType } = args;
    const processedValues = applyClickHouseCloudRequirements(
      this.connector.name,
      (clickhouseConnectorType as ClickHouseConnectorType) ||
        ("self-hosted" as ClickHouseConnectorType),
      values,
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
}
