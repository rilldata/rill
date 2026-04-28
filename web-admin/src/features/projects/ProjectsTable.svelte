<script lang="ts">
  import type { V1Project } from "@rilldata/web-admin/client";
  import ResourceList from "@rilldata/web-admin/features/resources/ResourceList.svelte";
  import ResourceListEmptyState from "@rilldata/web-admin/features/resources/ResourceListEmptyState.svelte";
  import CloudIcon from "@rilldata/web-common/components/icons/CloudIcon.svelte";
  import { renderComponent } from "tanstack-table-8-svelte-5";
  import ProjectsTableCompositeCell from "./ProjectsTableCompositeCell.svelte";

  export let organization: string;
  export let projects: V1Project[];

  const columns = [
    {
      id: "composite",
      cell: ({ row }) => {
        const project = row.original as V1Project;
        return renderComponent(ProjectsTableCompositeCell, {
          organization,
          project: project.name ?? "",
          description: project.description ?? "",
          isPublic: !!project.public,
          updatedOn: project.updatedOn,
        });
      },
    },
    {
      id: "name",
      accessorFn: (row: V1Project) => row.name,
    },
    {
      id: "description",
      accessorFn: (row: V1Project) => row.description ?? "",
    },
    {
      id: "updatedOn",
      accessorFn: (row: V1Project) => row.updatedOn,
    },
  ];

  const columnVisibility = {
    name: false,
    description: false,
    updatedOn: false,
  };

  const initialSorting = [{ id: "name", desc: false }];
</script>

<ResourceList
  kind="project"
  data={projects}
  {columns}
  {columnVisibility}
  {initialSorting}
>
  <ResourceListEmptyState
    slot="empty"
    icon={CloudIcon}
    message="You don't have any projects yet"
  />
</ResourceList>
