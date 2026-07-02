<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import Pivot from "../../../components/icons/Pivot.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  export let onClick: () => void;
</script>

<Button compact type="toolbar" {onClick}>
  <Pivot size="16px" />
  {m.dashboard_start_pivot()}
</Button>
