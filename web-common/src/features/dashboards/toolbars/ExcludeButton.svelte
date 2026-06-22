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
</script>

<Tooltip distance={8} location="top">
  <Button type="toolbar" {onClick}>
    <Switch checked={excludeMode} small />
    <p>{m.dashboards_toolbar_exclude_label()}</p>
  </Button>
  <TooltipContent slot="tooltip-content">
    <TooltipTitle>
      <svelte:fragment slot="name">
        {excludeMode
          ? m.dashboards_toolbar_exclude_output_exclude()
          : m.dashboards_toolbar_exclude_output_include()}
      </svelte:fragment>
    </TooltipTitle>
    <TooltipShortcutContainer>
      <div>
        {excludeMode
          ? m.dashboards_toolbar_exclude_toggle_include()
          : m.dashboards_toolbar_exclude_toggle_exclude()}
      </div>
      <Shortcut>{m.dashboards_toolbar_shortcut_click()}</Shortcut>
    </TooltipShortcutContainer>
  </TooltipContent>
</Tooltip>
