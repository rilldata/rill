<script lang="ts">
  import Back from "@rilldata/web-common/components/icons/Back.svelte";
  import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params";
  import { Button } from "../../../components/button";
</script>

<a class="flex items-center" href="?{ExploreStateURLParams.WebView}=explore">
  <Button forcedStyle="padding: 0; gap: 0px;" type="link">
    <Back size="16px" />
    <span>All measures</span>
  </Button>
</a>
