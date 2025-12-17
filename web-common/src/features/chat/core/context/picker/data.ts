import { createQuery } from "@tanstack/svelte-query";
import { getValidMetricsViewsQueryOptions } from "@rilldata/web-common/features/dashboards/selectors.ts";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
import { derived, get, readable, writable, type Readable } from "svelte/store";
import {
  type InlineContext,
  InlineContextType,
} from "@rilldata/web-common/features/chat/core/context/inline-context.ts";
import {
  getDimensionDisplayName,
  getMeasureDisplayName,
} from "@rilldata/web-common/features/dashboards/filters/getDisplayName.ts";
import type {
  InlineContextPickerChildSection,
  InlineContextPickerParentOption,
  InlineContextPickerSection,
} from "@rilldata/web-common/features/chat/core/context/picker/types.ts";
import {
  getClientFilteredResourcesQueryOptions,
  ResourceKind,
} from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
import {
  getQueryServiceTableColumnsQueryOptions,
  type V1Resource,
} from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store.ts";
import {
  type ActiveFileArtifact,
  getActiveFileArtifactStore,
} from "@rilldata/web-common/features/entity-management/nav-utils.ts";
import { getLastUsedMetricsViewNameStore } from "@rilldata/web-common/features/chat/core/context/picker/get-last-used-metrics-view.ts";
import { getActiveMetricsViewNameStore } from "@rilldata/web-common/features/dashboards/nav-utils.ts";
import { featureFlags } from "@rilldata/web-common/features/feature-flags.ts";

/**
 * Creates a store that contains a 2-level list of options for each valid metrics view and sources/models.
 * 1. Chooses top level options based on the args provided. Currently, the args has toggle for metrics views and sources/models.
 * 2. Any asynchronous options for 2nd level lists are fetched based on the open status of the top level option.
 * 3. Active metrics view and active source/model are tracked separately and filled in the resolved options.
 */
