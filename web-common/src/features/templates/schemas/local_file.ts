import type { MultiStepFormSchema } from "./types";
import {
  PossibleFileExtensions,
  PossibleZipExtensions,
} from "@rilldata/web-common/features/sources/modal/possible-file-extensions.ts";
import {
  UploadFileSizeLimitInBytes,
  UploadSizeExceededIsWarning,
} from "@rilldata/web-common/features/sources/upload-utils.ts";

export const localFileSchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  title: "Local File",
  "x-category": "fileStore",
  properties: {
    file: {
      title: "Path",
      description: "Local file path or glob (relative to project root)",
      "x-display": "file",
      "x-file-accept": [
        ...PossibleFileExtensions,
        ...PossibleZipExtensions,
      ].join(","),
      "x-step": "source",
      "x-hint": "CSV, TSV, Parquet, TXT or JSON",
      "x-file-size-limit": UploadFileSizeLimitInBytes,
      "x-file-size-soft-limit": UploadSizeExceededIsWarning,
      "x-file-size-limit-warning-message":
        "Files over 100MB can be used locally but development to Rill Cloud is not allowed. Consider storing the data externally (e.g. S3) if you plan to deploy this project",
    },
  },
  required: ["file"],
};
