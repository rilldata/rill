import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
import { OrgUserRoles } from "@rilldata/web-common/features/users/roles";

export function getOrgRolesOptions() {
  return [
    {
      value: OrgUserRoles.Admin,
      label: m.role_admin(),
      description: m.role_org_admin_desc(),
    },
    {
      value: OrgUserRoles.Editor,
      label: m.role_editor(),
      description: m.role_org_editor_desc(),
    },
    {
      value: OrgUserRoles.Viewer,
      label: m.role_viewer(),
      description: m.role_org_viewer_desc(),
    },
  ];
}
