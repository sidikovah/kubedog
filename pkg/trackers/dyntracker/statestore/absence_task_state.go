package statestore

import (
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/werf/kubedog/pkg/trackers/dyntracker/util"
)

type AbsenceTaskState struct {
	name             string
	namespace        string
	groupVersionKind schema.GroupVersionKind

	absentConditions  []AbsenceTaskConditionFn
	failureConditions []AbsenceTaskConditionFn

	resourceState *util.Concurrent[*ResourceState]
}

func NewAbsenceTaskState(name, namespace string, groupVersionKind schema.GroupVersionKind, opts AbsenceTaskStateOptions) *AbsenceTaskState {
	resourceState := util.NewConcurrent(NewResourceState(name, namespace, groupVersionKind))

	absentConditions := initAbsenceTaskStateAbsentConditions()
	failureConditions := []AbsenceTaskConditionFn{}

	return &AbsenceTaskState{
		name:              name,
		namespace:         namespace,
		groupVersionKind:  groupVersionKind,
		absentConditions:  absentConditions,
		failureConditions: failureConditions,
		resourceState:     resourceState,
	}
}

type AbsenceTaskStateOptions struct{}

func (s *AbsenceTaskState) Name() string {
	return s.name
}

func (s *AbsenceTaskState) Namespace() string {
	return s.namespace
}

func (s *AbsenceTaskState) GroupVersionKind() schema.GroupVersionKind {
	return s.groupVersionKind
}

func (s *AbsenceTaskState) ResourceState() *util.Concurrent[*ResourceState] {
	return s.resourceState
}

func (s *AbsenceTaskState) Status() AbsenceTaskStatus {
	for _, failureCondition := range s.failureConditions {
		if failureCondition(s) {
			return AbsenceTaskStatusFailed
		}
	}

	for _, absentCondition := range s.absentConditions {
		if !absentCondition(s) {
			return AbsenceTaskStatusProgressing
		}
	}

	return AbsenceTaskStatusAbsent
}

func initAbsenceTaskStateAbsentConditions() []AbsenceTaskConditionFn {
	var absentConditions []AbsenceTaskConditionFn

	absentConditions = append(absentConditions, func(taskState *AbsenceTaskState) bool {
		var absent bool
		taskState.resourceState.RTransaction(func(rs *ResourceState) {
			if rs.Status() == ResourceStatusDeleted {
				absent = true
			}
		})

		return absent
	})

	return absentConditions
}
