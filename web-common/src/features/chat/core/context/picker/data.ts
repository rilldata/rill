import { createQuery } from "@tanstack/svelte-query";
import { getValidMetricsViewsQueryOptions } from "@rilldata/web-common/features/dashboards/selectors.ts";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
import { derived, readable, writable, type Readable } from "svelte/store";
import {
  type InlineContext,
  InlineContextType,
} from "@rilldata/web-common/features/chat/core/context/inline-context.ts";
import {
  getDimensionDisplayName,
  getMeasureDisplayName,
} from "@rilldata/web-common/features/dashboards/filters/getDisplayName.ts";
import type { InlineContextPickerOption } from "@rilldata/web-common/features/chat/core/context/picker/types.ts";
import {
  getClientFilteredResourcesQueryOptions,
  ResourceKind,
} from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
import {
  getQueryServiceTableColumnsQueryOptions,
  type V1Resource,
} from "@rilldata/web-common/runtime-client";
import { splitFolderAndFileName } from "@rilldata/web-common/features/entity-management/file-path-utils.ts";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store.ts";
import { getActiveFileArtifactStore } from "@rilldata/web-common/features/entity-management/nav-utils.ts";
import { getLastUsedMetricsViewNameStore } from "@rilldata/web-common/features/chat/core/context/picker/get-last-used-metrics-view.ts";
import { getActiveMetricsViewNameStore } from "@rilldata/web-common/features/dashboards/nav-utils.ts";

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
        openStore: writable(false),
        children: [measures, dimensions],
      } satisfies InlineContextPickerOption;
    });
  });
}

/**
 * Creates a store that contains a 2-level list of sources/model resources.
 * 1st level: section for sources/models.
 * 2nd level: all the columns in the source/model resource.
 * NOTE: this only lists resources that are parsed as sources/models. Any parse errors will exlcude the file.
 */
function getModelsPickerOptions(): Readable<InlineContextPickerOption[]> {
  const modelResourcesQuery = createQuery(
    getClientFilteredResourcesQueryOptions(ResourceKind.Model),
    queryClient,
  );

  return derived(
    [runtime, modelResourcesQuery],
    ([{ instanceId }, modelResourcesResp]) => {
      const models = modelResourcesResp.data ?? [];
      return models.map((res) => {
        const modelName = res.meta?.name?.name ?? "";

        const openStore = writable(false);

        return {
          context: {
            type: InlineContextType.Model,
            model: modelName,
            value: modelName,
            label: modelName,
          },
          openStore,
          childrenQueryOptions: getModelColumnsQueryOptions(
            instanceId,
            res,
            modelName,
            openStore,
          ),
        } satisfies InlineContextPickerOption;
      });
    },
  );
}

function getModelColumnsQueryOptions(
  instanceId: string,
  modelRes: V1Resource | undefined,
  modelName: string,
  enabledStore: Readable<boolean>,
) {
  const connector = modelRes?.model?.spec?.outputConnector ?? "";
  const table = modelRes?.model?.state?.resultTable ?? "";
  return derived(enabledStore, (enabled) =>
    getQueryServiceTableColumnsQueryOptions(
      instanceId,
      table,
      {
        connector,
      },
      {
        query: {
          enabled: enabled && Boolean(table),
          select: (data) => [
            data.profileColumns?.map(
              (col) =>
                ({
                  type: InlineContextType.Column,
                  label: col.name,
                  value: col.name!,
                  column: col.name,
                  columnType: col.type,
                  model: modelName,
                }) satisfies InlineContext,
            ) ?? [],
          ],
        },
      },
    ),
  );
}

export type PickerArgs = {
  metricViews: boolean;
  files: boolean;
};
export const ParentPickerTypes = new Set([
  InlineContextType.MetricsView,
  InlineContextType.Model,
]);

function getPickerOptions({
  metricViews,
  files,
}: PickerArgs): Readable<InlineContextPickerOption[]> {
  const lastUsedMetricsViewStore = getLastUsedMetricsViewNameStore();
  const activeMetricsViewStore = getActiveMetricsViewNameStore();
  const activeFileArtifactStore = getActiveFileArtifactStore();

  return derived(
    [
      metricViews ? getMetricsViewPickerOptions() : readable(null),
      files ? getModelsPickerOptions() : readable(null),
    ],
    ([metricsViewOptions, filesOption], set) => {
      const allOptions = [metricsViewOptions, filesOption]
        .filter(Boolean)
        .flat() as InlineContextPickerOption[];
      const subOptionStores = allOptions.map((o) =>
        o.childrenQueryOptions
          ? createQuery(o.childrenQueryOptions, queryClient)
          : readable(null),
      );

      const resolvedOptionsStore = derived(
        [
          lastUsedMetricsViewStore,
          activeMetricsViewStore,
          activeFileArtifactStore,
          ...subOptionStores,
        ],
        ([
          lastUsedMv,
          activeMv,
          activeFileArtifact,
          ...subOptionStoresResp
        ]) => {
          const resolvedOptions = new Array<InlineContextPickerOption>(
            subOptionStoresResp.length,
          );
          subOptionStoresResp.forEach((subOptionStore, i) => {
            resolvedOptions[i] = {
              ...allOptions[i],
              children: subOptionStore?.data ?? allOptions[i].children ?? [],
            };

            if (
              resolvedOptions[i].context.type === InlineContextType.MetricsView
            ) {
              resolvedOptions[i].recentlyUsed =
                lastUsedMv === resolvedOptions[i].context.metricsView;
              resolvedOptions[i].currentlyActive =
                activeMv === resolvedOptions[i].context.metricsView;
            } else if (
              resolvedOptions[i].context.type === InlineContextType.Model
            ) {
              resolvedOptions[i].currentlyActive =
                activeFileArtifact.resource?.kind === ResourceKind.Model &&
                activeFileArtifact.resource?.name ===
                  resolvedOptions[i].context.model;
            }
          });
          return resolvedOptions;
        },
      );

      return resolvedOptionsStore.subscribe((resolvedOptions) =>
        set(resolvedOptions),
      );
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

      let recentlyUsed: InlineContextPickerOption | null = null;
      let currentlyActive: InlineContextPickerOption | null = null;

      const filteredOptions = options.map((option) => {
        const children =
          option.children
            ?.map((cc) =>
              cc.filter((c) => filterFunction(c.label ?? "", c.value)),
            )
            .filter((cc) => cc.length > 0) ?? [];

        const parentMatches = filterFunction(
          option.context.label ?? "",
          option.context.value,
        );

        if (!parentMatches && children.length === 0) return null;

        if (option.recentlyUsed) recentlyUsed = option;
        if (option.currentlyActive) currentlyActive = option;
        if (option.recentlyUsed || option.currentlyActive) return null; // these are added explicitly

        return {
          ...option,
          children,
        } satisfies InlineContextPickerOption;
      });

      if (recentlyUsed === currentlyActive) currentlyActive = null;

      return [recentlyUsed, currentlyActive, ...filteredOptions].filter(
        Boolean,
      ) as InlineContextPickerOption[];
    },
  );
}
