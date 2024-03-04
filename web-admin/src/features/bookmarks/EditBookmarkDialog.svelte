<script lang="ts">
  import { page } from "$app/stores";
  import Dialog from "@rilldata/web-common/components/dialog/Dialog.svelte";
  import {
    createAdminServiceUpdateBookmark,
    getAdminServiceListBookmarksQueryKey,
  } from "@rilldata/web-admin/client";
  import { Button } from "@rilldata/web-common/components/button";
  import InputV2 from "@rilldata/web-common/components/forms/InputV2.svelte";
  import Label from "@rilldata/web-common/components/forms/Label.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import BookmarkTimeRangeSwitch from "@rilldata/web-admin/features/bookmarks/BookmarkTimeRangeSwitch.svelte";
  import { getBookmarkDataForDashboard } from "@rilldata/web-admin/features/bookmarks/getBookmarkDataForDashboard";
  import {
    type BookmarkEntry,
    useProjectId,
  } from "@rilldata/web-admin/features/bookmarks/selectors";
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
    initialValues: {
      displayName: bookmark.resource.displayName ?? "",
      description: (bookmark.resource.description as string) ?? "",
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
      handleClose();
    },
  });

  const { form, errors, handleSubmit } = formState;

  function handleClose() {
    open = false;
  }
</script>

<Dialog on:close={handleClose} {open} widthOverride="w-[602px]">
  <svelte:fragment slot="title">Bookmark current view</svelte:fragment>
  <form
    class="flex flex-col gap-4 z-50"
    id="create-bookmark-dialog"
    on:submit|preventDefault={() => {
      /* Switch was triggering this causing clicking on them submitting the form */
    }}
    slot="body"
  >
    <InputV2
      bind:value={$form["displayName"]}
      error={$errors["displayName"]}
      id="displayName"
      label="Name"
    />
    <InputV2
      bind:value={$form["description"]}
      error={$errors["description"]}
      id="description"
      label="Description"
      optional
    />
    <div class="flex items-center space-x-2">
      <Switch bind:checked={$form["filtersOnly"]} id="filtersOnly" />
      <Label for="filtersOnly">Save filters only</Label>
    </div>
    <BookmarkTimeRangeSwitch
      bind:checked={$form["absoluteTimeRange"]}
      {metricsViewName}
    />
  </form>
  <div class="flex flex-row mt-4 gap-2" slot="footer">
    <Button on:click={handleClose} type="secondary">Cancel</Button>
    <Button on:click={handleSubmit} type="primary">Save</Button>
  </div>
</Dialog>
