<script lang="ts">
  import { onMount, setContext } from "svelte";
  import { createDummyProject, projectManager } from "./project-manager";
  import { fetchProject } from "@rilldata/web-admin/features/projects/selectors";

  export let organization: string | undefined = undefined;
  export let project: string | undefined = undefined;
  export let token: string | undefined = undefined;

  onMount(async () => {
    if (!organization || !project) return;

    if (organization === "default" && project === "default") {
      projectManager.addProject(createDummyProject());
      return;
    }

    const response = await fetchProject(organization, project, token);

    if (!response) return;

    projectManager.addProject(response);
  });

  setContext("organization", organization);
  setContext("project", project);
</script>

<slot />
