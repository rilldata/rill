<script lang="ts">
  import { page } from "$app/stores";
  import Dialog from "@rilldata/web-common/components/dialog/Dialog.svelte";
  import {
    createAdminServiceCreateBookmark,
    getAdminServiceListBookmarksQueryKey,
  } from "@rilldata/web-admin/client";
  import { Button, Switch } from "@rilldata/web-common/components/button";
  import InputV2 from "@rilldata/web-common/components/forms/InputV2.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import { getBookmarkForDashboard } from "@rilldata/web-common/features/bookmarks/bookmark-dashboard-state";
  import { useProjectId } from "@rilldata/web-common/features/bookmarks/selectors";
  import { useDashboardStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { createForm } from "svelte-forms-lib";
  import * as yup from "yup";

  export let open: boolean;
  export let metricsViewName: string;

  const queryClient = useQueryClient();

  $: dashboardStore = useDashboardStore(metricsViewName);

  $: projectId = useProjectId($page.params.organization, $page.params.project);

  const bookmarkCreator = createAdminServiceCreateBookmark();

  const formState = createForm({
    initialValues: {
      displayName: "Default Name",
      description: "",
      isGlobal: false,
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
          dashboardName: metricsViewName,
          isGlobal: values.isGlobal,
          data: getBookmarkForDashboard(
            $dashboardStore,
            values.filtersOnly,
            values.absoluteTimeRange,
          ),
        },
      });
      queryClient.refetchQueries(
        getAdminServiceListBookmarksQueryKey({
          projectId: $projectId.data ?? "",
          dashboardName: metricsViewName,
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
    class="flex flex-col gap-4"
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
    <Select
      bind:value={$form["isGlobal"]}
      id="isGlobal"
      label="Category"
      options={[
        { value: false, label: "Your bookmarks" },
        { value: true, label: "Default bookmarks" },
      ]}
    />
    <Switch bind:checked={$form["filtersOnly"]} id="filtersOnly">
      Save filters only
    </Switch>
    <Switch bind:checked={$form["absoluteTimeRange"]} id="absoluteTimeRange">
      Absolute time range (TODO range)
    </Switch>
  </form>
  <div class="flex flex-row mt-4 gap-2" slot="footer">
    <Button on:click={handleClose} type="secondary">Cancel</Button>
    <Button on:click={handleSubmit} type="primary">Save</Button>
  </div>
</Dialog>
