import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  getLocalServiceGetCurrentUserQueryKey,
  localServiceCheckOrgName,
  localServiceGetCurrentUser,
} from "@rilldata/web-common/runtime-client/local-service";

export async function getOrgName() {
  const userResp = await queryClient.fetchQuery({
    queryKey: getLocalServiceGetCurrentUserQueryKey(),
    queryFn: localServiceGetCurrentUser,
  });
  // TODO: handle non ascii names
  const userName = userResp.user!.displayName.replace(/ /g, "");

  let orgName = userName;
  let found = false;
  let i = 2;

  while (!found) {
    const resp = await localServiceCheckOrgName(orgName);
    found = resp.available;
    orgName = `${userName}-${i}`;
    i++;
  }

  return orgName;
}
