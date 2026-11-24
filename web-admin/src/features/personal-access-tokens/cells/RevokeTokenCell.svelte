<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  let { onRevoke }: { onRevoke: () => void } = $props();
</script>

<Button type="secondary-destructive" small onClick={onRevoke}>
  {m.access_tokens_revoke()}
</Button>
