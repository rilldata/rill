<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import Pivot from "../../../components/icons/Pivot.svelte";

  export let onClick: () => void;
</script>

<Button compact type="toolbar" {onClick}>
  <Pivot size="16px" />
  {m.dashboards_toolbar_start_pivot()}
</Button>
