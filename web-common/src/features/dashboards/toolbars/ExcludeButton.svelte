<script lang="ts">
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Shortcut from "@rilldata/web-common/components/tooltip/Shortcut.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import TooltipShortcutContainer from "@rilldata/web-common/components/tooltip/TooltipShortcutContainer.svelte";
  import TooltipTitle from "@rilldata/web-common/components/tooltip/TooltipTitle.svelte";

  export let excludeMode: boolean;
  export let onClick: () => void;

  const getMode = (isExclude: boolean) => (isExclude ? "exclude" : "include");
</script>

<Tooltip distance={8} location="top">
  <Button type="toolbar" {onClick}>
    <Switch checked={excludeMode} small />
    <p>Exclude</p>
  </Button>
  <TooltipContent slot="tooltip-content">
    <TooltipTitle>
      <svelte:fragment slot="name">
        Output {getMode(excludeMode)}s selected values
      </svelte:fragment>
    </TooltipTitle>
    <TooltipShortcutContainer>
      <div>Toggle to {getMode(!excludeMode)} values</div>
      <Shortcut>Click</Shortcut>
    </TooltipShortcutContainer>
  </TooltipContent>
</Tooltip>
