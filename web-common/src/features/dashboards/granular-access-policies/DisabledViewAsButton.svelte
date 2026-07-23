<script lang="ts">
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import EyeIcon from "@rilldata/web-common/components/icons/EyeIcon.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
</script>

<Tooltip distance={8}>
  <button
    disabled
    class="px-3 py-1.5 rounded flex flex-row gap-x-2 items-center opacity-50 cursor-not-allowed"
  >
    <EyeIcon size={"16px"} />
    <div class="flex items-center gap-x-1">
      <span>{m.dashboard_view_as()}</span><CaretDownIcon />
    </div>
  </button>
  <TooltipContent slot="tooltip-content">
    {m.dashboard_view_as_disabled()}
  </TooltipContent>
</Tooltip>
