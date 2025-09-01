import { page } from "$app/stores";
import { getNeverSubscribedIssue } from "@rilldata/web-common/features/billing/issues";
import {
  createLocalServiceGetMetadata,
  createLocalServiceListOrganizationsAndBillingMetadataRequest,
} from "@rilldata/web-common/runtime-client/local-service";
import { derived } from "svelte/store";

export function getPlanUpgradeUrl(orgName: string) {
  const metadataQuery = createLocalServiceGetMetadata();
  const orgsMetadataQuery =
    createLocalServiceListOrganizationsAndBillingMetadataRequest();

  return derived(
    [metadataQuery, orgsMetadataQuery, page],
    ([metadata, orgsMetadata, pageState]) => {
      const adminUrl = metadata.data?.adminUrl;
      if (!adminUrl) return "";

      const metadataForOrg = orgsMetadata?.data?.orgs.find(
        (o) => o.name === orgName,
      );
      const isEmptyOrg =
        !!metadataForOrg?.issues &&
        !!getNeverSubscribedIssue(metadataForOrg.issues);

      // TODO: Find a better solution and get a url from backend.
      //       We should add an endpoint to get frontendUrl from the urls.go util on cloud.
      let cloudUrl = adminUrl.replace("admin.rilldata", "ui.rilldata");
      // hack for dev env
      if (cloudUrl === "http://localhost:8080") {
        cloudUrl = "http://localhost:3000";
      }

      const url = new URL(cloudUrl);
      if (isEmptyOrg) {
        // Empty org wont have billing related options so show the general setting page in the background
        url.pathname = `/${orgName}/-/settings`;
      } else {
        url.pathname = `/${orgName}/-/settings/billing`;
      }
      url.searchParams.set("upgrade", "true");
      const redirectUrl = new URL(pageState.url);
      // set the org to avoid showing the org selector again
      redirectUrl.searchParams.set("org", orgName);
      url.searchParams.set("redirect", redirectUrl.toString());
      return url.toString();
    },
  );
}

export function getIsOrgOnTrial(orgName: string) {
  return derived(
    createLocalServiceListOrganizationsAndBillingMetadataRequest(),
    (orgsMetadata) => {
      const metadataForOrg = orgsMetadata?.data?.orgs.find(
        (o) => o.name === orgName,
      );
      return (
        !!orgName &&
        !!metadataForOrg?.issues &&
        !!getNeverSubscribedIssue(metadataForOrg.issues)
      );
    },
  );
}
