<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import {
    createAdminServiceUpdateBookmark,
    type V1Bookmark,
  } from "@rilldata/web-admin/client";
  import { getRpcErrorMessage } from "@rilldata/web-admin/components/errors/error-utils.ts";
  import { invalidateBookmarkQueries } from "@rilldata/web-admin/features/bookmarks/selectors.ts";
  import ProjectAccessControls from "@rilldata/web-admin/features/projects/ProjectAccessControls.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";
  import { defaults, superForm } from "sveltekit-superforms";
  import { yup } from "sveltekit-superforms/adapters";
  import { object, string } from "yup";

  // Edits a bookmark's name, description and category without touching its saved dashboard state.
  // Changing the saved state requires the dashboard itself (see BookmarksFormDialog).
  export let organization: string;
  export let project: string;
  export let bookmark: V1Bookmark;
  // Link to the bookmark on its dashboard, where its filters can be changed.
  export let href: string | undefined = undefined;
  export let onClose = () => {};

  const bookmarkUpdater = createAdminServiceUpdateBookmark();

  const initialValues = {
    displayName: bookmark.displayName ?? "",
    description: bookmark.description ?? "",
    shared: bookmark.shared ? "true" : "false",
  };

  const schema = yup(
    object({
      displayName: string().required("Required"),
      description: string(),
      shared: string(),
    }),
  );

  const formId = "edit-bookmark-metadata-dialog";

  const { form, errors, submit, enhance } = superForm(
    defaults(initialValues, schema),
    {
      SPA: true,
      validators: schema,
      async onUpdate({ form }) {
        if (!form.valid) return;
        const values = form.data;

        await $bookmarkUpdater.mutateAsync({
          data: {
            bookmarkId: bookmark.id,
            displayName: values.displayName,
            description: values.description,
            // Home bookmarks are always shared.
            shared: bookmark.default || values.shared === "true",
            urlSearch: bookmark.urlSearch ?? "",
          },
        });
        onClose();

        await invalidateBookmarkQueries();
        eventBus.emit("notification", {
          message: m.bookmark_updated(),
        });
      },
    },
  );

  $: error = getRpcErrorMessage($bookmarkUpdater.error ?? undefined);
</script>

<Dialog.Root
  open
  onOpenChange={(o) => {
    if (!o) onClose();
  }}
>
  <Dialog.Content>
    <Dialog.Header>
      <Dialog.Title>{m.bookmark_edit()}</Dialog.Title>
    </Dialog.Header>

    <form
      class="flex flex-col gap-4"
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
      {#if !bookmark.default}
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
      {/if}
      {#if href}
        <div class="text-sm text-fg-secondary">
          {m.bookmark_filters()}:
          <a {href}>{m.bookmark_edit_filters_on_dashboard()}</a>
        </div>
      {/if}
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
