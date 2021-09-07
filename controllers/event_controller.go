/*
Copyright 2021.

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

package controllers

import (
	"context"

	"github.com/fluxcd/pkg/runtime/events"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// EventReconciler reconciles a Event object
type EventReconciler struct {
	client.Client
	Scheme                *runtime.Scheme
	ExternalEventRecorder *events.Recorder
}

type EventReconcilerOptions struct {
	MaxConcurrentReconciles int
}

//+kubebuilder:rbac:groups=x.toolkit.fluxcd.io,resources=events,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=x.toolkit.fluxcd.io,resources=events/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=x.toolkit.fluxcd.io,resources=events/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Event object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *EventReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logr.FromContext(ctx)

	var event corev1.Event
	if err := r.Get(ctx, req.NamespacedName, &event); err != nil {
		log.Error(err, "unable to fect Event")

		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	log.Info("Received event",
		"type", event.Type,
		"reason", event.Reason,
		"message", event.Message,
		"involvedObject", event.InvolvedObject,
	)

	if r.ExternalEventRecorder != nil {
		severity := events.EventSeverityInfo
		if event.Type == corev1.EventTypeWarning {
			severity = events.EventSeverityError
		}

		if err := r.ExternalEventRecorder.Eventf(event.InvolvedObject, nil, severity, event.Reason, event.Message); err != nil {
			log.Error(err, "unable to send event")
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *EventReconciler) SetupWithManager(mgr ctrl.Manager, opts EventReconcilerOptions) error {
	return ctrl.NewControllerManagedBy(mgr).
		// Uncomment the following line adding a pointer to an instance of the controlled resource as an argument
		For(&corev1.Event{}).
		WithEventFilter(predicate.Funcs{
			DeleteFunc:  func(de event.DeleteEvent) bool { return false },
			GenericFunc: func(ge event.GenericEvent) bool { return false },
		}).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: opts.MaxConcurrentReconciles,
		}).
		Complete(r)
}