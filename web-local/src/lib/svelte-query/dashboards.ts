import { useRuntimeServiceListFiles } from "@rilldata/web-common/runtime-client";

export function useDashboardNames(repoId: string) {
  return useRuntimeServiceListFiles(
    repoId,
    {
      glob: "{sources,models,dashboards}/*.{yaml,sql}",
    },
    {
      query: {
        refetchInterval: 1000,
        select: (data) =>
          data.paths
            ?.filter((path) => path.includes("dashboards/"))
            .map((path) =>
              path.replace("/dashboards/", "").replace(".yaml", "")
            ),
      },
    }
  );
}
