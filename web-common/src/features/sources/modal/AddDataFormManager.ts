import { superForm, defaults } from "sveltekit-superforms";
import { yup } from "sveltekit-superforms/adapters";
import type { V1ConnectorDriver } from "@rilldata/web-common/runtime-client";
import type { AddDataFormType } from "./types";
import { getValidationSchemaForConnector, dsnSchema } from "./FormValidation";
import { getInitialFormValuesFromProperties } from "../sourceUtils";

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

  constructor(args: {
    connector: V1ConnectorDriver;
    formType: AddDataFormType;
    onParamsUpdate: any;
    onDsnUpdate: any;
  }) {
    const { connector, formType, onParamsUpdate, onDsnUpdate } = args;

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
}
