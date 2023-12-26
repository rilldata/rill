import { createRuntimeServiceListFiles } from "@rilldata/web-common/runtime-client";

/**
 * In dev mode we still need to get the files for left nav.
 * This is because parse errors will not have a resource.
 */
export function useMainEntityFiles(
  instanceId: string,
  prefix: "sources" | "models" | "dashboards"
) {
  let extension: string;
  switch (prefix) {
    case "sources":
    case "dashboards":
      extension = ".yaml";
      break;

    case "models":
      extension = ".sql";
  }

  return createRuntimeServiceListFiles(
    instanceId,
    {
      // We still use opinionated folder names. So we still need this filter
      glob: "{sources,models,dashboards}/*.{yaml,sql}",
    },
    {
      query: {
        select: (data) =>
          data.paths
            ?.filter((path) => path.includes(`${prefix}/`))
            .map((path) =>
              path.replace(`/${prefix}/`, "").replace(extension, "")
            )
            // sort alphabetically case-insensitive
            .sort((a, b) =>
              a.localeCompare(b, undefined, { sensitivity: "base" })
            ) ?? [],
      },
    }
  );
}
