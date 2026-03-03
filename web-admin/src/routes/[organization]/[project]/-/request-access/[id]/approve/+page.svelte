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
        message: `${$requestAccess.data.email} has been added to ${project} as a ${role}`,
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
  <h2 class="text-lg font-normal">Grant access to this project</h2>
  {#if $requestAccess.data}
    <div class="text-fg-secondary text-base">
      Select a role for <b>{$requestAccess.data.email}</b> to access the project
      <b>{project}</b>.
    </div>
    <Select
      bind:value={role}
      id="role"
      label=""
      options={[
        { value: ProjectUserRoles.Viewer, label: "Viewer" },
        { value: ProjectUserRoles.Editor, label: "Editor" },
        { value: ProjectUserRoles.Admin, label: "Admin" },
      ]}
    />
    <Button
      type="primary"
      wide
      onClick={onApprove}
      loading={$approveAccess.isPending}
      disabled={requested}
    >
      Grant access
    </Button>
  {/if}
</AccessRequestContainer>
