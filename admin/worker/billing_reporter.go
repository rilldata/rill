package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/rilldata/rill/admin/billing"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/metrics"
	"go.uber.org/zap"
)

var defaultReportableMetrics = []string{"data_dir_size_bytes"}

func (w *Worker) reportUsage(ctx context.Context) error {
	// Get reporting granularity
	var granularity time.Duration
	var validationGracePeriod time.Duration
	switch w.admin.Biller.GetReportingGranularity() {
	case billing.UsageReportingGranularityHour:
		granularity = time.Hour
		validationGracePeriod = 3 * time.Hour
	case billing.UsageReportingGranularityNone:
		w.logger.Debug("skipping usage reporting: no reporting granularity configured")
		return nil
	default:
		return fmt.Errorf("unsupported reporting granularity: %s", w.admin.Biller.GetReportingGranularity())
	}

	// round down to the nearest granularity, start is inclusive and end is exclusive
	// report usage for previous granularity
	bucketStart := time.Now().Truncate(granularity).Add(-granularity)
	bucketEnd := bucketStart.Add(granularity)

	// Get metrics client
	client, ok, err := w.admin.OpenMetricsProject(ctx)
	if err != nil {
		w.logger.Error("failed to report usage: unable to get metrics client", zap.Error(err))
		return err
	}
	if !ok {
		w.logger.Debug("skipping usage reporting: no metrics project configured")
		return nil
	}

	limit := 100
	afterName := ""
	stop := false
	for !stop {
		// get all orgs
		orgs, err := w.admin.DB.FindOrganizations(ctx, afterName, limit)
		if err != nil {
			w.logger.Error("failed to report usage: unable to fetch organizations", zap.Error(err))
			return err
		}
		if len(orgs) < limit {
			stop = true
		}
		if len(orgs) != 0 {
			afterName = orgs[len(orgs)-1].Name
		}

		for _, org := range orgs {
			err = w.reportOrg(ctx, client, org, bucketStart, bucketEnd, validationGracePeriod, granularity)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (w *Worker) reportOrg(ctx context.Context, client *metrics.Client, org *database.Organization, startTime, endTime time.Time, gracePeriod, granularity time.Duration) error {
	// Get reportable metrics for the org
	reportableMetrics, err := w.getReportableMetrics(ctx, org)
	if err != nil {
		return fmt.Errorf("failed to report usage for org %s: %w", org.Name, err)
	}
	if len(reportableMetrics) == 0 {
		w.logger.Warn("skipping usage reporting for org: no reportable metrics found", zap.String("org", org.Name))
		return nil
	}

	var usage []*billing.Usage
	var reportableProjects []string

	limit := 100
	afterName := ""
	stop := false
	for !stop {
		// Get all projects for the org
		projects, err := w.admin.DB.FindProjectsForOrganization(ctx, org.ID, afterName, limit)
		if err != nil {
			return fmt.Errorf("failed to report usage for org %s: unable to fetch projects: %w", org.Name, err)
		}

		if len(projects) < limit {
			stop = true
		}
		if len(projects) == 0 {
			break
		}
		afterName = projects[len(projects)-1].Name

		var filteredProjects []string
		projs := make(map[string]*database.Project)
		validationStart := startTime
		for _, project := range projects {
			if project.CreatedOn.Equal(endTime) || project.CreatedOn.After(endTime) {
				continue
			}

			if project.NextUsageReportingTime.Equal(endTime) || project.NextUsageReportingTime.After(endTime) {
				// this cannot happen, but just in case
				w.logger.Warn("skipping usage reporting for project: already reported", zap.String("org", org.Name), zap.String("project", project.Name))
				continue
			}

			filteredProjects = append(filteredProjects, project.ID)
			projs[project.ID] = project
			if project.NextUsageReportingTime.IsZero() {
				continue
			} else if project.NextUsageReportingTime.Before(validationStart) {
				validationStart = project.NextUsageReportingTime
			}
		}

		if len(filteredProjects) == 0 {
			continue
		}

		// add grace period buffer to lower and upper bound to ensure correct comparison
		validationStart = validationStart.Add(-gracePeriod)
		validationEnd := endTime.Add(gracePeriod)

		// Get availability
		availability, err := client.GetProjectUsageAvailability(ctx, filteredProjects, validationStart, validationEnd)
		if err != nil {
			return fmt.Errorf("failed to report usage for org %s: unable to get project availability: %w", org.Name, err)
		}
		u, a, err := w.collectProjectsUsage(ctx, client, org, projs, availability, startTime, endTime, granularity, reportableMetrics)
		if err != nil {
			return err
		}
		reportableProjects = append(reportableProjects, a...)
		usage = append(usage, u...)
	}

	if len(usage) > 0 {
		err = w.admin.Biller.ReportUsage(ctx, org.ID, usage)
		if err != nil {
			return fmt.Errorf("failed to report usage for org %s: %w", org.Name, err)
		}
		w.logger.Info("reported usage", zap.String("org", org.Name), zap.Int("num_usage_records", len(usage)))

		// update next usage reporting time for all projects
		err = w.admin.DB.UpdateProjectsNextUsageReportingTime(ctx, reportableProjects, endTime)
		if err != nil {
			return fmt.Errorf("failed to update next usage reporting time for org %s: %w", org.Name, err)
		}
	} else {
		w.logger.Info("no usage to report", zap.String("org", org.Name))
	}
	return nil
}

func (w *Worker) collectProjectsUsage(ctx context.Context, client *metrics.Client, org *database.Organization, projs map[string]*database.Project, availability []*metrics.ProjectUsageAvailability, startTime, endTime time.Time, grain time.Duration, metricNames []string) ([]*billing.Usage, []string, error) {
	var orgUsage []*billing.Usage
	var reportableProjects []string

	missingAvailability := make(map[string]struct{})
	for _, proj := range projs {
		missingAvailability[proj.Name] = struct{}{}
	}

	for _, av := range availability {
		proj, ok := projs[av.ProjectID]
		if !ok {
			return nil, nil, fmt.Errorf("project not found: %s", av.ProjectID)
		}
		delete(missingAvailability, proj.Name)

		projectReportingStartTime := startTime
		projectReportingEndTime := endTime

		// handle special cases first
		if proj.NextUsageReportingTime.IsZero() {
			// a new project - validate available min time is less than reporting end time. If yes then report usage and update next reporting time to reporting end time and continue
			// if not then may be usage is not available yet in metrics project or a corner case where usage event started reporting after the reporting end time
			// in both cases just continue and let the next iteration handle it
			// note - reporting window will shift to next granularity, so we may miss reporting for this granularity which is a rare corner case
			if av.MinTime.Before(projectReportingEndTime) {
				u, err := w.getProjectUsage(ctx, client, org.ID, proj.ID, projectReportingStartTime, projectReportingEndTime, grain, metricNames)
				if err != nil {
					return nil, nil, fmt.Errorf("failed to get usage for org: %s project %s: %w", org.Name, proj.Name, err)
				}
				reportableProjects = append(reportableProjects, proj.ID)
				orgUsage = append(orgUsage, u...)
			} else {
				w.logger.Warn("skipping usage reporting for project: no usage available", zap.String("org", org.Name), zap.String("project", proj.Name), zap.Time("min_time", av.MinTime), zap.Time("max_time", av.MaxTime), zap.Time("reporting_start_time", projectReportingStartTime), zap.Time("reporting_end_time", projectReportingEndTime), zap.Time("next_usage_reporting_time", proj.NextUsageReportingTime))
			}
		} else if proj.ProdDeploymentID == nil {
			// hibernating project - don't perform any validation, just report whatever is available and update next reporting time to projectReportingEndTime
			u, err := w.getProjectUsage(ctx, client, org.ID, proj.ID, projectReportingStartTime, projectReportingEndTime, grain, metricNames)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to get usage for org: %s project %s: %w", org.Name, proj.Name, err)
			}
			reportableProjects = append(reportableProjects, proj.ID)
			orgUsage = append(orgUsage, u...)
		} else {
			// active project
			projectReportingStartTime = proj.NextUsageReportingTime
			// validate availability against reporting bounds
			if av.MinTime.After(projectReportingStartTime) || projectReportingEndTime.After(av.MaxTime) {
				w.logger.Warn("skipping usage reporting for project: availability mismatch", zap.String("org", org.Name), zap.String("project", proj.Name), zap.Time("min_time", av.MinTime), zap.Time("max_time", av.MaxTime), zap.Time("reporting_start_time", projectReportingStartTime), zap.Time("reporting_end_time", projectReportingEndTime), zap.Time("next_usage_reporting_time", proj.NextUsageReportingTime))
			} else {
				u, err := w.getProjectUsage(ctx, client, org.ID, proj.ID, projectReportingStartTime, projectReportingEndTime, grain, metricNames)
				if err != nil {
					return nil, nil, fmt.Errorf("failed to get usage for org: %s project %s: %w", org.Name, proj.Name, err)
				}
				reportableProjects = append(reportableProjects, proj.ID)
				orgUsage = append(orgUsage, u...)
			}
		}
	}

	// reporting done for all available projects
	// now print warning for any missing projects if any, generally this should not happen
	for p := range missingAvailability {
		w.logger.Warn("skipping usage reporting for project: no availability found", zap.String("org", org.Name), zap.String("project", p))
	}

	return orgUsage, reportableProjects, nil
}

func (w *Worker) getProjectUsage(ctx context.Context, client *metrics.Client, orgID, projectID string, start, end time.Time, grain time.Duration, metricNames []string) ([]*billing.Usage, error) {
	usageMetadata := map[string]interface{}{"org_id": orgID, "project_id": projectID}
	var reportableUsage []*billing.Usage
	// generally start and end time should align with time gran but any ways doing basic checks
	for start.Before(end) {
		upperBound := start.Add(grain)
		// should not happen but just in case
		if upperBound.After(end) {
			upperBound = end
		}
		usage, err := client.GetProjectUsageMetrics(ctx, projectID, start, upperBound, metricNames)
		if err != nil {
			return nil, err
		}
		for _, u := range usage {
			reportableUsage = append(reportableUsage, &billing.Usage{
				CustomerID:     orgID,
				MetricName:     u.MetricName,
				Amount:         u.Amount,
				ReportingGrain: w.admin.Biller.GetReportingGranularity(),
				StartTime:      start,
				EndTime:        upperBound,
				Metadata:       usageMetadata,
			})
		}
		start = upperBound
	}
	return reportableUsage, nil
}

func (w *Worker) getReportableMetrics(ctx context.Context, org *database.Organization) ([]string, error) {
	if org.BillingCustomerID == "" {
		return nil, nil
	}
	subs, err := w.admin.Biller.GetSubscriptionsForCustomer(ctx, org.BillingCustomerID)
	if err != nil {
		return nil, err
	}
	var reportableMetrics []string
	for _, sub := range subs {
		reportableMetrics = append(reportableMetrics, sub.Plan.ReportableMetrics...)
	}

	if len(reportableMetrics) == 0 {
		return defaultReportableMetrics, nil
	}

	return unique(reportableMetrics), nil
}

func unique(s []string) []string {
	seen := make(map[string]struct{}, len(s))
	j := 0
	for _, v := range s {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		s[j] = v
		j++
	}
	return s[:j]
}
