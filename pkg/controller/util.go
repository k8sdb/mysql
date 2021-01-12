package controller

import (
	core "k8s.io/api/core/v1"
)

const (
	dbConditionTypeReady = core.PodConditionType("kubedb.com/Ready")
	dbConditionTypeOffline = "DBConditionTypeIsNotReadyForServerOffline"
	dbConditionTypeOnline = "DBConditionTypeReadyAndServerOnline"
)

// hasCondition returns "true" if the desired condition provided in "condType" is present in the condition list.
// Otherwise, it returns "false".
func hasPodCondition(conditions []core.PodCondition, condType core.PodConditionType) bool {
	for i := range conditions {
		if conditions[i].Type == condType {
			return true
		}
	}
	return false
}

// getPodCondition returns a pointer to the desired condition referred by "condType". Otherwise, it returns nil.
func getPodCondition(conditions []core.PodCondition, condType core.PodConditionType) (int, *core.PodCondition) {
	for i := range conditions {
		c := conditions[i]
		if c.Type == condType {
			return i, &c
		}
	}
	return -1, nil
}

// setPodCondition add/update the desired condition to the condition list. It does nothing if the condition is already in
// its desired state.
func setPodCondition(conditions []core.PodCondition, newCondition core.PodCondition) []core.PodCondition {
	idx, curCond := getPodCondition(conditions, newCondition.Type)
	// The desired conditions is not in the condition list or is not in its desired state.
	// If the current condition status is in its desired state, we have nothing to do. Just return the original condition list.
	// Update it if present in the condition list, or append the new condition if it does not present.
	if curCond == nil || idx == -1 {
		return append(conditions, newCondition)
	} else if curCond.Status == newCondition.Status {
		return conditions
	} else if curCond.Status != newCondition.Status {
		conditions[idx].Status = newCondition.Status
		conditions[idx].LastTransitionTime = newCondition.LastTransitionTime
		conditions[idx].Reason = newCondition.Reason
		conditions[idx].Message = newCondition.Message
	}
	return conditions
}

// RemovePodCondition remove a condition from the condition list referred by "condType" parameter.
func removePodCondition(conditions []core.PodCondition, condType core.PodConditionType) []core.PodCondition {
	idx, _ := getPodCondition(conditions, condType)
	if idx == -1 {
		// The desired condition is not present in the condition list. So, nothing to do.
		return conditions
	}
	return append(conditions[:idx], conditions[idx+1:]...)
}

// IsPodConditionTrue returns "true" if the desired condition is in true state.
// It returns "false" if the desired condition is not in "true" state or is not in the condition list.
func isPodConditionTrue(conditions []core.PodCondition, condType core.PodConditionType) bool {
	for i := range conditions {
		if conditions[i].Type == condType && conditions[i].Status == core.ConditionTrue {
			return true
		}
	}
	return false
}

// IsPodConditionFalse returns "true" if the desired condition is in false state.
// It returns "false" if the desired condition is not in "false" state or is not in the condition list.
func isPodConditionFalse(conditions []core.PodCondition, condType core.PodConditionType) bool {
	for i := range conditions {
		if conditions[i].Type == condType && conditions[i].Status == core.ConditionFalse {
			return true
		}
	}
	return false
}

func hasConditionTypeInReadinessGate(podReadinessGates []core.PodReadinessGate, podConditionType core.PodConditionType) bool {
	for i := range podReadinessGates {
		if podReadinessGates[i].ConditionType == podConditionType {
			return true
		}
	}
	return false
}