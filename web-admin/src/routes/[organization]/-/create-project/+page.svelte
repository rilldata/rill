<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/state";
  import { createAdminServiceListProjectsForOrganization } from "@rilldata/web-admin/client";
  import CreateProjectForm from "@rilldata/web-admin/features/projects/CreateProjectForm.svelte";
  import { getName } from "@rilldata/web-common/features/entity-management/name-utils.ts";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types.ts";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import RillLogoSquareNegative from "@rilldata/web-common/components/icons/RillLogoSquareNegative.svelte";
  import {
    type DeployError,
    isQuotaDeployError,
  } from "@rilldata/web-common/features/project/deploy/deploy-errors.ts";
  import { Button } from "@rilldata/web-common/components/button";
  import PricingDetails from "@rilldata/web-common/features/billing/PricingDetails.svelte";
  import CTAHeader from "@rilldata/web-common/components/calls-to-action/CTAHeader.svelte";
  import StartTeamPlanDialog from "@rilldata/web-admin/features/billing/plans/StartTeamPlanDialog.svelte";
  import type { TeamPlanDialogTypes } from "@rilldata/web-admin/features/billing/plans/types.ts";

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

  let deployError: DeployError | undefined = $state(undefined);
  let showStartTeamPlanDialog = $state(false);
  let startTeamPlanType: TeamPlanDialogTypes = $state("base");
  $effect(() => console.log(deployError));

  function handleCreate(frontendUrl: string) {
    setTimeout(() => void goto(`${frontendUrl}/-/welcome`));
  }
</script>

<div class="background">
  <div class="flex flex-col gap-4 mx-auto w-fit pt-48">
    {#if deployError && isQuotaDeployError(deployError)}
      <CTAHeader variant="bold">{deployError.title}</CTAHeader>
      <p class="text-base text-fg-secondary text-left w-[500px]">
        <PricingDetails extraText={deployError.message} />
      </p>
      <Button
        type="primary"
        onClick={() => {
          showStartTeamPlanDialog = true;
        }}
      >
        Upgrade
      </Button>
      <Button type="secondary" noStroke href="/{organization}">Back</Button>
    {:else}
      <RillLogoSquareNegative size="36px" />
      <div class="text-2xl font-extrabold text-fg-accent text-center">
        Create {hasProjects ? "your first" : "a new"} project
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
          <div class="h-36 w-[500px]">
            <Spinner status={EntityStatus.Running} size="2rem" duration={725} />
          </div>
        {:else}
          <CreateProjectForm
            {organization}
            defaultName={defaultProjectName}
            onCreate={handleCreate}
            onDeployError={(e) => (deployError = e)}
          />
        {/if}
      </div>
    {/if}
  </div>
</div>

<StartTeamPlanDialog
  bind:open={showStartTeamPlanDialog}
  type={startTeamPlanType}
  {organization}
/>

<style lang="postcss">
  .background {
    @apply flex flex-col w-full h-fit min-h-screen bg-no-repeat bg-cover;
    background-image: url("/img/welcome-bg-art.jpg");
  }

  :global(.dark) .background {
    background-image: url("/img/welcome-bg-art-dark.jpg");
  }
</style>
