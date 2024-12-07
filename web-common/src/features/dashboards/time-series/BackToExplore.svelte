<script lang="ts">
  import { page } from "$app/stores";
  import Back from "@rilldata/web-common/components/icons/Back.svelte";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { getUrlForWebView } from "@rilldata/web-common/features/dashboards/url-state/explore-web-view-store";
  import { V1ExploreWebView } from "@rilldata/web-common/runtime-client";
  import { Button } from "../../../components/button";

  const { defaultExploreState } = getStateManagers();

  $: href = getUrlForWebView(
    $page.url,
    V1ExploreWebView.EXPLORE_WEB_VIEW_EXPLORE,
    $defaultExploreState,
  );
</script>

<a class="flex items-center" {href}>
  <Button type="link" forcedStyle="padding: 0; gap: 0px;">
    <Back size="16px" />
    <span>All measures</span>
  </Button>
</a>
