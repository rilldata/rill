<script lang="ts">
  import ComponentError from "@rilldata/web-common/features/canvas/components/ComponentError.svelte";
  import type { V1ComponentSpecRendererProperties } from "@rilldata/web-common/runtime-client";
  import httpClient from "@rilldata/web-common/runtime-client/http-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import type { ImageSpec } from "./";
  import { getImagePosition } from "./util";

  export let rendererProperties: V1ComponentSpecRendererProperties;

  const instanceId = $runtime.instanceId;
  $: imageProperties = rendererProperties as ImageSpec;

  $: objectPosition = getImagePosition(imageProperties.alignment);

  let imageSrc: string | null = null;
  let errorMessage: string | null = null;
  $: {
    if (imageProperties.url) {
      fetchImage(imageProperties.url);
    } else {
      imageSrc = null;
      errorMessage = "No image URL provided";
    }
  }

  async function fetchImage(url: string) {
    try {
      imageSrc = await getImageURL(url);
      errorMessage = null;
    } catch (error) {
      imageSrc = null;
      errorMessage = error.message || "Failed to load image";
    }
  }

  const getImageURL = async (url: string): Promise<string> => {
    if (isValidURL(url)) return url;

    try {
      const response = (await httpClient({
        url: `/v1/instances/${instanceId}/assets/${url}`,
        method: "GET",
      })) as Response;

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const blob = await response.blob();
      return URL.createObjectURL(blob);
    } catch {
      throw new Error("Failed to fetch image from server");
    }
  };

  const isValidURL = (string: string) => {
    const regex = /^(https?):\/\/[^\s/$.?#].[^\s]*$/i;
    return regex.test(string);
  };
</script>

{#if errorMessage}
  <ComponentError error={errorMessage} />
{:else}
  <img
    src={imageSrc || ""}
    alt={"Canvas Image"}
    draggable="false"
    class="h-full w-full overflow-hidden object-contain"
    style:object-position={objectPosition}
  />
{/if}
