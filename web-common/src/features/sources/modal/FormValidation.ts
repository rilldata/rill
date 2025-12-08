import { dsnSchema, getYupSchema } from "./yupSchemas";
import type { ConnectorStep } from "./connectorStepStore";

export { dsnSchema };

export function getValidationSchemaForConnector(
  name: string,
  step?: ConnectorStep,
) {
  // For multi-step S3 connector, use step-specific schema
  if (name === "s3") {
    if (step === "source") {
      return getYupSchema["s3_source"];
    }
    // Default to connector step for multi-step connectors
    return getYupSchema["s3_connector"];
  }

  return getYupSchema[name as keyof typeof getYupSchema];
}
