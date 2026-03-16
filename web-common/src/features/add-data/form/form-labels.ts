import type { MultiStepFormSchema } from "@rilldata/web-common/features/templates/schemas/types.ts";
import { getSchemaButtonLabels } from "@rilldata/web-common/features/templates/schema-utils.ts";
import { ImportDataStep } from "@rilldata/web-common/features/add-data/steps/types.ts";

export type AddDataFormLabels = {
  primaryLoadingCopy: string;
  primaryButtonLabel: string;

  yamlPreviewTitle: string;
};

export const defaultFormLabels: AddDataFormLabels = {
  primaryLoadingCopy: "Saving...",
  primaryButtonLabel: "Save",

  yamlPreviewTitle: "YAML Preview",
};

const connectorFormLabels: AddDataFormLabels = {
  primaryLoadingCopy: "Testing connection...",
  primaryButtonLabel: "Test and Connect",

  yamlPreviewTitle: "Connector preview",
};

export function getLabelsForConnector(
  schema: MultiStepFormSchema | null,
  values: Record<string, unknown>,
) {
  const schemaSpecificLabels = getSchemaButtonLabels(schema, values);
  if (!schemaSpecificLabels) return connectorFormLabels;
  // Merge with legacy label calculations.
  // TODO: refactor getSchemaButtonLabels to output new labels object
  return {
    ...connectorFormLabels,
    primaryLoadingCopy: schemaSpecificLabels.loading,
    primaryButtonLabel: schemaSpecificLabels.idle,
  };
}

const importOnlySourceFormLabels: AddDataFormLabels = {
  primaryLoadingCopy: "Importing data...",
  primaryButtonLabel: "Import Data",

  yamlPreviewTitle: "Model preview",
};
const importAndGenerateSourceFormLabels: AddDataFormLabels = {
  primaryLoadingCopy: "Generating dashboard...",
  primaryButtonLabel: "Generate dashboard with AI",

  yamlPreviewTitle: "Model preview",
};

export function getLabelsForSource(steps: ImportDataStep[]) {
  const hasOnlyCreateStep =
    steps.length === 1 && steps[0] === ImportDataStep.CreateModel;
  return hasOnlyCreateStep
    ? importOnlySourceFormLabels
    : importAndGenerateSourceFormLabels;
}
