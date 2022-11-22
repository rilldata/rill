import { useRuntimeServiceListFiles } from "@rilldata/web-common/runtime-client";

/**
 * Calls {@link useRuntimeServiceListFiles} using glob to select only sources.
 * Returns just the source names from the files.
 */
export function useSourceNames(repoId: string) {
  return useRuntimeServiceListFiles(
    repoId,
    {
      glob: "sources/*.{sql,yaml}",
    },
    {
      query: {
        refetchInterval: 1000,
        select: (data) =>
          data.paths
            ?.filter((path) => path.includes("sources/"))
            .map((path) => path.replace("/sources/", "").replace(".yaml", "")),
      },
    }
  );
}
