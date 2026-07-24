<script lang="ts">
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import { removeLeadingSlash } from "../entity-management/entity-mappers";
  import { getFileHref } from "../../layout/navigation/editor-routing";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import {
    displayResourceKind,
    ResourceKind,
  } from "../entity-management/resource-selectors";
  // svelte-ignore custom_element_props_identifier
  let {
    resources,
    hrefForResource = undefined,
    ...triggerProps
  }: {
    resources: V1Resource[];
    hrefForResource?: (resource: V1Resource) => string;
    [key: string]: unknown;
  } = $props();

  let firstResource = $derived(resources?.[0]);
  // Spell out "canvas dashboard" to distinguish the target from the metrics
  // view's own explore dashboard (displayResourceKind flattens both to "dashboard").
  let firstResourceType = $derived(
    firstResource?.meta?.name?.kind === ResourceKind.Canvas
      ? "canvas dashboard"
      : displayResourceKind(
          firstResource?.meta?.name?.kind as ResourceKind | undefined,
        ),
  );
</script>

{#if resources?.length === 1 && firstResource.meta?.filePaths?.[0]}
  <div
    class="border-accent-primary-action flex items-center border h-7 rounded-[2px] bg-transparent text-accent-primary-action"
  >
    <a
      href={hrefForResource?.(firstResource) ??
        getFileHref(
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
