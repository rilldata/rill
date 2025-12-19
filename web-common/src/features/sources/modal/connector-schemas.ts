import type { MultiStepFormSchema } from "../../templates/schemas/types";
import { azureSchema } from "../../templates/schemas/azure";
import { gcsSchema } from "../../templates/schemas/gcs";
import { httpsSchema } from "../../templates/schemas/https";
import { s3Schema } from "../../templates/schemas/s3";

export const multiStepFormSchemas: Record<string, MultiStepFormSchema> = {
  s3: s3Schema,
  gcs: gcsSchema,
  azure: azureSchema,
  https: httpsSchema,
};

export function getConnectorSchema(
  connectorName: string,
): MultiStepFormSchema | null {
  const schema =
    multiStepFormSchemas[connectorName as keyof typeof multiStepFormSchemas];
  if (!schema?.properties) return null;
  return schema;
}

export function isStepMatch(
  schema: MultiStepFormSchema | null,
  key: string,
  step?: "connector" | "source" | string,
): boolean {
  if (!schema?.properties) return false;
  const prop = schema.properties[key];
  if (!prop) return false;
  if (!step) return true;
  const propStep = prop["x-step"];
  if (!propStep) return true;
  return propStep === step;
}