function getPickerOptions(): Readable<InlineContextPickerParentOption[]> {
  const lastUsedMetricsViewStore = getLastUsedMetricsViewNameStore();
  const activeMetricsViewStore = getActiveMetricsViewNameStore();
  const activeFileArtifactStore = getActiveFileArtifactStore();
  const isRillDev = !get(featureFlags.adminServer);

  return derived(
    [
      // Stable parts that only change when the underlying data is refetched.
      // Open store is created here to maintain the user selection.
      getMetricsViewPickerOptions(),
      isRillDev ? getModelsPickerOptions() : readable(null),
    ],
    ([metricsViewOptions, filesOption], set) => {
      const allOptions = [metricsViewOptions, filesOption]
        .filter(Boolean)
        .flat() as InlineContextPickerParentOption[];
      const subOptionStores = allOptions.map((o) =>
        o.childrenQueryOptions
          ? createQuery(o.childrenQueryOptions, queryClient)
          : readable(null),
      );

      // Unstable parts that are rerun when active entity changes or query status changes.
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
          const resolvedOptions = new Array<InlineContextPickerParentOption>(
            subOptionStoresResp.length,
          );
          subOptionStoresResp.forEach((subOptionStore, i) => {
            resolvedOptions[i] = {
              ...allOptions[i],
              children: subOptionStore?.data ?? allOptions[i].children ?? [],
              childrenLoading: subOptionStore?.isPending ?? false,
            };

            fillInRecentlyUsedOrActiveStatus(resolvedOptions[i], {
              lastUsedMv,
              activeMv,
              activeFileArtifact,
            });
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

/**
 * Creates a store that contains a list of options that match the search text.
 * 1. Directly calls {@link getPickerOptions} to get the initial list of options.
 * 2. Bubbles up the recently used and active top level options to the top of the list.
 * 3. Removes any top level options that don't match the search text including any 2nd level options within it.
 */
export function getFilteredPickerOptions(searchTextStore: Readable<string>) {
  return derived(
    [getPickerOptions(), searchTextStore],
    ([options, searchText]) => {
      const filterFunction = (label: string, value: string) =>
        searchText.length === 0 ||
        label.toLowerCase().includes(searchText.toLowerCase()) ||
        value.toLowerCase().includes(searchText.toLowerCase());

      let recentlyUsed: InlineContextPickerParentOption | null = null;
      let currentlyActive: InlineContextPickerParentOption | null = null;

      const filteredOptions = options
        .map((option) => {
          const children =
            option.children
              ?.map((cc) => ({
                type: cc.type,
                options: cc.options.filter((c) =>
                  filterFunction(c.label ?? "", c.value),
                ),
              }))
              .filter((cc) => cc.options.length > 0) ?? [];

          const parentMatches = filterFunction(
            option.context.label ?? "",
            option.context.value,
          );

          if (!parentMatches && children.length === 0) return null;

          if (!recentlyUsed && option.recentlyUsed) {
            recentlyUsed = option;
          }
          if (!currentlyActive && option.currentlyActive) {
            currentlyActive = option;
          }
          if (option.recentlyUsed || option.currentlyActive) return null; // these are added explicitly

          return {
            ...option,
            children,
          } satisfies InlineContextPickerParentOption;
        })
        .filter(Boolean) as InlineContextPickerParentOption[];

      if (recentlyUsed === currentlyActive) currentlyActive = null;

      const topSectionOptions = (
        [
          recentlyUsed,
          currentlyActive,
        ] as (InlineContextPickerParentOption | null)[]
      ).filter(Boolean) as InlineContextPickerParentOption[];
      const topSection = {
        type: "topSection",
        options: topSectionOptions,
      } satisfies InlineContextPickerSection;

      const otherSections = splitParentOptionsIntoSections(filteredOptions);

      return [topSection, ...otherSections].filter((s) => s.options.length > 0);
    },
  );
}

/**
 * Creates a store that contains a 2-level list of options for each valid metrics view.
 * 1st level: metrics view context options
 * 2nd level: measures and dimensions options for each metrics view
 */
function getMetricsViewPickerOptions(): Readable<
  InlineContextPickerParentOption[]
> {
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
      const measuresSection = {
        type: InlineContextType.Measure,
        options: measures,
      } satisfies InlineContextPickerChildSection;

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
      const dimensionsSection = {
        type: InlineContextType.Dimension,
        options: dimensions,
      } satisfies InlineContextPickerChildSection;

      return {
        context: {
          type: InlineContextType.MetricsView,
          metricsView: mvName,
          label: mvDisplayName,
          value: mvName,
        },
        openStore: writable(false),
        children: [measuresSection, dimensionsSection],
      } satisfies InlineContextPickerParentOption;
    });
  });
}

/**
 * Creates a store that contains a 2-level list of sources/model resources.
 * 1st level: section for sources/models.
 * 2nd level: all the columns in the source/model resource.
 * NOTE: this only lists resources that are parsed as sources/models. Any parse errors will exclude the file.
 */
function getModelsPickerOptions(): Readable<InlineContextPickerParentOption[]> {
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
        } satisfies InlineContextPickerParentOption;
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
          select: (data) => {
            const options =
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
              ) ?? [];
            return [
              {
                type: InlineContextType.Column,
                options,
              } satisfies InlineContextPickerChildSection,
            ];
          },
        },
      },
    ),
  );
}

export const ParentPickerTypes = new Set([
  InlineContextType.MetricsView,
  InlineContextType.Model,
]);

function fillInRecentlyUsedOrActiveStatus(
  option: InlineContextPickerParentOption,
  {
    lastUsedMv,
    activeMv,
    activeFileArtifact,
  }: {
    lastUsedMv: string | null;
    activeMv: string | null;
    activeFileArtifact: ActiveFileArtifact;
  },
) {
  if (option.context.type === InlineContextType.MetricsView) {
    option.recentlyUsed = lastUsedMv === option.context.metricsView;
    option.currentlyActive = activeMv === option.context.metricsView;
  } else if (option.context.type === InlineContextType.Model) {
    option.currentlyActive =
      activeFileArtifact.resource?.kind === ResourceKind.Model &&
      activeFileArtifact.resource?.name === option.context.model;
  }
}

function splitParentOptionsIntoSections(
  options: InlineContextPickerParentOption[],
) {
  if (options.length === 0) return [];

  let lastSection: InlineContextPickerSection | null = null;
  const sections: InlineContextPickerSection[] = [];
  options.forEach((option) => {
    if (
      lastSection === null ||
      option.context.type !== lastSection.options[0].context.type
    ) {
      lastSection = {
        type: option.context.type,
        options: [],
      } satisfies InlineContextPickerSection;
      sections.push(lastSection);
    }
    lastSection.options.push(option);
  });
  return sections;
}
