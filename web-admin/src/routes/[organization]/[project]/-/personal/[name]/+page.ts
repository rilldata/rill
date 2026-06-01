import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
import {
  adminServiceGetPersonalFile,
  getAdminServiceGetPersonalFileQueryKey,
} from "@rilldata/web-admin/client";

export const load = async ({ params: { organization, project, name } }) => {
  const personalFile = await queryClient.fetchQuery({
    queryKey: getAdminServiceGetPersonalFileQueryKey(
      organization,
      project,
      name,
    ),
    queryFn: () => adminServiceGetPersonalFile(organization, project, name),
  });

  return {
    personalFile,
  };
};
