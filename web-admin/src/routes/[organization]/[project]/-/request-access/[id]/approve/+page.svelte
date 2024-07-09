<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import {
    createAdminServiceApproveProjectAccess,
    createAdminServiceGetProjectAccess,
    type RpcStatus,
  } from "@rilldata/web-admin/client";
  import { parseAccessRequestError } from "@rilldata/web-admin/features/access-request/utils";
  import { Button } from "@rilldata/web-common/components/button";
  import AccessRequestContainer from "@rilldata/web-admin/features/access-request/AccessRequestContainer.svelte";
  import CheckCircle from "@rilldata/web-common/components/icons/CheckCircle.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import type { AxiosError } from "axios";

  $: organization = $page.params.organization;
  $: project = $page.params.project;
  $: id = $page.params.id;

  let requested = false;
  let role = "viewer";
  $: approveAccess = createAdminServiceApproveProjectAccess();
  function onApprove() {
    requested = true;
    void $approveAccess.mutateAsync({
      organization,
      project,
      id,
      data: {
        role,
      },
    });
    goto(`/${organization}/${project}`);
  }

  $: error = parseAccessRequestError(
    $approveAccess.error as unknown as AxiosError<RpcStatus>,
  );

  $: requestAccess = createAdminServiceGetProjectAccess(
    organization,
    project,
    id,
  );
</script>

<AccessRequestContainer>
  <CheckCircle size="40px" className="text-primary-500" />
  <h2 class="text-lg font-normal">Grant access to this project</h2>
  {#if $requestAccess.data}
    <div class="text-slate-500 text-base">
      Select a role for <b>{$requestAccess.data.email}</b> to access the project
      <b>{project}</b>.
    </div>
    <Select
      bind:value={role}
      id="role"
      label=""
      options={[
        { value: "viewer", label: "Viewer" },
        { value: "admin", label: "Admin" },
      ]}
    />
    <Button
      type="primary"
      wide
      on:click={onApprove}
      loading={$approveAccess.isLoading}
      disabled={requested}
    >
      Grant access
    </Button>
  {/if}
  {#if error}
    <div class="text-slate-500 text-base">
      {error}
    </div>
  {/if}
</AccessRequestContainer>
