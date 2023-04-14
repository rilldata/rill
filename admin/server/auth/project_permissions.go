package auth

type ProjectPermission int

const (
	ReadProject ProjectPermission = iota
	ManageProject
	ReadProdBranch
	ManageProdBranch
	ReadDevBranches
	ManageDevBranches
	ReadProjectMembers
	ManageProjectMembers
)
