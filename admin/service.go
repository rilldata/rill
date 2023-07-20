package admin

import (
	"context"

	"github.com/rilldata/rill/admin/database"
	"go.uber.org/zap"
)

// CreateServiceForOrganization creates a new service account for an organization.
func (s *Service) CreateServiceForOrganization(ctx context.Context, orgName, serviceName string) (*database.Service, error) {
	ctx, tx, err := s.DB.NewTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	service, err := s.DB.InsertService(ctx, &database.InsertServiceOptions{
		OrgName: orgName,
		Name:    serviceName,
	})
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	s.logger.Info("created service", zap.String("name", serviceName), zap.String("org_name", orgName))

	return service, nil
}
