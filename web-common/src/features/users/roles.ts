export enum OrgUserRoles {
  Guest = "guest",
  Viewer = "viewer",
  Editor = "editor",
  Admin = "admin",
}

export enum ProjectUserRoles {
  Viewer = "viewer",
  Editor = "editor",
  Admin = "admin",
}

export function formatProjectRole(role: string): string {
  if (!role) return "";
  return role.charAt(0).toUpperCase() + role.slice(1).toLowerCase();
}
