/**
 * This file implements variants of the `CreateAlert`, `EditAlert`, `DeleteAlert` and `UnsubscribeAlert` client code that authenticate using a bearer token.
 * See `report-client-using-token.ts`, which this mirrors.
 */

import { type QueryClient } from "@tanstack/svelte-query";
import {
  type AdminServiceCreateAlertBodyBody,
  type AdminServiceUnsubscribeAlertBodyBody,
  createAdminServiceCreateAlert,
  createAdminServiceDeleteAlert,
  createAdminServiceEditAlert,
  createAdminServiceUnsubscribeAlert,
  type V1CreateAlertResponse,
  type V1DeleteAlertResponse,
  type V1EditAlertResponse,
  type V1UnsubscribeAlertResponse,
} from "@rilldata/web-admin/client";
import { httpClientWithToken } from "@rilldata/web-admin/client/http-client.ts";

export const adminServiceCreateAlertUsingToken = (
  org: string,
  project: string,
  adminServiceCreateAlertBodyBody: AdminServiceCreateAlertBodyBody,
  token: string,
) => {
  return httpClientWithToken<V1CreateAlertResponse>(token, {
    url: `/v1/orgs/${org}/projects/${project}/alerts`,
    method: "POST",
    headers: { "Content-Type": "application/json" },
    data: adminServiceCreateAlertBodyBody,
  });
};

export const createAdminServiceCreateAlertUsingToken = (
  token: string,
  queryClient?: QueryClient,
) =>
  createAdminServiceCreateAlert(
    {
      mutation: {
        mutationFn: ({ org, project, data }) =>
          adminServiceCreateAlertUsingToken(org, project, data, token),
      },
    },
    queryClient,
  );

export const adminServiceEditAlertUsingToken = (
  org: string,
  project: string,
  name: string,
  adminServiceCreateAlertBodyBody: AdminServiceCreateAlertBodyBody,
  token: string,
) => {
  return httpClientWithToken<V1EditAlertResponse>(token, {
    url: `/v1/orgs/${org}/projects/${project}/alerts/${name}`,
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    data: adminServiceCreateAlertBodyBody,
  });
};

export const createAdminServiceEditAlertUsingToken = (
  token: string,
  queryClient?: QueryClient,
) =>
  createAdminServiceEditAlert(
    {
      mutation: {
        mutationFn: ({ org, project, name, data }) =>
          adminServiceEditAlertUsingToken(org, project, name, data, token),
      },
    },
    queryClient,
  );

export const adminServiceDeleteAlertUsingToken = (
  org: string,
  project: string,
  name: string,
  token: string,
) => {
  return httpClientWithToken<V1DeleteAlertResponse>(token, {
    url: `/v1/orgs/${org}/projects/${project}/alerts/${name}`,
    method: "DELETE",
  });
};

export const createAdminServiceDeleteAlertUsingToken = (
  token: string,
  queryClient?: QueryClient,
) => {
  return createAdminServiceDeleteAlert(
    {
      mutation: {
        mutationFn: ({ org, project, name }) =>
          adminServiceDeleteAlertUsingToken(org, project, name, token),
      },
    },
    queryClient,
  );
};

export const adminServiceUnsubscribeAlertUsingToken = (
  org: string,
  project: string,
  name: string,
  adminServiceUnsubscribeAlertBodyBody: AdminServiceUnsubscribeAlertBodyBody,
  token: string,
) => {
  return httpClientWithToken<V1UnsubscribeAlertResponse>(token, {
    url: `/v1/orgs/${org}/projects/${project}/alerts/${name}/unsubscribe`,
    method: "POST",
    headers: { "Content-Type": "application/json" },
    data: adminServiceUnsubscribeAlertBodyBody,
  });
};

export const createAdminServiceUnsubscribeAlertUsingToken = (
  token: string,
  queryClient?: QueryClient,
) => {
  return createAdminServiceUnsubscribeAlert(
    {
      mutation: {
        mutationFn: ({ org, project, name, data }) =>
          adminServiceUnsubscribeAlertUsingToken(
            org,
            project,
            name,
            data,
            token,
          ),
      },
    },
    queryClient,
  );
};
