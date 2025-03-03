<script lang="ts">
  import { page } from "$app/stores";
  import Back from "@rilldata/web-common/components/icons/Back.svelte";
  import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params";
  import { Button } from "../../../components/button";

  function makeExploreHref(pageUrl: URL) {
    const currentUrlParams = new URLSearchParams(pageUrl.search);
    currentUrlParams.set(ExploreStateURLParams.WebView, "explore");
    return `?${currentUrlParams.toString()}`;
  }
</script>

<a class="flex items-center" href={makeExploreHref($page.url)}>
  <Button type="link" forcedStyle="padding: 0; gap: 0px;">
    <Back size="16px" />
    <span>All measures</span>
  </Button>
</a>
