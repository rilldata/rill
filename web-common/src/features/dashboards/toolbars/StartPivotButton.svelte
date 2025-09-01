<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import Pivot from "../../../components/icons/Pivot.svelte";

  export let onClick: () => void;
</script>

<Button compact type="toolbar" {onClick}>
  <Pivot size="16px" />
  Start Pivot
</Button>
