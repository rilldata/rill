<script lang="ts">
  import { page } from "$app/stores";
  import BaseBookmarkForm from "@rilldata/web-admin/features/bookmarks/BaseBookmarkForm.svelte";
  import type { BookmarkFormValues } from "@rilldata/web-admin/features/bookmarks/form-utils";
  import Dialog from "@rilldata/web-common/components/dialog/Dialog.svelte";
  import {
    createAdminServiceUpdateBookmark,
    getAdminServiceListBookmarksQueryKey,
  } from "@rilldata/web-admin/client";
  import { Button } from "@rilldata/web-common/components/button";
  import { getBookmarkDataForDashboard } from "@rilldata/web-admin/features/bookmarks/getBookmarkDataForDashboard";
  import {
    type BookmarkEntry,
    useProjectId,
  } from "@rilldata/web-admin/features/bookmarks/selectors";
  import { notifications } from "@rilldata/web-common/components/notifications";
  import { useDashboardStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { createForm } from "svelte-forms-lib";
  import * as yup from "yup";

  export let open: boolean;
  export let metricsViewName: string;
  export let bookmark: BookmarkEntry;

  const queryClient = useQueryClient();

  $: dashboardStore = useDashboardStore(metricsViewName);

  $: projectId = useProjectId($page.params.organization, $page.params.project);

  const bookmarkUpdater = createAdminServiceUpdateBookmark();

  const formState = createForm({
    initialValues: <BookmarkFormValues>{
      displayName: bookmark.resource.displayName ?? "",
      description: bookmark.resource.description ?? "",
      filtersOnly: bookmark.filtersOnly,
      absoluteTimeRange: bookmark.absoluteTimeRange,
    },
    validationSchema: yup.object({
      displayName: yup.string().required("Required"),
      description: yup.string(),
    }),
    onSubmit: async (values) => {
      await $bookmarkUpdater.mutateAsync({
        data: {
          bookmarkId: bookmark.resource.id,
          displayName: values.displayName,
          description: values.description,
          data: getBookmarkDataForDashboard(
            $dashboardStore,
            values.filtersOnly,
            values.absoluteTimeRange,
          ),
        },
      });
      queryClient.refetchQueries(
        getAdminServiceListBookmarksQueryKey({
          projectId: $projectId.data ?? "",
          resourceKind: ResourceKind.MetricsView,
          resourceName: metricsViewName,
        }),
      );
      notifications.send({
        message: "Bookmark updated",
      });
      handleClose();
    },
  });

  const { handleSubmit } = formState;

  function handleClose() {
    open = false;
  }
</script>

<Dialog on:close={handleClose} {open} widthOverride="w-[602px]">
  <svelte:fragment slot="title">Bookmark current view</svelte:fragment>
  <BaseBookmarkForm editForm {formState} {metricsViewName} slot="body" />
  <div class="flex flex-row mt-4 gap-2" slot="footer">
    <div class="grow" />
    <Button on:click={handleClose} type="secondary">Cancel</Button>
    <Button on:click={handleSubmit} type="primary">Save</Button>
  </div>
</Dialog>
