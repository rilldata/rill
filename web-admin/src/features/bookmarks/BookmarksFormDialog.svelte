<script lang="ts">
  import { page } from "$app/state";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import {
    createAdminServiceCreateBookmark,
    createAdminServiceUpdateBookmark,
    getAdminServiceListBookmarksQueryKey,
  } from "@rilldata/web-admin/client";
  import {
    type BookmarkEntry,
    formatTimeRange,
    getBookmarkData,
  } from "@rilldata/web-admin/features/bookmarks/utils.ts";
  import ProjectAccessControls from "@rilldata/web-admin/features/projects/ProjectAccessControls.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import Label from "@rilldata/web-common/components/forms/Label.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { deriveInterval } from "@rilldata/web-common/features/dashboards/time-controls/new-time-controls";
  import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
  import {
    V1TimeGrain,
    type V1TimeRange,
  } from "@rilldata/web-common/runtime-client";
  import { InfoIcon } from "lucide-svelte";
  import type { Interval } from "luxon";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { defaults, superForm } from "sveltekit-superforms";
  import { yup } from "sveltekit-superforms/adapters";
  import { object, string, boolean } from "yup";
  import { getRpcErrorMessage } from "@rilldata/web-admin/components/errors/error-utils.ts";
  import ReadonlyExpressionFilters from "@rilldata/web-common/features/dashboards/filters/ReadonlyExpressionFilters.svelte";
  import { ExpressionFilterManager } from "@rilldata/web-common/features/dashboards/filters/ExpressionFilterManager.svelte.ts";
  import {
    CanvasDashboardConfigProvider,
    ExploreDashboardConfigProvider,
  } from "@rilldata/web-common/features/dashboards/providers/DashboardConfigProvider.svelte.ts";

  let {
    organization,
    project,
    projectId,
    resource,
    bookmark,
    defaultUrlParams,
    showFiltersOnly = true,
    onClose,
  }: {
    organization: string;
    project: string;
    projectId: string;
    resource: { name: string; kind: ResourceKind };
    bookmark: BookmarkEntry | null;
    defaultUrlParams: URLSearchParams | undefined;
    showFiltersOnly?: boolean;
    onClose: () => void;
  } = $props();

  const runtimeClient = useRuntimeClient();

  let { name: resourceName, kind: resourceKind } = $derived(resource);
  let dashboardConfigProvider = $derived(
    resourceKind === ResourceKind.Canvas
      ? new CanvasDashboardConfigProvider(runtimeClient, resourceName)
      : new ExploreDashboardConfigProvider(runtimeClient, resourceName),
  );
  let expressionFilterManager = $derived(
    new ExpressionFilterManager(
      dashboardConfigProvider.metricsViewsProvider,
      dashboardConfigProvider.yamlConfigProvider,
    ),
  );
  let { setUrlParams } = $derived(expressionFilterManager);

  let curUrlParams = $derived(page.url.searchParams);
  $effect(() => setUrlParams(curUrlParams));

  let timeFilterState = $state<
    | {
        queryTimeStart: string;
        queryTimeEnd: string;
        displayTimeRange: V1TimeRange;
        selectedTimeRange: string;
      }
    | undefined
  >(undefined);

  $effect(() => void processTimeFromUrl());
  async function processTimeFromUrl() {
    const searchParamsObj = new URLSearchParams(curUrlParams);
    const rangeExpression = searchParamsObj.get(
      ExploreStateURLParams.TimeRange,
    );
    const timeRange = <V1TimeRange>{
      expression: rangeExpression || "",
    };

    const timeZone =
      searchParamsObj.get(ExploreStateURLParams.TimeZone) || "UTC";

    try {
      const promises =
        dashboardConfigProvider.metricsViewsProvider.metricsViewNames.map(
          (mvName) =>
            deriveInterval(
              timeRange.expression || "",
              runtimeClient,
              mvName,
              timeZone,
            ),
        );

      const intervals = await Promise.all(promises);
      let intervalWithLatestEndPoint:
        | {
            interval: Interval;
            grain?: V1TimeGrain | undefined;
            error?: string;
          }
        | undefined;
      intervals.forEach((response) => {
        if (
          !intervalWithLatestEndPoint ||
          (response.interval.end && intervalWithLatestEndPoint.interval.end
            ? response.interval.end > intervalWithLatestEndPoint.interval.end
            : false)
        ) {
          intervalWithLatestEndPoint = response;
        }
      });

      const start = intervalWithLatestEndPoint.interval.start.toISO();
      const end = intervalWithLatestEndPoint.interval.end.toISO();

      const grain =
        (searchParamsObj.get(ExploreStateURLParams.TimeGrain) as V1TimeGrain) ||
        intervalWithLatestEndPoint.grain ||
        V1TimeGrain.TIME_GRAIN_MINUTE;

      const selectedTimeRange = formatTimeRange(start, end, grain, timeZone);

      timeFilterState = {
        queryTimeStart: start,
        queryTimeEnd: end,
        displayTimeRange: timeRange,
        selectedTimeRange,
      };
    } catch {
      timeFilterState = undefined;
    }
  }

  const bookmarkCreator = createAdminServiceCreateBookmark();
  const bookmarkUpdater = createAdminServiceUpdateBookmark();

  // svelte-ignore state_referenced_locally
  const initialValues = {
    displayName: bookmark?.resource.displayName || m.bookmark_default_label(),
    description: bookmark?.resource.description ?? "",
    shared: bookmark?.resource.shared ? "true" : "false",
    filtersOnly: bookmark?.filtersOnly ?? false,
    absoluteTimeRange: bookmark?.absoluteTimeRange ?? false,
  };

  const schema = yup(
    object({
      displayName: string().required("Required"),
      description: string(),
      shared: string(),
      filtersOnly: boolean(),
      absoluteTimeRange: boolean(),
    }),
  );

  const formId = "create-bookmark-dialog";

  const { form, errors, submit, reset, enhance } = superForm(
    defaults(initialValues, schema),
    {
      SPA: true,
      validators: schema,
      async onUpdate({ form }) {
        if (!form.valid) return;
        const values = form.data;

        const bookmarkData = getBookmarkData({
          curUrlParams,
          defaultUrlParams,
          filtersOnly: values.filtersOnly,
          absoluteTimeRange: values.absoluteTimeRange,
        });

        if (bookmark) {
          await $bookmarkUpdater.mutateAsync({
            data: {
              bookmarkId: bookmark.resource.id,
              displayName: values.displayName,
              description: values.description,
              shared: values.shared === "true",
              urlSearch: bookmarkData,
            },
          });
        } else {
          await $bookmarkCreator.mutateAsync({
            data: {
              displayName: values.displayName,
              description: values.description,
              projectId,
              resourceKind,
              resourceName,
              shared: values.shared === "true",
              urlSearch: bookmarkData,
            },
          });
          reset();
        }
        onClose();

        await queryClient.refetchQueries({
          queryKey: getAdminServiceListBookmarksQueryKey({
            projectId,
            resourceKind,
            resourceName,
          }),
        });
        eventBus.emit("notification", {
          message: bookmark ? m.bookmark_updated() : m.bookmark_created(),
        });
      },
    },
  );

  let error = $derived(
    getRpcErrorMessage($bookmarkCreator.error ?? $bookmarkUpdater.error),
  );
