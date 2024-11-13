import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  getLocalServiceGetUserOrgMetadataRequestQueryKey,
  localServiceGetUserOrgMetadataRequest,
} from "@rilldata/web-common/runtime-client/local-service";

export const load = async ({ url }) => {
  const orgMetadata = await queryClient.fetchQuery({
    queryKey: getLocalServiceGetUserOrgMetadataRequestQueryKey(),
    queryFn: () => localServiceGetUserOrgMetadataRequest(),
    staleTime: Infinity,
  });
  return {
    orgParam: url.searchParams.get("org") ?? "",
    orgMetadata,
  };
};
