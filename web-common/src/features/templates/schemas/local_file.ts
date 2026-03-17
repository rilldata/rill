import type { MultiStepFormSchema } from "./types";
import {
  PossibleFileExtensions,
  PossibleZipExtensions,
} from "@rilldata/web-common/features/sources/modal/possible-file-extensions.ts";

export const localFileSchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  title: "Local File",
  "x-category": "fileStore",
  properties: {
    file: {
      type: "object",
      title: "Path",
      description: "Local file path or glob (relative to project root)",
      "x-display": "file",
      "x-file-accept": [
        ...PossibleFileExtensions,
        ...PossibleZipExtensions,
      ].join(","),
      "x-step": "source",
    },
  },
  required: ["path", "name"],
};