</script>

<Dialog.Root
  open
  onOpenChange={(o) => {
    if (!o) onClose();
  }}
>
  <Dialog.Content>
    <Dialog.Header>
      <Dialog.Title>
        {bookmark ? m.bookmark_edit() : m.bookmark_current_view()}
      </Dialog.Title>
    </Dialog.Header>

    <form
      class="flex flex-col gap-4 z-50"
      id={formId}
      use:enhance
      onsubmit={(e) => {
        e.preventDefault();
        submit(e);
      }}
    >
      <Input
        bind:value={$form["displayName"]}
        errors={$errors["displayName"]}
        id="displayName"
        label={m.bookmark_label()}
      />
      <Input
        bind:value={$form["description"]}
        errors={$errors["description"]}
        id="description"
        label={m.bookmark_description()}
        optional
      />
      <div class="flex flex-col gap-y-2">
        <Label class="flex flex-col gap-y-1 text-sm">
          <div class="text-fg-primary font-medium">{m.bookmark_filters()}</div>
          <div class="text-fg-secondary">
            {m.bookmark_filters_inherited()}
          </div>
        </Label>
        {#if timeFilterState}
          <ReadonlyExpressionFilters
            {expressionFilterManager}
            displayTimeRange={timeFilterState.displayTimeRange}
            queryTimeStart={timeFilterState.queryTimeStart}
            queryTimeEnd={timeFilterState.queryTimeEnd}
          />
        {/if}
      </div>
      <ProjectAccessControls {organization} {project}>
        <Select
          bind:value={$form["shared"]}
          id="shared"
          label={m.bookmark_category()}
          options={[
            { value: "false", label: m.bookmark_your_bookmarks() },
            { value: "true", label: m.bookmark_managed_bookmarks() },
          ]}
          slot="manage-project"
          tooltip={m.bookmark_category_tooltip()}
        />
      </ProjectAccessControls>
      {#if showFiltersOnly}
        <div class="flex items-center space-x-2">
          <Switch
            bind:checked={$form["filtersOnly"]}
            id="filtersOnly"
            label={m.bookmark_filters_only_label()}
          />

          <Label
            class="font-normal flex gap-x-1 items-center"
            for="filtersOnly"
          >
            <span>{m.bookmark_save_filters_only()}</span>
            <Tooltip distance={8}>
              <InfoIcon class="text-fg-secondary" size="14px" strokeWidth={2} />
              <TooltipContent
                class="whitespace-pre-line"
                maxWidth="600px"
                slot="tooltip-content"
              >
                {m.bookmark_filters_only_tooltip()}
              </TooltipContent>
            </Tooltip>
          </Label>
        </div>
      {/if}
      <div class="flex items-center space-x-2">
        <Switch
          bind:checked={$form["absoluteTimeRange"]}
          id="absoluteTimeRange"
          label={m.bookmark_absolute_time_range()}
        />
        <Label class="flex flex-col font-normal" for="absoluteTimeRange">
          <div class="text-left text-sm flex gap-x-1 items-center">
            <span>{m.bookmark_absolute_time_range()}</span>
            <Tooltip distance={8}>
              <InfoIcon class="text-fg-secondary" size="14px" strokeWidth={2} />
              <TooltipContent
                class="whitespace-pre-line"
                maxWidth="600px"
                slot="tooltip-content"
              >
                {m.bookmark_absolute_time_tooltip()}
              </TooltipContent>
            </Tooltip>
          </div>
          {#if timeFilterState}
            <div class="text-fg-secondary text-sm">
              {timeFilterState.selectedTimeRange}
            </div>
          {/if}
        </Label>
      </div>
      {#if error}
        <div class="text-red-500 text-sm py-px">
          {error}
        </div>
      {/if}
    </form>

    <div class="flex flex-row mt-4 gap-2">
      <div class="grow"></div>
      <Button onClick={onClose} type="secondary">{m.common_cancel()}</Button>
      <Button onClick={submit} type="primary">{m.bookmark_save()}</Button>
    </div>
  </Dialog.Content>
</Dialog.Root>
