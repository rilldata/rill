<script lang="ts">
  import { page } from "$app/stores";
  import BaseBookmarkForm from "@rilldata/web-admin/features/bookmarks/BaseBookmarkForm.svelte";
  import type { BookmarkFormValues } from "@rilldata/web-admin/features/bookmarks/form-utils";
  import { useProjectId } from "@rilldata/web-admin/features/projects/selectors";
  import Dialog from "@rilldata/web-common/components/dialog/Dialog.svelte";
  import {
    createAdminServiceCreateBookmark,
    getAdminServiceListBookmarksQueryKey,
  } from "@rilldata/web-admin/client";
  import { Button } from "@rilldata/web-common/components/button";
  import { getBookmarkDataForDashboard } from "@rilldata/web-admin/features/bookmarks/getBookmarkDataForDashboard";
  import { notifications } from "@rilldata/web-common/components/notifications";
  import { useDashboardStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { createForm } from "svelte-forms-lib";
  import * as yup from "yup";

  export let open: boolean;
  export let metricsViewName: string;

  const queryClient = useQueryClient();

  $: dashboardStore = useDashboardStore(metricsViewName);

  $: projectId = useProjectId($page.params.organization, $page.params.project);

  const bookmarkCreator = createAdminServiceCreateBookmark();

  const formState = createForm<BookmarkFormValues>({
    initialValues: {
      displayName: "Default Name",
      description: "",
      shared: "false",
      filtersOnly: false,
      absoluteTimeRange: false,
    },
    validationSchema: yup.object({
      displayName: yup.string().required("Required"),
      description: yup.string(),
    }),
    onSubmit: async (values) => {
      await $bookmarkCreator.mutateAsync({
        data: {
          displayName: values.displayName,
          description: values.description,
          projectId: $projectId.data ?? "",
          resourceKind: ResourceKind.MetricsView,
          resourceName: metricsViewName,
          shared: values.shared === "true",
          data: getBookmarkDataForDashboard(
            $dashboardStore,
            values.filtersOnly,
            values.absoluteTimeRange,
          ),
        },
      });
      handleReset();
      queryClient.refetchQueries(
        getAdminServiceListBookmarksQueryKey({
          projectId: $projectId.data ?? "",
          resourceKind: ResourceKind.MetricsView,
          resourceName: metricsViewName,
        }),
      );
      notifications.send({
        message: "Bookmark created",
      });
      handleClose();
    },
  });

  const { handleSubmit, handleReset } = formState;

  function handleClose() {
    open = false;
  }
</script>

<Dialog on:close={handleClose} {open}>
  <svelte:fragment slot="title">Bookmark current view</svelte:fragment>
  <BaseBookmarkForm {formState} {metricsViewName} slot="body" />
  <div class="flex flex-row mt-4 gap-2" slot="footer">
    <div class="grow" />
    <Button on:click={handleClose} type="secondary">Cancel</Button>
    <Button on:click={handleSubmit} type="primary">Save</Button>
  </div>
</Dialog>
