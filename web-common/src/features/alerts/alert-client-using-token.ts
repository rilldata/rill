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
import httpClient from "@rilldata/web-admin/client/http-client";

export const adminServiceCreateAlertUsingToken = (
  org: string,
  project: string,
  adminServiceCreateAlertBodyBody: AdminServiceCreateAlertBodyBody,
  token: string,
) => {
  return httpClient<V1CreateAlertResponse>({
    url: `/v1/orgs/${org}/projects/${project}/alerts`,
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      // We use the bearer token to authenticate the request
      Authorization: `Bearer ${token}`,
    },
    data: adminServiceCreateAlertBodyBody,
    // To be explicit, we don't need to send credentials (cookies) with the request
    withCredentials: false,
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
  return httpClient<V1EditAlertResponse>({
    url: `/v1/orgs/${org}/projects/${project}/alerts/${name}`,
    method: "PUT",
    headers: {
      "Content-Type": "application/json",
      // We use the bearer token to authenticate the request
      Authorization: `Bearer ${token}`,
    },
    data: adminServiceCreateAlertBodyBody,
    // To be explicit, we don't need to send credentials (cookies) with the request
    withCredentials: false,
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
  return httpClient<V1DeleteAlertResponse>({
    url: `/v1/orgs/${org}/projects/${project}/alerts/${name}`,
    method: "DELETE",
    headers: {
      // We use the bearer token to authenticate the request
      Authorization: `Bearer ${token}`,
    },
    // To be explicit, we don't need to send credentials (cookies) with the request
    withCredentials: false,
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
  return httpClient<V1UnsubscribeAlertResponse>({
    url: `/v1/orgs/${org}/projects/${project}/alerts/${name}/unsubscribe`,
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      // We use the bearer token to authenticate the request
      Authorization: `Bearer ${token}`,
    },
    data: adminServiceUnsubscribeAlertBodyBody,
    // To be explicit, we don't need to send credentials (cookies) with the request
    withCredentials: false,
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
