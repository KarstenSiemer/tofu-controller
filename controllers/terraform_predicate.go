package controllers

import (
	"github.com/fluxcd/pkg/apis/meta"
	"github.com/fluxcd/pkg/runtime/conditions"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	infrav1 "github.com/flux-iac/tofu-controller/api/v1alpha2"
)

// TerraformDependencyReadyTransitionPredicate fires only when a Terraform
// transitions from "not ready" to "ready and converged" — i.e. matches the
// exact condition that checkDependencies treats as a satisfied dependency:
// Generation == ObservedGeneration AND Ready=True.
//
// It is used to enqueue dependents the moment their dependency becomes ready,
// so they no longer have to wait for the next RequeueAfter tick.
type TerraformDependencyReadyTransitionPredicate struct {
	predicate.Funcs
}

func (TerraformDependencyReadyTransitionPredicate) Create(_ event.CreateEvent) bool {
	return false
}

func (TerraformDependencyReadyTransitionPredicate) Delete(_ event.DeleteEvent) bool {
	return false
}

func (TerraformDependencyReadyTransitionPredicate) Generic(_ event.GenericEvent) bool {
	return false
}

func (TerraformDependencyReadyTransitionPredicate) Update(e event.UpdateEvent) bool {
	if e.ObjectOld == nil || e.ObjectNew == nil {
		return false
	}
	oldT, ok := e.ObjectOld.(*infrav1.Terraform)
	if !ok {
		return false
	}
	newT, ok := e.ObjectNew.(*infrav1.Terraform)
	if !ok {
		return false
	}
	return !isReadyAndConverged(oldT) && isReadyAndConverged(newT)
}

func isReadyAndConverged(t *infrav1.Terraform) bool {
	return t.Generation == t.Status.ObservedGeneration && conditions.IsTrue(t, meta.ReadyCondition)
}
