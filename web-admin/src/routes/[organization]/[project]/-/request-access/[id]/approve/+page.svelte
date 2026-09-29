<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import {
    createAdminServiceApproveProjectAccess,
    createAdminServiceGetProjectAccessRequest,
  } from "@rilldata/web-admin/client";
  import AccessRequestContainer from "@rilldata/web-admin/features/access-request/AccessRequestContainer.svelte";
  import { parseAccessRequestError } from "@rilldata/web-admin/features/access-request/utils";
  import { Button } from "@rilldata/web-common/components/button";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import CheckCircle from "@rilldata/web-common/components/icons/CheckCircle.svelte";
  import { ProjectUserRoles } from "@rilldata/web-common/features/users/roles.ts";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { escapeHtml } from "@rilldata/web-common/lib/i18n";
  import type { AxiosError } from "axios";

  $: organization = $page.params.organization;
  $: project = $page.params.project;
  $: id = $page.params.id;
  let role = $page.url.searchParams.get("role") ?? ProjectUserRoles.Viewer;

  let requested = false;
  $: approveAccess = createAdminServiceApproveProjectAccess();
  $: requestAccess = createAdminServiceGetProjectAccessRequest(id);

  async function onApprove() {
    requested = true;
    try {
      await $approveAccess.mutateAsync({
        id,
        data: {
          role,
        },
      });
      eventBus.emit("notification", {
        type: "success",
        message: m.auth_user_added_to_project({
          email: $requestAccess.data.email,
          project,
          role,
        }),
        options: {
          persisted: true,
        },
      });
    } catch {
      eventBus.emit("notification", {
        type: "error",
        message: parseAccessRequestError(
          project,
          $requestAccess.error as unknown as AxiosError,
        ),
        options: {
          persisted: true,
        },
      });
    }
    return goto(`/${organization}/${project}`);
  }

  $: if ($requestAccess.error) {
    eventBus.emit("notification", {
      type: "error",
      message: parseAccessRequestError(
        project,
        $requestAccess.error as unknown as AxiosError,
      ),
      options: {
        persisted: true,
      },
    });
    goto(`/${organization}/${project}`);
  }
</script>

<AccessRequestContainer>
  <CheckCircle size="40px" className="text-primary-500" />
  <h2 class="text-lg font-normal">{m.auth_grant_access_title()}</h2>
  {#if $requestAccess.data}
    <div class="text-fg-secondary text-base">
      {@html m.auth_grant_access_description({
        email: `<b>${escapeHtml($requestAccess.data.email)}</b>`,
        project: `<b>${escapeHtml(project)}</b>`,
      })}
    </div>
    <Select
      bind:value={role}
      id="role"
      label=""
      options={[
        { value: ProjectUserRoles.Viewer, label: m.role_viewer() },
        { value: ProjectUserRoles.Editor, label: m.role_editor() },
        { value: ProjectUserRoles.Admin, label: m.role_admin() },
      ]}
    />
    <Button
      type="primary"
      wide
      onClick={onApprove}
      loading={$approveAccess.isPending}
      disabled={requested}
    >
      {m.auth_grant_access()}
    </Button>
  {/if}
</AccessRequestContainer>
