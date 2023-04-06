package auth

type OrganizationPermission int

const (
	ReadOrg OrganizationPermission = iota
	ManageOrg
	ReadProjects
	CreateProjects
	ManageProjects
	ReadOrgMembers
	ManageOrgMembers
)
