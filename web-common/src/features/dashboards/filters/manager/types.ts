import type { V1Expression } from "@rilldata/web-common/runtime-client";

export type RawParsedFilter = {
  expr: V1Expression | undefined;
  dimensionsWithInlistFilter: string[];
};
