package controller

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/crossplane/crossplane-runtime/pkg/errors"
	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type jsonOp string

const (
	opRemove  jsonOp = "remove"
	opAdd     jsonOp = "add"
	opNone    jsonOp = "none"
	opReplace jsonOp = "replace"
)

type jsonpatch struct {
	Op    jsonOp `json:"op,omitempty"`
	Path  string `json:"path,omitempty"`
	Value string `json:"value,omitempty"`
}

func handleDeletionProtection(ctx context.Context, log logr.Logger, baseObj client.Object, inst client.Object, c client.Client, enabled bool, retention int) (time.Duration, error) {

	requeueTime := time.Second * 30

	op := opNone

	if !enabled {
		log.Info("DeletionProtection is not enabled, ensuring no finalizer set", "objectName", inst.GetName())
		removed := controllerutil.RemoveFinalizer(inst, finalizerName)

		if removed {
			op = opRemove
		}
		return requeueTime, patchObjectFinalizer(ctx, log, baseObj, inst, op, c)
	}

	if !controllerutil.ContainsFinalizer(inst, finalizerName) && inst.GetDeletionTimestamp() == nil {
		added := controllerutil.AddFinalizer(inst, finalizerName)
		if added {
			log.Info("Added finalizer to the object", "objectName", inst.GetName())
			op = opAdd
			return requeueTime, patchObjectFinalizer(ctx, log, baseObj, inst, op, c)
		}
	}

	if inst.GetDeletionTimestamp() != nil {
		requeueTime, op = checkRetention(inst, log, retention)
	}

	return requeueTime, patchObjectFinalizer(ctx, log, baseObj, inst, op, c)
}

func checkRetention(inst client.Object, log logr.Logger, retention int) (time.Duration, jsonOp) {
	timestamp := inst.GetDeletionTimestamp()
	expireDate := timestamp.AddDate(0, 0, retention)
	op := opNone
	if time.Now().After(expireDate) {
		log.Info("Retention expired, removing finalizer")
		removed := controllerutil.RemoveFinalizer(inst, finalizerName)
		if removed {
			op = opRemove
		}
	}
	log.V(1).Info("Deletion in: " + expireDate.Sub(time.Now()).String())
	return expireDate.Sub(time.Now()), op
}

func patchObjectFinalizer(ctx context.Context, log logr.Logger, baseObj client.Object, inst client.Object, op jsonOp, c client.Client) error {

	if op == opNone {
		return nil
	}

	// handle the case if crossplane or something else decides to add more finalizers, or if
	// the finalizer is already there.
	index := len(inst.GetFinalizers())
	for i, finalizer := range inst.GetFinalizers() {
		if finalizer == finalizerName {
			index = i
		}
	}

	log.V(1).Info("Index size", "size", index, "found finalzers", inst.GetFinalizers())

	patchOps := []jsonpatch{
		{
			Op:    op,
			Path:  "/metadata/finalizers/" + strconv.Itoa(index),
			Value: finalizerName,
		},
	}

	patch, err := json.Marshal(patchOps)
	if err != nil {
		return errors.Wrap(err, "can't marshal patch")
	}

	log.V(1).Info("Patching object", "patch", string(patch))

	err = c.Patch(ctx, baseObj, client.RawPatch(types.JSONPatchType, patch))
	if err != nil && apierrors.IsNotFound(err) {
		return nil
	} else if err != nil {
		return errors.Wrap(err, "Could not patch object")
	}

	return nil
}
