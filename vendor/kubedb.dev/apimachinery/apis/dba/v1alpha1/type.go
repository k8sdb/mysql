/*
Copyright The KubeDB Authors.

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

package v1alpha1

type RequestConditionType string

// These are the possible conditions for a certificate request.
const (
	AccessApproved         RequestConditionType = "Approved"
	AccessDenied           RequestConditionType = "Denied"
	Processing             RequestConditionType = "Processing"
	PauseDatabase          RequestConditionType = "PauseDatabase"
	ResumeDatabase         RequestConditionType = "ResumeDatabase"
	Successful             RequestConditionType = "Successful"
	Failed                 RequestConditionType = "Failed"
	UpgradeDatabaseVersion RequestConditionType = "UpgradeDatabaseVersion"
	ScaleDatabase          RequestConditionType = "ScaleDatabase"
)

type ModificationRequestPhase string

const (
	// used for modification requests that are currently processing
	ModificationRequestPhaseProcessing ModificationRequestPhase = "Processing"
	// used for modification requests that are executed successfully
	ModificationRequestPhaseSuccessful ModificationRequestPhase = "Successful"
	// used for modification requests that are waiting for approval
	ModificationRequestPhaseWaitingForApproval ModificationRequestPhase = "WaitingForApproval"
	// used for modification requests that are failed
	ModificationRequestPhaseFailed ModificationRequestPhase = "Failed"
	// used for modification requests that are approved
	ModificationRequestApproved ModificationRequestPhase = "Approved"
	// used for modification requests that are denied
	ModificationRequestDenied ModificationRequestPhase = "Denied"
)

type ModificationRequestType string

const (
	// used for HorizontalScaling operation
	ModificationRequestTypeHorizontalScaling ModificationRequestType = "HorizontalScaling"
	// used for VerticalScaling operation
	ModificationRequestTypeVerticalScaling ModificationRequestType = "VerticalScaling"
	// used for Upgrade operation
	ModificationRequestTypeUpgrade ModificationRequestType = "Upgrade"
	// used for Downgrade operation
	ModificationRequestTypeDowngrade ModificationRequestType = "Downgrade"
)
