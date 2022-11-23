import { useRuntimeServiceListFiles } from "@rilldata/web-common/runtime-client";

export function useModelNames(repoId: string) {
  return useRuntimeServiceListFiles(
    repoId,
    {
      glob: "models/*.sql",
    },
    {
      query: {
        refetchInterval: 1000,
        select: (data) =>
          data.paths
            ?.filter((path) => path.includes("models/"))
            .map((path) => path.replace("/models/", "").replace(".sql", "")),
      },
    }
  );
}
