package worker

import (
	"context"
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
	}

	// round down to the nearest granularity, start is inclusive and end is exclusive
	// report usage for previous granularity
	workerTimeBucketStart := time.Now().Truncate(granularity).Add(-granularity)
	workerTimeBucketEnd := workerTimeBucketStart.Add(granularity)

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

	// get all orgs
	orgs, err := w.admin.DB.FindOrganizations(ctx, "", 100000)
	if err != nil {
		w.logger.Error("failed to report usage: unable to fetch organizations", zap.Error(err))
		return err
	}

	for _, org := range orgs {
		// Get reportable metrics for the org
		reportableMetrics, err := w.getReportableMetrics(ctx, org)
		if err != nil {
			w.logger.Error("failed to report usage: unable to get reportable metrics", zap.String("org", org.Name), zap.Error(err))
			return err
		}
		if len(reportableMetrics) == 0 {
			w.logger.Warn("skipping usage reporting for org: no reportable metrics found", zap.String("org", org.Name))
			continue
		}

		// Get all projects for the org
		projects, err := w.admin.DB.FindProjectsForOrganization(ctx, org.ID, "", 100000)
		if err != nil {
			w.logger.Error("failed to report usage: unable to fetch projects", zap.String("org", org.Name), zap.Error(err))
			return err
		}

		if len(projects) == 0 {
			w.logger.Warn("skipping usage reporting for org: no projects found", zap.String("org", org.Name))
			continue
		}

		var reportableProjects []string
		remaining := make(map[string]*database.Project)
		validationLowerTimeBound := workerTimeBucketStart
		for _, project := range projects {
			if project.CreatedOn.Equal(workerTimeBucketEnd) || project.CreatedOn.After(workerTimeBucketEnd) {
				continue
			}

			if project.NextUsageReportingTime.Equal(workerTimeBucketEnd) || project.NextUsageReportingTime.After(workerTimeBucketEnd) {
				// this cannot happen, but just in case
				w.logger.Warn("skipping usage reporting for project: already reported", zap.String("org", org.Name), zap.String("project", project.Name))
				continue
			}

			reportableProjects = append(reportableProjects, project.ID)
			remaining[project.ID] = project
			if project.NextUsageReportingTime.IsZero() {
				continue
			} else if project.NextUsageReportingTime.Before(validationLowerTimeBound) {
				validationLowerTimeBound = project.NextUsageReportingTime
			}
		}

		if len(reportableProjects) == 0 {
			w.logger.Warn("skipping usage reporting for org: no projects to report", zap.String("org", org.Name))
			continue
		}

		// add grace period buffer to lower and upper bound to ensure correct comparison
		validationLowerTimeBound = validationLowerTimeBound.Add(-validationGracePeriod)
		validationUpperTimeBound := workerTimeBucketEnd.Add(validationGracePeriod)

		// Get availability
		availability, err := client.GetProjectUsageAvailability(ctx, reportableProjects, validationLowerTimeBound, validationUpperTimeBound)
		if err != nil {
			w.logger.Error("failed to report usage: unable to get availability", zap.String("org", org.Name), zap.Error(err))
			continue
		}

		for _, av := range availability {
			proj := remaining[av.ProjectID]
			if proj == nil {
				// cannot happen, but just in case
				w.logger.Error("failed to report usage: project id mismatch", zap.String("org", org.Name), zap.String("project", av.ProjectID))
				continue
			}
			delete(remaining, av.ProjectID)

			projectReportingStartTime := workerTimeBucketStart
			projectReportingEndTime := workerTimeBucketEnd

			// handle special cases first
			if proj.NextUsageReportingTime.IsZero() {
				// a new project - validate available min time is less than reporting end time. If yes then report usage and update next reporting time to reporting end time and continue
				// if not then may be usage is not available yet in metrics project or a corner case where usage event started reporting after the reporting end time
				// in both cases just continue and let the next iteration handle it
				// note - reporting window will shift to next granularity, so we may miss reporting for this granularity which is a rare corner case
				if av.MinTime.Before(projectReportingEndTime) {
					// report usage and update next reporting time to reporting end time
					err = w.reportProjectUsage(ctx, client, org.ID, proj.ID, projectReportingStartTime, projectReportingEndTime, granularity, reportableMetrics)
					if err != nil {
						w.logger.Error("failed to report usage", zap.String("org", org.Name), zap.String("project", proj.Name), zap.Error(err))
						return err
					}
					err = w.admin.DB.UpdateProjectNextUsageReportingTime(ctx, proj.ID, projectReportingEndTime)
					if err != nil {
						w.logger.Error("failed to update project usage reporting time", zap.String("org", org.Name), zap.String("project", proj.Name), zap.Error(err))
						return err
					}
				} else {
					w.logger.Warn("skipping usage reporting for project: no usage available", zap.String("org", org.Name), zap.String("project", proj.Name), zap.Time("min_time", av.MinTime), zap.Time("max_time", av.MaxTime), zap.Time("reporting_start_time", projectReportingStartTime), zap.Time("reporting_end_time", projectReportingEndTime), zap.Time("next_usage_reporting_time", proj.NextUsageReportingTime))
				}
			} else if proj.ProdDeploymentID == nil {
				// hibernating project - don't perform any validation, just report whatever is available and update next reporting time to projectReportingEndTime
				err = w.reportProjectUsage(ctx, client, org.ID, proj.ID, projectReportingStartTime, projectReportingEndTime, granularity, reportableMetrics)
				if err != nil {
					w.logger.Error("failed to report usage of hibernating project", zap.String("org", org.Name), zap.String("project", proj.Name), zap.Error(err))
					return err
				}
			} else {
				// active project
				projectReportingStartTime = proj.NextUsageReportingTime
				// validate availability against reporting bounds
				if av.MinTime.After(projectReportingStartTime) || projectReportingEndTime.After(av.MaxTime) {
					w.logger.Warn("skipping usage reporting for project: availability mismatch", zap.String("org", org.Name), zap.String("project", proj.Name), zap.Time("min_time", av.MinTime), zap.Time("max_time", av.MaxTime), zap.Time("reporting_start_time", projectReportingStartTime), zap.Time("reporting_end_time", projectReportingEndTime), zap.Time("next_usage_reporting_time", proj.NextUsageReportingTime))
				} else {
					err = w.reportProjectUsage(ctx, client, org.ID, proj.ID, projectReportingStartTime, projectReportingEndTime, granularity, reportableMetrics)
					if err != nil {
						w.logger.Error("failed to report usage", zap.String("org", org.Name), zap.String("project", proj.Name), zap.Error(err))
						return err
					}
					err = w.admin.DB.UpdateProjectNextUsageReportingTime(ctx, proj.ID, projectReportingEndTime)
					if err != nil {
						w.logger.Error("failed to update project usage reporting time", zap.String("org", org.Name), zap.String("project", proj.Name), zap.Error(err))
						return err
					}
				}
			}
		}
		// reporting done for all available projects
		// now print warning for any missing projects if any, generally this should not happen
		for _, proj := range remaining {
			w.logger.Warn("skipping usage reporting for project: no availability found", zap.String("org", org.Name), zap.String("project", proj.Name))
		}
		// done with org, move to next org
	}
	return nil
}

func (w *Worker) reportProjectUsage(ctx context.Context, client *metrics.Client, orgID, projectID string, start, end time.Time, gran time.Duration, metricNames []string) error {
	usageMetadata := map[string]interface{}{"org_id": orgID, "project_id": projectID}
	// generally start and end time should align with time gran but any ways doing basic checks
	for start.Before(end) {
		upperBound := start.Add(gran)
		// should not happen but just in case
		if upperBound.After(end) {
			upperBound = end
		}
		usage, err := client.GetProjectUsageMetrics(ctx, projectID, start, upperBound, metricNames)
		if err != nil {
			return err
		}
		var reportableUsage []*billing.Usage
		// set usage time to 1 sec before upper bound to assign usage to current bucket
		for _, u := range usage {
			reportableUsage = append(reportableUsage, &billing.Usage{
				MetricName:    u.MetricName,
				Amount:        u.Amount,
				ReportingGran: w.admin.Biller.GetReportingGranularity(),
				StartTime:     start,
				EndTime:       upperBound,
				Metadata:      usageMetadata,
			})
		}
		err = w.admin.Biller.ReportUsage(ctx, orgID, reportableUsage)
		if err != nil {
			return err
		}
		start = upperBound
	}
	return nil
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
