<script lang="ts">
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Shortcut from "@rilldata/web-common/components/tooltip/Shortcut.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import TooltipShortcutContainer from "@rilldata/web-common/components/tooltip/TooltipShortcutContainer.svelte";
  import TooltipTitle from "@rilldata/web-common/components/tooltip/TooltipTitle.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  export let excludeMode: boolean;
  export let onClick: () => void;

  const getMode = (isExclude: boolean) => (isExclude ? "exclude" : "include");
</script>

<Tooltip distance={8} location="top">
  <Button type="toolbar" {onClick}>
    <Switch checked={excludeMode} small />
    <p>{m.dashboard_exclude()}</p>
  </Button>
  <TooltipContent slot="tooltip-content">
    <TooltipTitle>
      <svelte:fragment slot="name">
        {excludeMode
          ? m.dashboard_output_excludes()
          : m.dashboard_output_includes()}
      </svelte:fragment>
    </TooltipTitle>
    <TooltipShortcutContainer>
      <div>
        {excludeMode
          ? m.dashboard_toggle_include()
          : m.dashboard_toggle_exclude()}
      </div>
      <Shortcut>{m.footer_shortcut_click()}</Shortcut>
    </TooltipShortcutContainer>
  </TooltipContent>
</Tooltip>
