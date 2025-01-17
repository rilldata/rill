<script lang="ts">
  import Github from "@rilldata/web-common/components/icons/Github.svelte";
  import InfoCircle from "@rilldata/web-common/components/icons/InfoCircle.svelte";
  import Shortcut from "@rilldata/web-common/components/tooltip/Shortcut.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import TooltipShortcutContainer from "@rilldata/web-common/components/tooltip/TooltipShortcutContainer.svelte";
  import TooltipTitle from "@rilldata/web-common/components/tooltip/TooltipTitle.svelte";
  import type { ApplicationBuildMetadata } from "@rilldata/web-common/layout/build-metadata";
  import { getContext } from "svelte";
  import type { Writable } from "svelte/store";
  import { fly } from "svelte/transition";

  const appBuildMetaStore: Writable<ApplicationBuildMetadata> =
    getContext("rill:app:metadata");

  const lineItems = [
    {
      icon: Github,
      label: "Report an issue",
      href: "https://github.com/rilldata/rill/issues/new?assignees=&labels=bug&template=bug_report.md&title=",
      className: "fill-gray-800",
      shrinkIcon: true,
    },
  ];
</script>

<div
  class="flex flex-col pt-3 pb-3 gap-y-1 bg-gray-50 border-t border-gray-200 sticky bottom-0"
>
  {#each lineItems as lineItem, i (i)}
    <a href={lineItem.href} target="_blank" rel="noreferrer noopener"
      ><div
        class="flex flex-row items-center px-4 py-1 gap-x-2 text-gray-700 font-normal hover:bg-gray-200"
      >
        <!-- workaround to resize the github and discord icons to match -->
        <div
          class="grid place-content-center"
          style:width="16px"
          style:height="16px"
        >
          <svelte:component
            this={lineItem.icon}
            className={lineItem.className}
            size="14px"
          />
        </div>
        {lineItem.label}
      </div></a
    >
  {/each}
  <div
    class="px-4 py-1 text-gray-600 flex flex-row w-full gap-x-2 truncate line-clamp-1"
    style:font-size="10px"
  >
    <span class="text-gray-400">
      <Tooltip alignment="start" distance={16} location="top">
        <a
          href="https://docs.rilldata.com"
          target="_blank"
          rel="noreferrer noopener"
          class="text-gray-400"
        >
          <InfoCircle size="16px" />
        </a>
        <div
          slot="tooltip-content"
          transition:fly|global={{ duration: 100, y: 8 }}
        >
          <TooltipContent>
            <TooltipTitle>
              <svelte:fragment slot="name">Rill Developer</svelte:fragment>
            </TooltipTitle>
            <TooltipShortcutContainer>
              <div>View documentation</div>
              <Shortcut>Click</Shortcut>
            </TooltipShortcutContainer>
          </TooltipContent>
        </div>
      </Tooltip>
    </span>
    <span class="truncate">
      version {$appBuildMetaStore.version
        ? $appBuildMetaStore.version
        : "unknown (built from source)"}{$appBuildMetaStore.commitHash
        ? ` – ${$appBuildMetaStore.commitHash}`
        : ""}
    </span>
  </div>
</div>
