package runtime

import (
	"context"
	"errors"
	"maps"
	"time"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/ctxsync"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.uber.org/zap"
)

// configReloader handles reloading instance configurations from admin service
// periodically reload configs in background
// config reloads happen in following scenarios:
// 1. via rt.ReloadConfig whenever admin wants runtime to reload its configs
// 2. whenever runtime is started to pick up any configs changed while it was down
// 3. periodically every hour to pick up any changes done outside of runtime. This just adds extra resilience
type configReloader struct {
	rt *Runtime
	// cancel background operations on close
	cancel context.CancelFunc
	mu     ctxsync.RWMutex
	// to avoid repo handshake refresh on each reload we track last updatedon of each deployment
	// if the deployment has not changed skip the repo pull
	//
	// this can further be optimised to only check for properties that affect repo like url, branch etc but keeping it simple for now
	updatedOn map[string]time.Time
}

func newConfigReloader(rt *Runtime) *configReloader {
	bgctx, bgcancel := context.WithCancel(context.Background())
	c := &configReloader{
		rt:        rt,
		cancel:    bgcancel,
		mu:        ctxsync.NewRWMutex(),
		updatedOn: make(map[string]time.Time),
	}

	go c.periodicallyReloadConfigs(bgctx)
	return c
}

func (r *configReloader) reloadConfig(ctx context.Context, instanceID string) error {
	err := r.mu.Lock(ctx)
	if err != nil {
		return err
	}
	defer r.mu.Unlock()

	inst, err := r.rt.Instance(ctx, instanceID)
	if err != nil {
		return err
	}

	admin, release, err := r.rt.Admin(ctx, instanceID)
	if err != nil {
		if errors.Is(err, ErrAdminNotConfigured) {
			return nil
		}
		return err
	}
	defer release()

	r.rt.Logger.Info("Reloading config for instance", zap.String("instance_id", instanceID), observability.ZapCtx(ctx))

	cfg, err := admin.GetDeploymentConfig(ctx)
	if err != nil {
		return err
	}

	// Clone for editing
	tmp := *inst
	inst = &tmp
	restartController := false

	// Update variables
	varsChanged := !maps.Equal(inst.Variables, cfg.Variables)
	if varsChanged {
		inst.Variables = cfg.Variables
		restartController = true
	}
	inst.Annotations = cfg.Annotations
	inst.FrontendURL = cfg.FrontendURL

	// Force the repo to refresh its handshake if the deployment has changed
	updatedOn, ok := r.updatedOn[instanceID]
	if !ok || cfg.UpdatedOn.After(updatedOn) {
		repo, release, err := r.rt.Repo(ctx, inst.ID)
		if err != nil {
			return err
		}
		defer release()

		err = repo.Pull(ctx, &drivers.PullOptions{ForceHandshake: true})
		if err != nil {
			r.rt.Logger.Error("ReloadConfig: failed to pull repo", zap.String("instance_id", inst.ID), zap.Error(err), observability.ZapCtx(ctx))
		}

		// Update the last updatedOn time
		r.updatedOn[instanceID] = cfg.UpdatedOn
		// changes in archive asset IDs are correctly propogated via repo connection reopen only
		restartController = restartController || cfg.UsesArchive
	}

	err = r.rt.EditInstance(ctx, inst, restartController)
	if err != nil {
		return err
	}
	return nil
}

func (r *configReloader) periodicallyReloadConfigs(ctx context.Context) {
	reloadAllInstances := func() {
		r.rt.Logger.Info("periodicallyReloadConfigs: reloading configs for all instances")
		instances, err := r.rt.Instances(ctx)
		if err != nil {
			r.rt.Logger.Error("periodicallyReloadConfigs: failed to list instances", zap.Error(err))
			return
		}
		for _, inst := range instances {
			err := r.reloadConfig(ctx, inst.ID)
			if err != nil {
				r.rt.Logger.Error("periodicallyReloadConfigs: failed to reload config", zap.String("instance_id", inst.ID), zap.Error(err))
			}
		}
	}
	// first reload immediately
	reloadAllInstances()

	// then periodically every hour
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			reloadAllInstances()
		case <-ctx.Done():
			return
		}
	}
}

func (r *configReloader) close() {
	r.cancel()
}
