<script lang="ts">
  import Discord from "@rilldata/web-common/components/icons/Discord.svelte";
  import Docs from "@rilldata/web-common/components/icons/Docs.svelte";
  import Github from "@rilldata/web-common/components/icons/Github.svelte";
  import InfoCircle from "@rilldata/web-common/components/icons/InfoCircle.svelte";
  import Shortcut from "@rilldata/web-common/components/tooltip/Shortcut.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import TooltipShortcutContainer from "@rilldata/web-common/components/tooltip/TooltipShortcutContainer.svelte";
  import TooltipTitle from "@rilldata/web-common/components/tooltip/TooltipTitle.svelte";
  import type { ApplicationBuildMetadata } from "@rilldata/web-local/lib/application-state-stores/build-metadata";
  import { getContext } from "svelte";
  import type { Writable } from "svelte/store";
  import { fly } from "svelte/transition";

  const appBuildMetaStore: Writable<ApplicationBuildMetadata> =
    getContext("rill:app:metadata");

  const lineItems = [
    {
      icon: Docs,
      label: "Documentation",
      href: "https://docs.rilldata.com",
      className: "fill-gray-600",
      shrinkIcon: false,
    },
    {
      icon: Discord,
      label: "Ask a question",
      href: "http://bit.ly/3jg4IsF",
      className: "fill-gray-500",
      shrinkIcon: true,
    },
    {
      icon: Github,
      label: "Report an issue",
      href: "https://github.com/rilldata/rill-developer/issues/new?assignees=&labels=bug&template=bug_report.md&title=",
      className: "fill-gray-500",
      shrinkIcon: true,
    },
  ];
</script>

<div
  class="flex flex-col  pt-3 pb-3 gap-y-1 bg-gray-50 border-t border-gray-200 sticky bottom-0"
>
  {#each lineItems as lineItem}
    <a href={lineItem.href} target="_blank"
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
    class="px-4 py-1 text-gray-600 flex flex-row  gap-x-2"
    style:font-size="10px"
  >
    <span class="text-gray-400">
      <Tooltip alignment="start" distance={16} location="top">
        <a
          href="https://docs.rilldata.com"
          target="_blank"
          class="text-gray-400"
        >
          <InfoCircle size="16px" />
        </a>
        <div slot="tooltip-content" transition:fly={{ duration: 100, y: 8 }}>
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
    version {$appBuildMetaStore.version
      ? $appBuildMetaStore.version
      : "unknown (built from source)"}{$appBuildMetaStore.commitHash
      ? ` – ${$appBuildMetaStore.commitHash}`
      : ""}
  </div>
</div>
