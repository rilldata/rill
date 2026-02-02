import type { MultiStepFormSchema } from "../../templates/schemas/types";
import { azureSchema } from "../../templates/schemas/azure";
import { gcsSchema } from "../../templates/schemas/gcs";
import { s3Schema } from "../../templates/schemas/s3";

export const multiStepFormSchemas: Record<string, MultiStepFormSchema> = {
  s3: s3Schema,
  gcs: gcsSchema,
  azure: azureSchema,
};

export function getConnectorSchema(
  connectorName: string,
): MultiStepFormSchema | null {
  const schema =
    multiStepFormSchemas[connectorName as keyof typeof multiStepFormSchemas];
  if (!schema?.properties) return null;
  return schema;
}
