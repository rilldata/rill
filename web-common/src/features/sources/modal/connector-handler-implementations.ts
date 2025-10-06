import type { V1ConnectorDriver } from "@rilldata/web-common/runtime-client";
import type * as yup from "yup";
import type { ConnectorHandler, ConnectorProperty } from "./connector-handlers";
import type { AddDataFormType } from "./types";
import { getYupSchema, dsnSchema } from "./yupSchemas";
import { getInitialFormValuesFromProperties } from "../sourceUtils";
import { compileConnectorYAML } from "../../connectors/code-utils";
import { prepareSourceFormData, compileSourceYAML } from "../sourceUtils";
import {
  submitAddConnectorForm,
  submitAddSourceForm,
} from "./submitAddDataForm";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";

export abstract class BaseConnectorHandler implements ConnectorHandler {
  abstract getConnectorName(): string;

  supports(connector: V1ConnectorDriver): boolean {
    return connector.name === this.getConnectorName();
  }

  getProperties(
    connector: V1ConnectorDriver,
    formType: AddDataFormType,
  ): ConnectorProperty[] {
    if (formType === "source") {
      return connector.sourceProperties || [];
    } else {
      return (
        connector.configProperties?.filter(
          (property) => property.key !== "dsn",
        ) || []
      );
    }
  }

  getFilteredProperties(
    connector: V1ConnectorDriver,
    formType: AddDataFormType,
  ): ConnectorProperty[] {
    const properties = this.getProperties(connector, formType);

    // Apply connector-specific filtering
    return this.filterProperties(properties, connector);
  }

  protected filterProperties(
    properties: ConnectorProperty[],
    connector: V1ConnectorDriver,
  ): ConnectorProperty[] {
    // Default filtering logic
    if (connector.name === "duckdb") {
      return properties.filter(
        (property) => property.key !== "attach" && property.key !== "mode",
      );
    }

    // For other connectors, filter out noPrompt properties
    return properties.filter((property) => !property.noPrompt);
  }

  getValidationSchema(): yup.ObjectSchema<any> {
    const schema =
      getYupSchema[this.getConnectorName() as keyof typeof getYupSchema];
    if (!schema) {
      throw new Error(
        `No validation schema found for connector: ${this.getConnectorName()}`,
      );
    }
    return schema;
  }

  getInitialValues(
    connector: V1ConnectorDriver,
    formType: AddDataFormType,
  ): Record<string, unknown> {
    const properties = this.getProperties(connector, formType);
    return getInitialFormValuesFromProperties(properties as any);
  }

  getFormId(connector: V1ConnectorDriver, formType: "params" | "dsn"): string {
    const connectorName = connector.name;
    if (formType === "dsn") {
      return `add-data-${connectorName}-dsn-form`;
    } else {
      return `add-data-${connectorName}-form`;
    }
  }

  hasDsnFormOption(connector: V1ConnectorDriver): boolean {
    return Boolean(
      connector.configProperties?.some((property) => property.key === "dsn") &&
        connector.configProperties?.some((property) => property.key !== "dsn"),
    );
  }

  hasOnlyDsn(connector: V1ConnectorDriver): boolean {
    return (
      Boolean(
        connector.configProperties?.some((property) => property.key === "dsn"),
      ) &&
      !Boolean(
        connector.configProperties?.some((property) => property.key !== "dsn"),
      )
    );
  }

  getDsnProperties(connector: V1ConnectorDriver): ConnectorProperty[] {
    return (
      connector.configProperties?.filter(
        (property) => property.key === "dsn",
      ) || []
    );
  }

  async handleSubmit(
    connector: V1ConnectorDriver,
    formType: AddDataFormType,
    values: Record<string, unknown>,
  ): Promise<void> {
    // Apply any special logic before submission
    const processedValues =
      this.handleSpecialLogic?.(connector, formType, values) || values;

    if (formType === "source") {
      await submitAddSourceForm(queryClient, connector, processedValues);
    } else {
      await submitAddConnectorForm(queryClient, connector, processedValues);
    }
  }

  getYamlPreview(
    connector: V1ConnectorDriver,
    formType: AddDataFormType,
    values: Record<string, unknown>,
  ): string {
    if (formType === "source") {
      const [rewrittenConnector, rewrittenFormValues] = prepareSourceFormData(
        connector,
        values,
      );

      // Check if the connector was rewritten to DuckDB
      const isRewrittenToDuckDb = rewrittenConnector.name === "duckdb";

      if (isRewrittenToDuckDb) {
        return compileSourceYAML(rewrittenConnector, rewrittenFormValues);
      } else {
        return this.getConnectorYamlPreview(
          rewrittenConnector,
          rewrittenFormValues,
        );
      }
    } else {
      return this.getConnectorYamlPreview(connector, values);
    }
  }

  protected getConnectorYamlPreview(
    connector: V1ConnectorDriver,
    values: Record<string, unknown>,
  ): string {
    const properties = this.getFilteredProperties(connector, "connector");
    return compileConnectorYAML(connector, values, {
      fieldFilter: (property) => !property.noPrompt,
      orderedProperties: properties as any,
    });
  }

  handleSpecialLogic?(
    connector: V1ConnectorDriver,
    formType: AddDataFormType,
    values: Record<string, unknown>,
  ): Record<string, unknown>;
}

// Standard connector handler for most connectors
export class StandardConnectorHandler extends BaseConnectorHandler {
  constructor(private connectorName: string) {
    super();
  }

  getConnectorName(): string {
    return this.connectorName;
  }
}

// ClickHouse-specific handler
export class ClickHouseConnectorHandler extends BaseConnectorHandler {
  getConnectorName(): string {
    return "clickhouse";
  }

  protected filterProperties(
    properties: ConnectorProperty[],
    connector: V1ConnectorDriver,
  ): ConnectorProperty[] {
    return properties.filter(
      (property) => !property.noPrompt && property.key !== "managed",
    );
  }

  getFormId(connector: V1ConnectorDriver, formType: "params" | "dsn"): string {
    const connectorName = connector.name;
    if (formType === "dsn") {
      return `add-clickhouse-data-${connectorName}-dsn-form`;
    } else {
      return `add-clickhouse-data-${connectorName}-form`;
    }
  }

  handleSpecialLogic(
    connector: V1ConnectorDriver,
    formType: AddDataFormType,
    values: Record<string, unknown>,
  ): Record<string, unknown> {
    // ClickHouse Cloud specific requirements
    const processedValues = { ...values };

    // This would need to be passed from the component state
    // For now, we'll handle this in the component
    return processedValues;
  }

  getYamlPreview(
    connector: V1ConnectorDriver,
    formType: AddDataFormType,
    values: Record<string, unknown>,
  ): string {
    // ClickHouse-specific YAML preview logic
    // This would need connector type information from component state
    return this.getConnectorYamlPreview(connector, values);
  }
}
