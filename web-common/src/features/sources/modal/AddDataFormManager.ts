import { superForm, defaults } from "sveltekit-superforms";
import { yup } from "sveltekit-superforms/adapters";
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
import { MULTI_STEP_CONNECTORS } from "./constants";
import {
  connectorStepStore,
  setConnectorConfig,
  setStep,
} from "./connectorStepStore";
import { get } from "svelte/store";

export class AddDataFormManager {
  formHeight: string;
  paramsFormId: string;
  dsnFormId: string;
  hasDsnFormOption: boolean;
  hasOnlyDsn: boolean;
  properties: any[];
  filteredParamsProperties: any[];
  dsnProperties: any[];
  filteredDsnProperties: any[];

  // superforms instances
  params: ReturnType<typeof superForm>;
  dsn: ReturnType<typeof superForm>;
  private connector: V1ConnectorDriver;
  private formType: AddDataFormType;

  constructor(args: {
    connector: V1ConnectorDriver;
    formType: AddDataFormType;
    onParamsUpdate: any;
    onDsnUpdate: any;
  }) {
    const { connector, formType, onParamsUpdate, onDsnUpdate } = args;
    this.connector = connector;
    this.formType = formType;

    // Layout height
    this.formHeight = ["clickhouse", "snowflake", "salesforce"].includes(
      connector.name ?? "",
    )
      ? "max-h-[38.5rem] min-h-[38.5rem]"
      : "max-h-[34.5rem] min-h-[34.5rem]";

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
    const schema = yup(
      getValidationSchemaForConnector(connector.name as string),
    );
    const initialFormValues = getInitialFormValuesFromProperties(
      this.properties,
    );
    this.params = superForm(initialFormValues, {
      SPA: true,
      validators: schema,
      onUpdate: onParamsUpdate,
      resetForm: false,
    } as any);

    // Superforms: dsn
    const dsnYupSchema = yup(dsnSchema);
    this.dsn = superForm(defaults(dsnYupSchema), {
      SPA: true,
      validators: dsnYupSchema,
      onUpdate: onDsnUpdate,
      resetForm: false,
    } as any);
  }

  destroy() {}

  // Business logic methods (minimal extraction)

  makeOnUpdate(args: {
    onClose: () => void;
    queryClient: any;
    getConnectionTab: () => "parameters" | "dsn";
    setParamsError: (message: string | null, details?: string) => void;
    setDsnError: (message: string | null, details?: string) => void;
  }) {
    const {
      onClose,
      queryClient,
      getConnectionTab,
      setParamsError,
      setDsnError,
    } = args;
    const connector = this.connector;
    const isMultiStepConnector = MULTI_STEP_CONNECTORS.includes(
      connector.name ?? "",
    );
    const isConnectorForm = this.formType === "connector";

    return async (event: any) => {
      if (!event.form.valid) return;

      const values = event.form.data as Record<string, unknown>;

      try {
        const stepState = get(connectorStepStore) as any;
        if (isMultiStepConnector && stepState.step === "source") {
          await submitAddSourceForm(queryClient, connector, values);
          onClose();
        } else if (isMultiStepConnector && stepState.step === "connector") {
          await submitAddConnectorForm(queryClient, connector, values, true);
          setConnectorConfig(values);
          setStep("source");
          return;
        } else if (this.formType === "source") {
          await submitAddSourceForm(queryClient, connector, values);
          onClose();
        } else {
          await submitAddConnectorForm(queryClient, connector, values, true);
          onClose();
        }
      } catch (e) {
        const { message, details } = normalizeConnectorError(
          connector.name ?? "",
          e,
        );
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

  onStringInputChange = (event: Event) => {
    const target = event.target as HTMLInputElement;
    const { name, value } = target;
    if (name === "path") {
      const tainted: any = get(this.params.tainted) as any;
      if (tainted?.name) return;
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
    } catch (error: any) {
      if (error instanceof SyntaxError) {
        throw new Error(`Invalid JSON file: ${error.message}`);
      }
      throw new Error(`Failed to read file: ${error.message}`);
    }
  }
}
