/*
Copyright 2016 The Kubernetes Authors.

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

package dockershim

import (
	"fmt"
	"strings"

	dockertypes "github.com/docker/engine-api/types"

	runtimeApi "k8s.io/kubernetes/pkg/kubelet/api/v1alpha1/runtime"
)

// This file contains helper functions to convert docker API types to runtime
// API types, or vice versa.

const (
	// Status of a container returned by docker ListContainers
	statusRunningPrefix = "Up"
	statusCreatedPrefix = "Created"
	statusExitedPrefix  = "Exited"
)

func toRuntimeAPIImage(image *dockertypes.Image) (*runtimeApi.Image, error) {
	if image == nil {
		return nil, fmt.Errorf("unable to convert a nil pointer to a runtime API image")
	}

	size := uint64(image.VirtualSize)
	return &runtimeApi.Image{
		Id:          &image.ID,
		RepoTags:    image.RepoTags,
		RepoDigests: image.RepoDigests,
		Size_:       &size,
	}, nil
}

func toRuntimeAPIContainer(c *dockertypes.Container) (*runtimeApi.Container, error) {
	state := toRuntimeAPIContainerState(c.Status)
	metadata, err := parseContainerName(c.Names[0])
	if err != nil {
		return nil, err
	}
	labels, annotations := extractLabels(c.Labels)
	sandboxID := c.Labels[sandboxIDLabelKey]
	return &runtimeApi.Container{
		Id:           &c.ID,
		PodSandboxId: &sandboxID,
		Metadata:     metadata,
		Image:        &runtimeApi.ImageSpec{Image: &c.Image},
		ImageRef:     &c.ImageID,
		State:        &state,
		Labels:       labels,
		Annotations:  annotations,
	}, nil
}

func toDockerContainerStatus(state runtimeApi.ContainerState) string {
	switch state {
	case runtimeApi.ContainerState_CREATED:
		return "created"
	case runtimeApi.ContainerState_RUNNING:
		return "running"
	case runtimeApi.ContainerState_EXITED:
		return "exited"
	case runtimeApi.ContainerState_UNKNOWN:
		fallthrough
	default:
		return "unknown"
	}
}

func toRuntimeAPIContainerState(state string) runtimeApi.ContainerState {
	// Parse the state string in dockertypes.Container. This could break when
	// we upgrade docker.
	switch {
	case strings.HasPrefix(state, statusRunningPrefix):
		return runtimeApi.ContainerState_RUNNING
	case strings.HasPrefix(state, statusExitedPrefix):
		return runtimeApi.ContainerState_EXITED
	case strings.HasPrefix(state, statusCreatedPrefix):
		return runtimeApi.ContainerState_CREATED
	default:
		return runtimeApi.ContainerState_UNKNOWN
	}
}

func toRuntimeAPISandboxState(state string) runtimeApi.PodSandBoxState {
	// Parse the state string in dockertypes.Container. This could break when
	// we upgrade docker.
	switch {
	case strings.HasPrefix(state, statusRunningPrefix):
		return runtimeApi.PodSandBoxState_READY
	default:
		return runtimeApi.PodSandBoxState_NOTREADY
	}
}

func toRuntimeAPISandbox(c *dockertypes.Container) (*runtimeApi.PodSandbox, error) {
	state := toRuntimeAPISandboxState(c.Status)
	metadata, err := parseSandboxName(c.Names[0])
	if err != nil {
		return nil, err
	}
	labels, annotations := extractLabels(c.Labels)
	return &runtimeApi.PodSandbox{
		Id:          &c.ID,
		Metadata:    metadata,
		State:       &state,
		CreatedAt:   &c.Created,
		Labels:      labels,
		Annotations: annotations,
	}, nil
}
