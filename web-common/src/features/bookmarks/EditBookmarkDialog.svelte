<script lang="ts">
  import { page } from "$app/stores";
  import Dialog from "@rilldata/web-common/components/dialog/Dialog.svelte";
  import {
    createAdminServiceUpdateBookmark,
    getAdminServiceListBookmarksQueryKey,
    type V1Bookmark,
  } from "@rilldata/web-admin/client";
  import { Button, Switch } from "@rilldata/web-common/components/button";
  import InputV2 from "@rilldata/web-common/components/forms/InputV2.svelte";
  import { getBookmarkDataForDashboard } from "@rilldata/web-common/features/bookmarks/getBookmarkDataForDashboard";
  import { useProjectId } from "@rilldata/web-common/features/bookmarks/selectors";
  import { useDashboardStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { createForm } from "svelte-forms-lib";
  import * as yup from "yup";

  export let open: boolean;
  export let metricsViewName: string;
  export let bookmark: V1Bookmark;

  const queryClient = useQueryClient();

  $: dashboardStore = useDashboardStore(metricsViewName);

  $: projectId = useProjectId($page.params.organization, $page.params.project);

  const bookmarkUpdater = createAdminServiceUpdateBookmark();

  const formState = createForm({
    initialValues: {
      displayName: bookmark.displayName,
      description: (bookmark.description as string) ?? "",
      filtersOnly: false, // TODO
      absoluteTimeRange: false, // TODO
    },
    validationSchema: yup.object({
      displayName: yup.string().required("Required"),
      description: yup.string(),
    }),
    onSubmit: async (values) => {
      await $bookmarkUpdater.mutateAsync({
        data: {
          bookmarkId: bookmark.id,
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
    on:submit|preventDefault={handleSubmit}
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
    <Switch
      checked={$form["filtersOnly"]}
      id="filtersOnly"
      on:click={() => ($form["filtersOnly"] = !$form["filtersOnly"])}
    >
      Save filters only
    </Switch>
    <Switch
      bind:checked={$form["absoluteTimeRange"]}
      id="absoluteTimeRange"
      on:click={() =>
        ($form["absoluteTimeRange"] = !$form["absoluteTimeRange"])}
    >
      Absolute time range (TODO range)
    </Switch>
  </form>
  <div class="flex flex-row mt-4 gap-2" slot="footer">
    <Button on:click={handleClose} type="secondary">Cancel</Button>
    <Button on:click={handleSubmit} type="primary">Save</Button>
  </div>
</Dialog>
