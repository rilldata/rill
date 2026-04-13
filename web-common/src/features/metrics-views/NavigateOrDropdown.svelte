<script lang="ts">
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import { removeLeadingSlash } from "../entity-management/entity-mappers";
  import { getFileHref } from "../workspaces/edit-routing";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import {
    displayResourceKind,
    ResourceKind,
  } from "../entity-management/resource-selectors";
  // svelte-ignore custom_element_props_identifier
  let {
    resources,
    ...triggerProps
  }: {
    resources: V1Resource[];
    [key: string]: unknown;
  } = $props();

  let firstResource = $derived(resources?.[0]);
  let firstResourceType = $derived(
    displayResourceKind(
      firstResource?.meta?.name?.kind as ResourceKind | undefined,
    ),
  );
</script>

{#if resources?.length === 1 && firstResource.meta?.filePaths?.[0]}
  <div
    class="border-accent-primary-action flex items-center border h-7 rounded-[2px] bg-transparent text-accent-primary-action"
  >
    <a
      href={getFileHref(
        `/${removeLeadingSlash(firstResource.meta?.filePaths?.[0])}`,
      )}
      class="text-inherit font-medium flex items-center border-r px-3 size-full hover:bg-surface-hover border-accent-primary-action hover:text-fg-accent"
    >
      Go to {firstResourceType}
    </a>
    <button
      {...triggerProps}
      aria-label="Create resource menu"
      class="text-inherit h-full aspect-square grid place-content-center hover:bg-surface-hover hover:text-fg-accent"
    >
      <CaretDownIcon />
    </button>
  </div>
{:else}
  <Button {...triggerProps} type="secondary">
    Go to {firstResourceType}
    <CaretDownIcon />
  </Button>
{/if}
