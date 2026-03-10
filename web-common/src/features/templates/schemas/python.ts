import type { MultiStepFormSchema } from "./types";

export const pythonSchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  title: "Python",
  "x-category": "fileStore",
  "x-button-labels": {
    "*": { "*": { idle: "Continue", loading: "Continuing..." } },
  },
  properties: {
    code_path: {
      type: "string",
      title: "Script path",
      description:
        "Path to the Python script relative to the project root. The script must write a Parquet file to the RILL_OUTPUT_PATH environment variable.",
      "x-placeholder": "scripts/extract.py",
      "x-step": "source",
    },
    name: {
      type: "string",
      title: "Model name",
      description: "Name of the model",
      "x-placeholder": "my_python_model",
      "x-step": "source",
    },
  },
  required: ["code_path", "name"],
};
