<script lang="ts">
  import * as Dropdown from "@rilldata/web-common/components/dropdown-menu";
  import { determineDropdownAlign } from "@rilldata/web-admin/features/organizations/user-management/table/dropdownAlignment.ts";
  import {
    createAdminServiceListProjectsForOrganizationAndUser,
    type V1Project,
  } from "@rilldata/web-admin/client";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import CaretUpIcon from "@rilldata/web-common/components/icons/CaretUpIcon.svelte";
  import { browser } from "$app/environment";
  import { onDestroy, onMount, tick } from "svelte";

  export let organization: string;
  export let userId: string;
  export let projectCount: number;

  let isDropdownOpen = false;
  $: hasUserId = !!userId;

  $: userProjectsQuery = createAdminServiceListProjectsForOrganizationAndUser(
    organization,
    { userId },
    {
      query: {
        enabled: !!userId && isDropdownOpen,
      },
    },
  );
  $: ({ data, isPending, error } = $userProjectsQuery);
  let projects: V1Project[];
  $: projects = data?.projects ?? [];

  $: hasProjects = projectCount > 0;
  let dropdownAlign: "start" | "end" = "start";
  let dropdownTriggerEl: HTMLElement | null = null;
  let dropdownContentEl: HTMLElement | null = null;

  async function updateDropdownAlignment() {
    if (!browser || !isDropdownOpen || !dropdownTriggerEl) return;
    await tick();

    const menuWidth =
      dropdownContentEl?.offsetWidth ??
      dropdownTriggerEl?.offsetWidth ??
      200;

    dropdownAlign = determineDropdownAlign({
      triggerRect: dropdownTriggerEl.getBoundingClientRect(),
      menuWidth,
      viewportWidth: window.innerWidth,
    });
  }

  function handleWindowResize() {
    void updateDropdownAlignment();
  }

  onMount(() => {
    if (!browser) return;
    window.addEventListener("resize", handleWindowResize);
  });

  onDestroy(() => {
    if (!browser) return;
    window.removeEventListener("resize", handleWindowResize);
  });

  $: if (isDropdownOpen) {
    void updateDropdownAlignment();
  }

  $: if (isDropdownOpen && dropdownContentEl) {
    void updateDropdownAlignment();
  }

  $: if (isDropdownOpen && projects?.length !== undefined) {
    void updateDropdownAlignment();
  }

  function getProjectShareUrl(projectName: string) {
    // Link the user to the project dashboard list and open the share popover immediately.
    return `/${organization}/${projectName}/-/dashboards?share=true`;
  }
</script>

{#if hasUserId && hasProjects}
  <Dropdown.Root bind:open={isDropdownOpen}>
    <Dropdown.Trigger
      bind:this={dropdownTriggerEl}
      class="w-18 flex flex-row gap-1 items-center rounded-sm {isDropdownOpen
        ? 'bg-slate-200'
        : 'hover:bg-slate-100'} px-2 py-1"
    >
      <span class="capitalize">
        {projectCount} Project{projectCount !== 1 ? "s" : ""}
      </span>
      {#if isDropdownOpen}
        <CaretUpIcon size="12px" />
      {:else}
        <CaretDownIcon size="12px" />
      {/if}
    </Dropdown.Trigger>
    <Dropdown.Content bind:this={dropdownContentEl} align={dropdownAlign}>
      {#if isPending}
        Loading...
      {:else if error}
        Error
      {:else}
        {#each projects as project (project.id)}
          <Dropdown.Item href={getProjectShareUrl(project.name)}>
            {project.name}
          </Dropdown.Item>
        {/each}
      {/if}
    </Dropdown.Content>
  </Dropdown.Root>
{:else}
  <div class="w-18 rounded-sm px-2 py-1 text-gray-400">No projects</div>
{/if}
