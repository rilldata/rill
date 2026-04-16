<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/state";
  import { createAdminServiceListProjectsForOrganization } from "@rilldata/web-admin/client";
  import CreateProjectForm from "@rilldata/web-admin/features/projects/CreateProjectForm.svelte";
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaHeader from "@rilldata/web-common/components/calls-to-action/CTAHeader.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import { getName } from "@rilldata/web-common/features/entity-management/name-utils.ts";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types.ts";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";

  let organization = $derived(page.params.organization);

  let projectsQuery = $derived(
    createAdminServiceListProjectsForOrganization(organization, undefined),
  );

  let defaultProjectName = $derived(
    getName(
      "new_project",
      $projectsQuery.data?.projects?.map((p) => p.name) ?? [],
    ),
  );

  function handleCreate(frontendUrl: string) {
    setTimeout(() => void goto(`${frontendUrl}/-/welcome`));
  }
</script>

<CtaLayoutContainer>
  <CtaContentContainer>
    {#if $projectsQuery.isPending}
      <div class="h-36">
        <Spinner status={EntityStatus.Running} size="7rem" duration={725} />
      </div>
    {:else}
      <CtaHeader variant="bold">Name your project</CtaHeader>
      <CreateProjectForm
        {organization}
        defaultName={defaultProjectName}
        onCreate={handleCreate}
      />
    {/if}
  </CtaContentContainer>
</CtaLayoutContainer>
