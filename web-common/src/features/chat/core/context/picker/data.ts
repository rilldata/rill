import { createQuery } from "@tanstack/svelte-query";
import { getValidMetricsViewsQueryOptions } from "@rilldata/web-common/features/dashboards/selectors.ts";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
import { derived, readable, type Readable } from "svelte/store";
import {
  type InlineContext,
  InlineContextType,
} from "@rilldata/web-common/features/chat/core/context/inline-context.ts";
import {
  getDimensionDisplayName,
  getMeasureDisplayName,
} from "@rilldata/web-common/features/dashboards/filters/getDisplayName.ts";
import type { InlineContextPickerOption } from "@rilldata/web-common/features/chat/core/context/picker/types.ts";
import { getStableListFilesQueryOptions } from "@rilldata/web-common/features/entity-management/file-selectors.ts";
import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts.ts";

/**
 * Creates a store that contains a 2-level list of options for each valid metrics view.
 * 1st level: metrics view context options
 * 2nd level: measures and dimensions options for each metrics view
 */
function getMetricsViewPickerOptions(): Readable<InlineContextPickerOption[]> {
  const metricsViewsQuery = createQuery(
    getValidMetricsViewsQueryOptions(),
    queryClient,
  );

  return derived(metricsViewsQuery, (metricsViewsResp) => {
    const metricsViews = metricsViewsResp.data ?? [];
    return metricsViews.map((mv) => {
      const mvName = mv.meta?.name?.name ?? "";
      const metricsViewSpec = mv.metricsView?.state?.validSpec ?? {};
      const mvDisplayName = metricsViewSpec?.displayName || mvName;

      const measures =
        metricsViewSpec?.measures?.map(
          (m) =>
            ({
              type: InlineContextType.Measure,
              label: getMeasureDisplayName(m),
              value: m.name!,
              metricsView: mvName,
              measure: m.name!,
            }) satisfies InlineContext,
        ) ?? [];

      const dimensions =
        metricsViewSpec?.dimensions?.map(
          (d) =>
            ({
              type: InlineContextType.Dimension,
              label: getDimensionDisplayName(d),
              value: d.name!,
              metricsView: mvName,
              dimension: d.name!,
            }) satisfies InlineContext,
        ) ?? [];

      return {
        context: {
          type: InlineContextType.MetricsView,
          metricsView: mvName,
          label: mvDisplayName,
          value: mvName,
        },
        childContextCategories: [measures, dimensions],
      } satisfies InlineContextPickerOption;
    });
  });
}

/**
 * Creates a store that contains a 2-level list of resource files.
 * 1st level: section for files
 * 2nd level: all the resource files in the projects.
 */
function getFilesPickerOptions() {
  const filesQuery = createQuery(getStableListFilesQueryOptions(), queryClient);

  return derived(filesQuery, (filesResp) => {
    const files = filesResp.data?.files ?? [];
    const options = files
      .map((file) => {
        const filePath = file.path ?? "";
        if (file.isDir || !fileArtifacts.hasFileArtifact(filePath)) return null;

        return {
          type: InlineContextType.File,
          filePath,
          value: filePath,
          label: filePath.split("/").pop() ?? "",
        } satisfies InlineContext;
      })
      .filter(Boolean) as InlineContext[];
    return {
      context: {
        type: InlineContextType.Files,
        label: "Files",
        value: "files",
      },
      childContextCategories: [options],
    } satisfies InlineContextPickerOption;
  });
}

export type PickerArgs = {
  metricViews: boolean;
  files: boolean;
};
export const ParentPickerTypes = new Set([
  InlineContextType.MetricsView,
  InlineContextType.Files,
]);

function getPickerOptions({ metricViews, files }: PickerArgs) {
  return derived(
    [
      metricViews ? getMetricsViewPickerOptions() : readable(null),
      files ? getFilesPickerOptions() : readable(null),
    ],
    ([metricsViewOptions, filesOption]) => {
      return [metricsViewOptions, filesOption]
        .filter(Boolean)
        .flat() as InlineContextPickerOption[];
    },
  );
}

export function getFilterPickerOptions(
  args: PickerArgs,
  searchTextStore: Readable<string>,
) {
  return derived(
    [getPickerOptions(args), searchTextStore],
    ([options, searchText]) => {
      const filterFunction = (label: string, value: string) =>
        searchText.length === 0 ||
        label.toLowerCase().includes(searchText.toLowerCase()) ||
        value.toLowerCase().includes(searchText.toLowerCase());

      return options
        .map((option) => {
          const childOptions =
            option.childContextCategories
              ?.map((cc) =>
                cc.filter((c) => filterFunction(c.label ?? "", c.value)),
              )
              .filter((cc) => cc.length > 0) ?? [];

          const parentMatches = filterFunction(
            option.context.label ?? "",
            option.context.value,
          );

          if (!parentMatches && childOptions.length === 0) return null;

          return {
            context: option.context,
            childContextCategories: childOptions,
          } satisfies InlineContextPickerOption;
        })
        .filter(Boolean) as InlineContextPickerOption[];
    },
  );
}
