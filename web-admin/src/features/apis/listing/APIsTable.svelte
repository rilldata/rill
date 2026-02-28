<script lang="ts">
  import ResourceList from "@rilldata/web-admin/features/resources/ResourceList.svelte";
  import ResourceListEmptyState from "@rilldata/web-admin/features/resources/ResourceListEmptyState.svelte";
  import APIIcon from "@rilldata/web-common/components/icons/APIIcon.svelte";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import { flexRender, type ColumnDef } from "@tanstack/svelte-table";
  import APIsTableCompositeCell from "./APIsTableCompositeCell.svelte";

  export let data: V1Resource[];
  export let organization: string;
  export let project: string;

  const columns: ColumnDef<V1Resource, string>[] = [
    {
      id: "composite",
      cell: (info) =>
        flexRender(APIsTableCompositeCell, {
          id: info.row.original.meta.name.name,
          title:
            info.row.original.api?.spec?.displayName ||
            info.row.original.meta.name.name,
          description: info.row.original.api?.spec?.description,
          resolver: info.row.original.api?.spec?.resolver,
          openapiSummary: info.row.original.api?.spec?.openapiSummary,
          securityRules: info.row.original.api?.spec?.securityRules ?? [],
          reconcileError: info.row.original.meta?.reconcileError,
        }),
    },
    {
      id: "name",
      accessorFn: (row) => row.meta.name.name,
    },
  ];

  const columnVisibility = {
    name: false,
  };
</script>

<!-- TODO: test with live API resources to verify data mapping -->
<ResourceList {columns} {data} {columnVisibility} kind="API">
  <ResourceListEmptyState
    slot="empty"
    icon={APIIcon}
    message="You don't have any APIs yet"
  >
    <span slot="action">
      Create <a
        href="https://docs.rilldata.com/reference/project-files/apis"
        target="_blank"
        rel="noopener noreferrer"
      >
        APIs
      </a>
      via code.
    </span>
  </ResourceListEmptyState>
</ResourceList>
