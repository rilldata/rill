import type {
  V1OrganizationMemberUser,
  V1OrganizationInvite,
  V1MemberUsergroup,
} from "@rilldata/web-admin/client";

export interface SearchListItem {
  identifier: string;
  type: "user" | "group";
  name?: string;
  photoUrl?: string;
  orgRoleName?: string;
  invitedBy?: string;
  groupCount?: number;
}

// Optimized searchList computation with memoization and O(1) lookups
export function buildSearchList(
  allOrgMemberUsersRows: V1OrganizationMemberUser[],
  allOrgInvitesRows: V1OrganizationInvite[],
  orgMemberUsergroups: V1MemberUsergroup[],
  projectMemberEmailSet: Set<string>,
  projectInviteEmailSet: Set<string>,
  projectUserGroupNameSet: Set<string>,
): SearchListItem[] {
  const result: SearchListItem[] = [];

  // Process org member users
  for (const member of allOrgMemberUsersRows) {
    if (
      member.userEmail &&
      !projectMemberEmailSet.has(member.userEmail) &&
      !projectInviteEmailSet.has(member.userEmail)
    ) {
      result.push({
        identifier: member.userEmail,
        type: "user",
        name: member.userName,
        photoUrl: member.userPhotoUrl,
        orgRoleName: member.roleName,
      });
    }
  }

  // Process org invites
  for (const invite of allOrgInvitesRows) {
    if (
      invite.email &&
      !projectMemberEmailSet.has(invite.email) &&
      !projectInviteEmailSet.has(invite.email)
    ) {
      result.push({
        identifier: invite.email,
        type: "user",
        name: invite.email,
        photoUrl: undefined,
        orgRoleName: invite.roleName,
        invitedBy: invite.invitedBy,
      });
    }
  }

  // Process org member usergroups
  for (const group of orgMemberUsergroups) {
    if (
      group.groupName &&
      !group.groupManaged &&
      !projectUserGroupNameSet.has(group.groupName)
    ) {
      result.push({
        identifier: group.groupName,
        groupCount: group.usersCount,
        type: "group",
      });
    }
  }

  return result;
}

export function buildCopyLink(
  pageUrl: URL,
  organization: string,
  project: string,
): string {
  return `${pageUrl.protocol}//${pageUrl.host}/${organization}/${project}`;
}
