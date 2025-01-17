// Copyright 2017 The Kubernetes Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package priorityclass

import (
	"log"

	"github.com/kubernetes/dashboard/src/app/backend/api"
	"github.com/kubernetes/dashboard/src/app/backend/errors"
	"github.com/kubernetes/dashboard/src/app/backend/resource/common"
	"github.com/kubernetes/dashboard/src/app/backend/resource/dataselect"
	scheduling "k8s.io/api/scheduling/v1"
	"k8s.io/client-go/kubernetes"
)

type PriorityClassList struct {
	ListMeta api.ListMeta    `json:"listMeta"`
	Items    []PriorityClass `json:"items"`

	// List of non-critical errors, that occurred during resource retrieval.
	Errors []error `json:"errors"`
}

type PriorityClass struct {
	ObjectMeta api.ObjectMeta `json:"objectMeta"`
	TypeMeta   api.TypeMeta   `json:"typeMeta"`
}

func GetPriorityClassList(client kubernetes.Interface, dsQuery *dataselect.DataSelectQuery) (*PriorityClassList, error) {
	log.Println("Getting list of priority class")
	channels := &common.ResourceChannels{
		PriorityClassList: common.GetPriorityClassListChannel(client, 1),
	}

	return GetPriorityClassListFromChannels(channels, dsQuery)
}

func GetPriorityClassListFromChannels(channels *common.ResourceChannels, dsQuery *dataselect.DataSelectQuery) (*PriorityClassList, error) {
	priorityClasses := <-channels.PriorityClassList.List
	err := <-channels.PriorityClassList.Error
	nonCriticalErrors, criticalError := errors.HandleError(err)
	if criticalError != nil {
		return nil, criticalError
	}

	result := toPriorityClassLists(priorityClasses.Items, nonCriticalErrors, dsQuery)
	return result, nil
}

func toPriorityClass(priorityclass scheduling.PriorityClass) PriorityClass {
	return PriorityClass{
		ObjectMeta: api.NewObjectMeta(priorityclass.ObjectMeta),
		TypeMeta:   api.NewTypeMeta(api.ResourceKindPriorityClass),
	}
}

// toPriorityClassLists merges a list of PriorityClass with a list of PriorityClass to create a simpler, unified list
func toPriorityClassLists(priorityClasses []scheduling.PriorityClass, nonCriticalErrors []error,
	dsQuery *dataselect.DataSelectQuery) *PriorityClassList {
	result := &PriorityClassList{
		ListMeta: api.ListMeta{TotalItems: len(priorityClasses)},
		Errors:   nonCriticalErrors,
	}

	items := make([]PriorityClass, 0)
	for _, item := range priorityClasses {
		items = append(items, toPriorityClass(item))
	}

	roleCells, filteredTotal := dataselect.GenericDataSelectWithFilter(toCells(items), dsQuery)
	result.ListMeta = api.ListMeta{TotalItems: filteredTotal}
	result.Items = fromCells(roleCells)
	return result
}
