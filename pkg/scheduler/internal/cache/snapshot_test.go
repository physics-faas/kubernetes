/*
Copyright 2018 The Kubernetes Authors.

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

package cache

import (
	"fmt"
	"reflect"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)

const mb int64 = 1024 * 1024

func TestGetNodeImageStates(t *testing.T) {
	tests := []struct {
		node               *v1.Node
		imageExistenceMap  map[string]sets.String
		layersExistenceMap map[string]sets.String
		expected           map[string]*framework.ImageStateSummary
	}{
		{
			node: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{Name: "node-0"},
				Status: v1.NodeStatus{
					Images: []v1.ContainerImage{
						{
							Names: []string{
								"gcr.io/10:v1",
							},
							SizeBytes: int64(10 * mb),
							Layers: map[string]int64{
								"layer1": int64(10 * mb),
							},
						},
						{
							Names: []string{
								"gcr.io/200:v1",
							},
							SizeBytes: int64(200 * mb),
							Layers: map[string]int64{
								"layer2": int64(200 * mb),
							},
						},
					},
				},
			},
			imageExistenceMap: map[string]sets.String{
				"gcr.io/10:v1":  sets.NewString("node-0", "node-1"),
				"gcr.io/200:v1": sets.NewString("node-0"),
			},
			layersExistenceMap: map[string]sets.String{
				"layer1": sets.NewString("node-1"),
				"layer2": sets.NewString(),
			},
			expected: map[string]*framework.ImageStateSummary{
				"gcr.io/10:v1": {
					Size:     int64(10 * mb),
					NumNodes: 2,
					LayersOnNodes: map[string]sets.String{
						"layer1": sets.NewString("node-1"),
					},
					LayersSize: map[string]int64{
						"layer1": int64(10 * mb),
					},
				},
				"gcr.io/200:v1": {
					Size:     int64(200 * mb),
					NumNodes: 1,
					LayersOnNodes: map[string]sets.String{
						"layer2": sets.NewString(),
					},
					LayersSize: map[string]int64{
						"layer2": int64(200 * mb),
					},
				},
			},
		},
		{
			node: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{Name: "node-0"},
				Status:     v1.NodeStatus{},
			},
			imageExistenceMap: map[string]sets.String{
				"gcr.io/10:v1":  sets.NewString("node-1"),
				"gcr.io/200:v1": sets.NewString(),
			},
			layersExistenceMap: map[string]sets.String{
				"layer1": sets.NewString("node-1"),
				"layer2": sets.NewString(),
			},
			expected: map[string]*framework.ImageStateSummary{},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("case_%d", i), func(t *testing.T) {
			imageStates := getNodeImageStates(test.node, test.imageExistenceMap, test.layersExistenceMap)
			if !reflect.DeepEqual(test.expected, imageStates) {
				t.Errorf("expected: %#v, got: %#v", test.expected, imageStates)
			}
		})
	}
}

func TestCreateImageExistenceMap(t *testing.T) {
	tests := []struct {
		nodes    []*v1.Node
		expected map[string]sets.String
	}{
		{
			nodes: []*v1.Node{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "node-0"},
					Status: v1.NodeStatus{
						Images: []v1.ContainerImage{
							{
								Names: []string{
									"gcr.io/10:v1",
								},
								SizeBytes: int64(10 * mb),
							},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "node-1"},
					Status: v1.NodeStatus{
						Images: []v1.ContainerImage{
							{
								Names: []string{
									"gcr.io/10:v1",
								},
								SizeBytes: int64(10 * mb),
							},
							{
								Names: []string{
									"gcr.io/200:v1",
								},
								SizeBytes: int64(200 * mb),
							},
						},
					},
				},
			},
			expected: map[string]sets.String{
				"gcr.io/10:v1":  sets.NewString("node-0", "node-1"),
				"gcr.io/200:v1": sets.NewString("node-1"),
			},
		},
		{
			nodes: []*v1.Node{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "node-0"},
					Status:     v1.NodeStatus{},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "node-1"},
					Status: v1.NodeStatus{
						Images: []v1.ContainerImage{
							{
								Names: []string{
									"gcr.io/10:v1",
								},
								SizeBytes: int64(10 * mb),
							},
							{
								Names: []string{
									"gcr.io/200:v1",
								},
								SizeBytes: int64(200 * mb),
							},
						},
					},
				},
			},
			expected: map[string]sets.String{
				"gcr.io/10:v1":  sets.NewString("node-1"),
				"gcr.io/200:v1": sets.NewString("node-1"),
			},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("case_%d", i), func(t *testing.T) {
			imageMap := createImageExistenceMap(test.nodes)
			if !reflect.DeepEqual(test.expected, imageMap) {
				t.Errorf("expected: %#v, got: %#v", test.expected, imageMap)
			}
		})
	}
}
