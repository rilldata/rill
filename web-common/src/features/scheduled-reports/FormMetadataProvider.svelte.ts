import { page } from "$app/state";
import {
  createAdminServiceGetCurrentUser,
  createAdminServiceListProjectMemberUsers,
} from "@rilldata/web-admin/client";
import { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { EmbedStore } from "@rilldata/web-common/features/embeds/embed-store.ts";
import {
  createRuntimeServiceGetInstance,
  createRuntimeServiceListNotifierConnectors,
} from "@rilldata/web-common/runtime-client";

/**
 * Creates the create/edit mutation for a form.
 * Admin mutations authenticate using cookies, embed mutations authenticate using the embed's admin token.
 */
export type MutationFactory<M> = {
  admin: () => M;
  embed: (adminToken: string) => M;
};

/**
 * Provides the metadata needed by alert and report forms,
 * abstracting away whether the form is rendered in Rill Cloud or in an embed.
 */
export interface FormMetadataProvider<M> {
  mutation: M;

  organization: string;
  project: string;
  pageBasePath: string;

  userEmail: string;
  // True until the current user has loaded.
  isLoading: boolean;
  projectMembersSet: Set<string>;
  hasSlackNotifier: boolean;

  cleanup(): void;
}

export function createFormMetadataProvider<M>(
  client: RuntimeClient,
  mutationFactory: MutationFactory<M>,
): FormMetadataProvider<M> {
  return EmbedStore.isEmbedded()
    ? new EmbedFormMetadataProvider(client, mutationFactory.embed)
    : new AdminFormMetadataProvider(client, mutationFactory.admin);
}

class AdminFormMetadataProvider<M> implements FormMetadataProvider<M> {
  public mutation: M;

  public organization: string;
  public project: string;
  public pageBasePath: string;

  public userEmail = $state<string>("");
  public isLoading = $state(true);
  public projectMembersSet = $state<Set<string>>(new Set());
  public hasSlackNotifier = $state(false);

  private readonly userUnsub: () => void;
  private readonly projectMembersUnsub: () => void;
  private readonly notifiersUnsub: () => void;

  public constructor(
    client: RuntimeClient,
    createMutation: MutationFactory<M>["admin"],
  ) {
    this.mutation = createMutation();

    this.organization = page.params.organization;
    this.project = page.params.project;
    this.pageBasePath = `/${this.organization}/${this.project}/-`;

    const user = createAdminServiceGetCurrentUser();
    this.userUnsub = user.subscribe((userResp) => {
      this.isLoading = userResp.isPending;
      if (!userResp.data?.user?.email) return;
      this.userEmail = userResp.data.user.email;
    });

    const listProjectMemberUsersQuery =
      createAdminServiceListProjectMemberUsers(this.organization, this.project);
    this.projectMembersUnsub = listProjectMemberUsersQuery.subscribe(
      (projectMembersResp) => {
        this.projectMembersSet = new Set(
          projectMembersResp.data?.members?.map((m) => m.userEmail!) ?? [],
        );
      },
    );

    const notifierConnectorsQuery = createRuntimeServiceListNotifierConnectors(
      client,
      {},
    );
    this.notifiersUnsub = notifierConnectorsQuery.subscribe(
      (notifierConnectorsResp) => {
        this.hasSlackNotifier =
          !!notifierConnectorsResp?.data?.connectors?.some(
            (c) => c.name === "slack",
          );
      },
    );
  }

  public cleanup() {
    this.userUnsub();
    this.projectMembersUnsub();
    this.notifiersUnsub();
  }
}

class EmbedFormMetadataProvider<M> implements FormMetadataProvider<M> {
  public mutation: M;

  public organization = "";
  public project = "";
  public pageBasePath = "/-/embed";
  public userEmail = "";
  public isLoading = false;
  public projectMembersSet = new Set<string>();
  public hasSlackNotifier = false; // not support in embed

  private readonly instanceUnsub: () => void;

  public constructor(
    client: RuntimeClient,
    createMutation: MutationFactory<M>["embed"],
  ) {
    const embedInstance = EmbedStore.getInstance();
    if (!embedInstance) {
      throw new Error("missing embed instance");
    }
    this.userEmail = embedInstance.userEmail;
    this.projectMembersSet = new Set([this.userEmail]);

    this.mutation = createMutation(embedInstance.adminToken);

    const instanceQuery = createRuntimeServiceGetInstance(client, {});
    this.instanceUnsub = instanceQuery.subscribe((instanceResp) => {
      const url = instanceResp.data?.instance?.frontendUrl;
      if (!url) return;
      const parts = url.split("/");
      this.organization = parts[parts.length - 2] ?? "";
      this.project = parts[parts.length - 1] ?? "";
    });
  }

  public cleanup() {
    this.instanceUnsub();
  }
}
