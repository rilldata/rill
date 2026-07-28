<script lang="ts">
  import { page } from "$app/state";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import {
    createAdminServiceCreateBookmark,
    createAdminServiceUpdateBookmark,
    type V1Bookmark,
  } from "@rilldata/web-admin/client";
  import { getBookmarkData } from "@rilldata/web-admin/features/bookmarks/utils.ts";
  import ProjectAccessControls from "@rilldata/web-admin/features/projects/ProjectAccessControls.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import Label from "@rilldata/web-common/components/forms/Label.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import ExploreFilterChipsReadOnly from "@rilldata/web-common/features/dashboards/filters/ExploreFilterChipsReadOnly.svelte";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";
  import { InfoIcon } from "lucide-svelte";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import CanvasFilterChipsReadOnly from "@rilldata/web-common/features/dashboards/filters/CanvasFilterChipsReadOnly.svelte";
  import { defaults, superForm } from "sveltekit-superforms";
  import { yup } from "sveltekit-superforms/adapters";
  import { object, string, boolean } from "yup";
  import { getRpcErrorMessage } from "@rilldata/web-admin/components/errors/error-utils.ts";
  import { FilterStateFromUrl } from "@rilldata/web-admin/features/project-wide-bookmarks/FilterStateFromUrl.svelte.ts";
  import { getDashboardResourceFromPage } from "@rilldata/web-common/features/dashboards/nav-utils.ts";
  import { ParsedBookmark } from "@rilldata/web-admin/features/project-wide-bookmarks/ParsedBookmark.svelte.ts";
  import { invalidateBookmarksQuery } from "@rilldata/web-admin/features/project-wide-bookmarks/utils.ts";
  import { ExploreSpecProvider } from "@rilldata/web-common/features/dashboards/ExploreSpecProvider.svelte.ts";

  let {
    organization,
    project,
    projectId,
    bookmark,
    onClose,
  }: {
    organization: string;
    project: string;
    projectId: string;
    bookmark: V1Bookmark | undefined;
    onClose: () => void;
  } = $props();

  const runtimeClient = useRuntimeClient();

  let parsedBookmark = $derived(
    bookmark ? new ParsedBookmark(runtimeClient, bookmark) : undefined,
  );

  const filterStateFromUrl = new FilterStateFromUrl(runtimeClient);

  let resource = $derived(getDashboardResourceFromPage(page));
  let showFiltersOnly = $derived(resource?.kind === ResourceKind.Explore);

  let exploreSpecProvider = $derived(
    resource?.kind === ResourceKind.Explore
      ? new ExploreSpecProvider(runtimeClient, resource.name)
      : undefined,
  );
  let exploreDefaultUrlParams = $derived(
    exploreSpecProvider?.getDefaultUrlParams(),
  );

  const bookmarkCreator = createAdminServiceCreateBookmark();
  const bookmarkUpdater = createAdminServiceUpdateBookmark();

  const initialValues = {
    // eslint-disable-next-line svelte/valid-compile
    displayName: bookmark?.displayName || m.bookmark_default_label(),
    // eslint-disable-next-line svelte/valid-compile
    description: bookmark?.description ?? "",
    // eslint-disable-next-line svelte/valid-compile
    shared: bookmark?.shared ? "true" : "false",
    // eslint-disable-next-line svelte/valid-compile
    filtersOnly: parsedBookmark?.filtersOnly ?? false,
    // eslint-disable-next-line svelte/valid-compile
    absoluteTimeRange: parsedBookmark?.absoluteTimeRange ?? false,
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

        if (bookmark) {
          await $bookmarkUpdater.mutateAsync({
            data: {
              bookmarkId: bookmark.id,
              displayName: values.displayName,
              description: values.description,
              shared: values.shared === "true",
            },
          });
        } else {
          // Bookmark data is only updated when creating a bookmark
          const bookmarkData = getBookmarkData({
            curUrlParams: page.url.searchParams,
            defaultUrlParams: exploreDefaultUrlParams,
            filtersOnly: values.filtersOnly,
            absoluteTimeRange: values.absoluteTimeRange,
          });
          await $bookmarkCreator.mutateAsync({
            data: {
              displayName: values.displayName,
              description: values.description,
              projectId,
              resourceKind: resource.kind,
              resourceName: resource.name,
              shared: values.shared === "true",
              urlSearch: bookmarkData,
            },
          });
          reset();
        }
        onClose();

        await invalidateBookmarksQuery(projectId);
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
        {#if filterStateFromUrl.filterState && "uiFilters" in filterStateFromUrl.filterState}
          <CanvasFilterChipsReadOnly
            col={false}
            uiFilters={filterStateFromUrl.filterState.uiFilters}
            timeRangeString={filterStateFromUrl.filterState.displayTimeRange
              .expression}
            comparisonRange={undefined}
            timeStart={filterStateFromUrl.filterState.queryTimeStart}
            timeEnd={filterStateFromUrl.filterState.queryTimeEnd}
          />
        {:else if filterStateFromUrl.filterState && "filters" in filterStateFromUrl.filterState}
          <ExploreFilterChipsReadOnly
            filters={filterStateFromUrl.filterState.filters}
            dimensionsWithInlistFilter={filterStateFromUrl.filterState
              .dimensionsWithInlistFilter}
            dimensionThresholdFilters={filterStateFromUrl.filterState
              .dimensionThresholdFilters}
            displayTimeRange={filterStateFromUrl.filterState.displayTimeRange}
            queryTimeStart={filterStateFromUrl.filterState.queryTimeStart}
            queryTimeEnd={filterStateFromUrl.filterState.queryTimeEnd}
            metricsViewNames={exploreSpecProvider.metricsViewNames ?? []}
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
          {#if filterStateFromUrl.filterState}
            <div class="text-fg-secondary text-sm">
              {filterStateFromUrl.filterState.selectedTimeRange}
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
