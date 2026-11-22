<script lang="ts">
  import {
    createAdminServiceGetProject,
    createAdminServiceUpdateProject,
    getAdminServiceGetProjectQueryKey,
    getAdminServiceListProjectsForOrganizationQueryKey,
    type RpcStatus,
  } from "@rilldata/web-admin/client";
  import SettingsContainer from "@rilldata/web-admin/features/organizations/settings/SettingsContainer.svelte";
  import { listProjectsForOrgQueryOptions } from "@rilldata/web-admin/features/projects/list-projects-query-options";
  import {
    diffProjectTags,
    getProjectTags,
    isValidProjectTag,
    MAX_PROJECT_TAGS,
    sanitizeProjectTag,
  } from "@rilldata/web-admin/features/projects/project-tags";
  import { Button } from "@rilldata/web-common/components/button";
  import TagInput from "@rilldata/web-common/components/forms/TagInput.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { createQuery } from "@tanstack/svelte-query";
  import type { AxiosError } from "axios";

  let { organization, project }: { organization: string; project: string } =
    $props();

  const updateProjectMutation = createAdminServiceUpdateProject();

  let projectResp = $derived(
    createAdminServiceGetProject(organization, project),
  );
  let savedTags = $derived($projectResp.data?.project?.tags ?? []);

  // Suggest the tags of the org's other projects, so the org converges on a shared set of tags.
  let projectsQuery = $derived(
    createQuery(listProjectsForOrgQueryOptions(organization)),
  );
  let suggestions = $derived(
    getProjectTags($projectsQuery.data?.projects ?? []),
  );

  // Unsaved edits; while there are none, the input shows the saved tags.
  let editedTags = $state<string[] | undefined>(undefined);
  let tags = $derived(editedTags ?? savedTags);
  let changes = $derived(diffProjectTags(savedTags, tags));
  let changed = $derived(
    changes.addTags.length > 0 || changes.removeTags.length > 0,
  );
  let error = $state("");

  function onChange(newTags: string[]) {
    const sanitized = newTags.map(sanitizeProjectTag);
    const invalid = sanitized.findIndex((t) => !isValidProjectTag(t));
    if (invalid !== -1) {
      error = m.settings_project_tags_invalid({ tag: newTags[invalid].trim() });
      return;
    }
    const unique = [...new Set(sanitized)];
    if (unique.length > MAX_PROJECT_TAGS) {
      error = m.settings_project_tags_too_many({ max: MAX_PROJECT_TAGS });
      return;
    }
    error = "";
    editedTags = unique;
  }

  async function save() {
    try {
      await $updateProjectMutation.mutateAsync({
        org: organization,
        project,
        data: changes,
      });
    } catch (err) {
      const axiosError = err as AxiosError<RpcStatus>;
      error =
        axiosError.response?.data?.message ??
        m.settings_project_tags_update_failed();
      return;
    }

    await Promise.all([
      queryClient.refetchQueries({
        queryKey: getAdminServiceGetProjectQueryKey(organization, project),
      }),
      queryClient.invalidateQueries({
        queryKey:
          getAdminServiceListProjectsForOrganizationQueryKey(organization),
      }),
    ]);
    editedTags = undefined;
    error = "";
    eventBus.emit("notification", {
      message: m.settings_project_tags_updated_notification(),
    });
  }
</script>

<SettingsContainer title={m.settings_project_tags_title()}>
  <div class="flex flex-col gap-y-2 max-w-[520px]">
    <TagInput
      {tags}
      {suggestions}
      label={m.settings_project_tags_label()}
      placeholder={m.settings_project_tags_placeholder()}
      {error}
      {onChange}
    />
    <p>{m.settings_project_tags_description()}</p>
  </div>
  {#snippet action()}
    <Button
      type="primary"
      onClick={save}
      loading={$updateProjectMutation.isPending}
      disabled={!changed}
    >
      {m.settings_save_button()}
    </Button>
  {/snippet}
</SettingsContainer>
