<script lang="ts">
  import { setContext } from "svelte";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { createBusinessModel, DEFAULT_STORE_KEY } from "./business-model";

  export let metricsViewName: string;

  const queryClient = useQueryClient();
  const businessModel = createBusinessModel({ queryClient, metricsViewName });
  setContext(DEFAULT_STORE_KEY, businessModel);

  $: {
    // update metrics view name
    businessModel.setMetricsViewName(metricsViewName);
  }
</script>

<slot />
