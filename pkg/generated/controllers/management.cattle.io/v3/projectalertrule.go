/*
Copyright 2024 Rancher Labs, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by main. DO NOT EDIT.

package v3

import (
	"context"
	"sync"
	"time"

	v3 "github.com/rancher/rancher/pkg/apis/management.cattle.io/v3"
	"github.com/rancher/wrangler/v2/pkg/apply"
	"github.com/rancher/wrangler/v2/pkg/condition"
	"github.com/rancher/wrangler/v2/pkg/generic"
	"github.com/rancher/wrangler/v2/pkg/kv"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ProjectAlertRuleController interface for managing ProjectAlertRule resources.
type ProjectAlertRuleController interface {
	generic.ControllerInterface[*v3.ProjectAlertRule, *v3.ProjectAlertRuleList]
}

// ProjectAlertRuleClient interface for managing ProjectAlertRule resources in Kubernetes.
type ProjectAlertRuleClient interface {
	generic.ClientInterface[*v3.ProjectAlertRule, *v3.ProjectAlertRuleList]
}

// ProjectAlertRuleCache interface for retrieving ProjectAlertRule resources in memory.
type ProjectAlertRuleCache interface {
	generic.CacheInterface[*v3.ProjectAlertRule]
}

// ProjectAlertRuleStatusHandler is executed for every added or modified ProjectAlertRule. Should return the new status to be updated
type ProjectAlertRuleStatusHandler func(obj *v3.ProjectAlertRule, status v3.AlertStatus) (v3.AlertStatus, error)

// ProjectAlertRuleGeneratingHandler is the top-level handler that is executed for every ProjectAlertRule event. It extends ProjectAlertRuleStatusHandler by a returning a slice of child objects to be passed to apply.Apply
type ProjectAlertRuleGeneratingHandler func(obj *v3.ProjectAlertRule, status v3.AlertStatus) ([]runtime.Object, v3.AlertStatus, error)

// RegisterProjectAlertRuleStatusHandler configures a ProjectAlertRuleController to execute a ProjectAlertRuleStatusHandler for every events observed.
// If a non-empty condition is provided, it will be updated in the status conditions for every handler execution
func RegisterProjectAlertRuleStatusHandler(ctx context.Context, controller ProjectAlertRuleController, condition condition.Cond, name string, handler ProjectAlertRuleStatusHandler) {
	statusHandler := &projectAlertRuleStatusHandler{
		client:    controller,
		condition: condition,
		handler:   handler,
	}
	controller.AddGenericHandler(ctx, name, generic.FromObjectHandlerToHandler(statusHandler.sync))
}

// RegisterProjectAlertRuleGeneratingHandler configures a ProjectAlertRuleController to execute a ProjectAlertRuleGeneratingHandler for every events observed, passing the returned objects to the provided apply.Apply.
// If a non-empty condition is provided, it will be updated in the status conditions for every handler execution
func RegisterProjectAlertRuleGeneratingHandler(ctx context.Context, controller ProjectAlertRuleController, apply apply.Apply,
	condition condition.Cond, name string, handler ProjectAlertRuleGeneratingHandler, opts *generic.GeneratingHandlerOptions) {
	statusHandler := &projectAlertRuleGeneratingHandler{
		ProjectAlertRuleGeneratingHandler: handler,
		apply:                             apply,
		name:                              name,
		gvk:                               controller.GroupVersionKind(),
	}
	if opts != nil {
		statusHandler.opts = *opts
	}
	controller.OnChange(ctx, name, statusHandler.Remove)
	RegisterProjectAlertRuleStatusHandler(ctx, controller, condition, name, statusHandler.Handle)
}

type projectAlertRuleStatusHandler struct {
	client    ProjectAlertRuleClient
	condition condition.Cond
	handler   ProjectAlertRuleStatusHandler
}

// sync is executed on every resource addition or modification. Executes the configured handlers and sends the updated status to the Kubernetes API
func (a *projectAlertRuleStatusHandler) sync(key string, obj *v3.ProjectAlertRule) (*v3.ProjectAlertRule, error) {
	if obj == nil {
		return obj, nil
	}

	origStatus := obj.Status.DeepCopy()
	obj = obj.DeepCopy()
	newStatus, err := a.handler(obj, obj.Status)
	if err != nil {
		// Revert to old status on error
		newStatus = *origStatus.DeepCopy()
	}

	if a.condition != "" {
		if errors.IsConflict(err) {
			a.condition.SetError(&newStatus, "", nil)
		} else {
			a.condition.SetError(&newStatus, "", err)
		}
	}
	if !equality.Semantic.DeepEqual(origStatus, &newStatus) {
		if a.condition != "" {
			// Since status has changed, update the lastUpdatedTime
			a.condition.LastUpdated(&newStatus, time.Now().UTC().Format(time.RFC3339))
		}

		var newErr error
		obj.Status = newStatus
		newObj, newErr := a.client.UpdateStatus(obj)
		if err == nil {
			err = newErr
		}
		if newErr == nil {
			obj = newObj
		}
	}
	return obj, err
}

type projectAlertRuleGeneratingHandler struct {
	ProjectAlertRuleGeneratingHandler
	apply apply.Apply
	opts  generic.GeneratingHandlerOptions
	gvk   schema.GroupVersionKind
	name  string
	seen  sync.Map
}

// Remove handles the observed deletion of a resource, cascade deleting every associated resource previously applied
func (a *projectAlertRuleGeneratingHandler) Remove(key string, obj *v3.ProjectAlertRule) (*v3.ProjectAlertRule, error) {
	if obj != nil {
		return obj, nil
	}

	obj = &v3.ProjectAlertRule{}
	obj.Namespace, obj.Name = kv.RSplit(key, "/")
	obj.SetGroupVersionKind(a.gvk)

	if a.opts.UniqueApplyForResourceVersion {
		a.seen.Delete(key)
	}

	return nil, generic.ConfigureApplyForObject(a.apply, obj, &a.opts).
		WithOwner(obj).
		WithSetID(a.name).
		ApplyObjects()
}

// Handle executes the configured ProjectAlertRuleGeneratingHandler and pass the resulting objects to apply.Apply, finally returning the new status of the resource
func (a *projectAlertRuleGeneratingHandler) Handle(obj *v3.ProjectAlertRule, status v3.AlertStatus) (v3.AlertStatus, error) {
	if !obj.DeletionTimestamp.IsZero() {
		return status, nil
	}

	objs, newStatus, err := a.ProjectAlertRuleGeneratingHandler(obj, status)
	if err != nil {
		return newStatus, err
	}
	if !a.isNewResourceVersion(obj) {
		return newStatus, nil
	}

	err = generic.ConfigureApplyForObject(a.apply, obj, &a.opts).
		WithOwner(obj).
		WithSetID(a.name).
		ApplyObjects(objs...)
	if err != nil {
		return newStatus, err
	}
	a.storeResourceVersion(obj)
	return newStatus, nil
}

// isNewResourceVersion detects if a specific resource version was already successfully processed.
// Only used if UniqueApplyForResourceVersion is set in generic.GeneratingHandlerOptions
func (a *projectAlertRuleGeneratingHandler) isNewResourceVersion(obj *v3.ProjectAlertRule) bool {
	if !a.opts.UniqueApplyForResourceVersion {
		return true
	}

	// Apply once per resource version
	key := obj.Namespace + "/" + obj.Name
	previous, ok := a.seen.Load(key)
	return !ok || previous != obj.ResourceVersion
}

// storeResourceVersion keeps track of the latest resource version of an object for which Apply was executed
// Only used if UniqueApplyForResourceVersion is set in generic.GeneratingHandlerOptions
func (a *projectAlertRuleGeneratingHandler) storeResourceVersion(obj *v3.ProjectAlertRule) {
	if !a.opts.UniqueApplyForResourceVersion {
		return
	}

	key := obj.Namespace + "/" + obj.Name
	a.seen.Store(key, obj.ResourceVersion)
}