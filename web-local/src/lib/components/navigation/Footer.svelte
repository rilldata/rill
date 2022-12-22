<script lang="ts">
  import Discord from "@rilldata/web-common/components/icons/Discord.svelte";
  import Docs from "@rilldata/web-common/components/icons/Docs.svelte";
  import Github from "@rilldata/web-common/components/icons/Github.svelte";
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

  const actionItems = [
    {
      icon: Docs,
      tooltipTitle: "Documentation",
      tooltipCTA: "Read the docs",
      href: "https://docs.rilldata.com",
      className: "fill-gray-600",
    },
    {
      icon: Discord,
      tooltipTitle: "Discord",
      tooltipCTA: "Ask a question",
      href: "http://bit.ly/3jg4IsF",
      className: "fill-gray-500",
    },
    {
      icon: Github,
      tooltipTitle: "GitHub",
      tooltipCTA: "Report an issue",
      href: "https://github.com/rilldata/rill-developer/issues/new?assignees=&labels=bug&template=bug_report.md&title=",
      className: "fill-gray-500",
    },
  ];
</script>

<div
  class="flex flex-row flex-wrap px-3 py-3 gap-y-1 gap-x-1 bg-gray-50 border-t border-gray-200 sticky bottom-0 items-center justify-between"
>
  <div class="flex flex-row gap-x-2 text-gray-600" style:font-size="10px">
    {#each actionItems as actionItem}
      <Tooltip alignment="start" distance={16} location="top">
        <a href={actionItem.href} target="_blank">
          <!-- workaround to resize the github and discord icons to match -->
          <div
            class="grid place-content-center text-gray-700 font-normal hover:bg-gray-200"
            style:width="22px"
            style:height="22px"
          >
            <svelte:component
              this={actionItem.icon}
              className={actionItem.className}
              size="16px"
            />
          </div>
        </a>
        <div slot="tooltip-content" transition:fly={{ duration: 100, y: 8 }}>
          <TooltipContent>
            <TooltipTitle>
              <svelte:fragment slot="name">
                {actionItem.tooltipTitle}
              </svelte:fragment>
            </TooltipTitle>
            <TooltipShortcutContainer>
              <div>{actionItem.tooltipCTA}</div>
              <Shortcut>Click</Shortcut>
            </TooltipShortcutContainer>
          </TooltipContent>
        </div>
      </Tooltip>
    {/each}
  </div>
  <div class="italic">
    version {$appBuildMetaStore.version
      ? $appBuildMetaStore.version
      : "unknown (built from source)"}{$appBuildMetaStore.commitHash
      ? ` – ${$appBuildMetaStore.commitHash}`
      : ""}
  </div>
</div>
