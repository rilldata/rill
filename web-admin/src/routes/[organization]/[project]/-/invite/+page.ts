import { getSingleUseUrlParam } from "@rilldata/web-admin/features/navigation/getSingleUseUrlParam";

export const load = async ({ url }) => {
  const showWelcomeDialog = !!getSingleUseUrlParam(
    url,
    "welcome",
    "rill:app:showWelcome",
  );
  return {
    showWelcomeDialog,
  };
};
