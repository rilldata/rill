<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/state";
  import { createAdminServiceListProjectsForOrganization } from "@rilldata/web-admin/client";
  import CreateProjectForm from "@rilldata/web-admin/features/projects/CreateProjectForm.svelte";
  import CtaHeader from "@rilldata/web-common/components/calls-to-action/CTAHeader.svelte";
  import { getName } from "@rilldata/web-common/features/entity-management/name-utils.ts";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types.ts";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import RillLogoSquareNegative from "@rilldata/web-common/components/icons/RillLogoSquareNegative.svelte";

  let organization = $derived(page.params.organization);

  let projectsQuery = $derived(
    createAdminServiceListProjectsForOrganization(organization, undefined),
  );
  let hasProjects = $derived(projectsQuery.data?.projects?.length > 0);

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

<div class="background">
  <div class="flex flex-col gap-4 mx-auto w-fit pt-48">
    <RillLogoSquareNegative size="36px" />
    <div class="text-2xl font-extrabold text-fg-accent text-center">
      Create your {hasProjects ? "first " : ""}project
    </div>

    <div
      class="flex flex-col gap-6 text-left p-6 border rounded-md bg-surface-overlay"
    >
      <div>
        <div class="text-base font-semibold">
          Name your {hasProjects ? "first " : ""}project
        </div>
        <div class="text-sm text-fg-muted">
          You can rename anytime from project settings.
        </div>
      </div>

      {#if $projectsQuery.isPending}
        <div class="h-36">
          <Spinner status={EntityStatus.Running} size="7rem" duration={725} />
        </div>
      {:else}
        <CreateProjectForm
          {organization}
          defaultName={defaultProjectName}
          onCreate={handleCreate}
        />
      {/if}
    </div>
  </div>
</div>

<style lang="postcss">
  .background {
    @apply flex flex-col w-full h-fit min-h-screen bg-no-repeat bg-cover;
    background-image: url("/img/welcome-bg-art.jpg");
  }

  :global(.dark) .background {
    background-image: url("/img/welcome-bg-art-dark.jpg");
  }
</style>
