package auth

type ProjectPermission int

const (
	ReadProject ProjectPermission = iota
	ManageProject
	ReadProd
	ReadProdStatus
	ManageProd
	ReadDev
	ReadDevStatus
	ManageDev
	ReadProjectMembers
	ManageProjectMembers
)
