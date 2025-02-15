package cluster

// Copyright (c) Microsoft Corporation.
// Licensed under the Apache License 2.0.

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"

	mock_metrics "github.com/Azure/ARO-RP/pkg/util/mocks/metrics"
)

func TestEmitDebugPodsCount(t *testing.T) {
	ctx := context.Background()

	cli := fake.NewSimpleClientset(
		&corev1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "master-1",
			},
		},
		&corev1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "master-2",
			},
		},
	)

	cli.PrependReactor("list", "events", reactorFn)

	controller := gomock.NewController(t)
	defer controller.Finish()

	m := mock_metrics.NewMockEmitter(controller)
	mon := &Monitor{
		cli: cli,
		m:   m,
	}

	m.EXPECT().EmitGauge("debugpods.count", int64(1), map[string]string{})
	err := mon.emitDebugPodsCount(ctx)
	if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}
}

func reactorFn(_ ktesting.Action) (handled bool, ret kruntime.Object, err error) {
	now := metav1.Now()
	longAgo := metav1.Date(1991, time.August, 24, 0, 0, 0, 0, time.UTC)
	return true, &corev1.EventList{Items: []corev1.Event{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "master-1-started",
				Namespace: "default",
			},
			Reason: "Started",
			InvolvedObject: corev1.ObjectReference{
				Kind: "Pod",
				Name: "master-1-debug",
			},
			LastTimestamp: longAgo,
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "master-2-started",
				Namespace: "default",
			},
			Reason: "Started",
			InvolvedObject: corev1.ObjectReference{
				Kind: "Pod",
				Name: "master-2-debug",
			},
			LastTimestamp: now,
		},
	}}, nil
}

func TestEventIsNew(t *testing.T) {
	for _, tt := range []struct {
		event corev1.Event
		isNew bool
	}{
		{
			event: corev1.Event{
				LastTimestamp: metav1.Now(),
			},
			isNew: true,
		},
		{
			event: corev1.Event{
				LastTimestamp: metav1.NewTime(metav1.Now().Add(-1 * time.Minute)),
			},
			isNew: true,
		},
		{
			event: corev1.Event{
				LastTimestamp: metav1.Date(2020, 02, 18, 0, 0, 0, 0, time.UTC),
			},
			isNew: false,
		},
	} {
		if tt.isNew != eventIsNew(tt.event) {
			t.Fatalf("test failed for the event: %v", tt.event)
		}
	}
}

func TestGetDebugPodNames(t *testing.T) {
	for _, tt := range []struct {
		node                 corev1.Node
		expectedDebugPodName string
	}{
		{
			node: corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "master-0",
				},
			},
			expectedDebugPodName: "master-0-debug",
		},
		{
			node: corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "-master-",
				},
			},
			expectedDebugPodName: "master-debug",
		},
	} {
		if tt.expectedDebugPodName != getDebugPodNames([]corev1.Node{tt.node})[0] {
			t.Fatalf("test failed for the node: %v", tt.node.Name)
		}
	}
}

func TestCountDebugPods(t *testing.T) {
	now := metav1.Now()
	longAgo := metav1.Date(2020, 02, 18, 0, 0, 0, 0, time.UTC)

	for _, tt := range []struct {
		debugPodNames []string
		events        []corev1.Event
		expectedCount int
	}{
		{
			debugPodNames: []string{"m-0-debug", "m-1-debug", "m-2-debug", "m-3-debug"},
			events: []corev1.Event{
				{
					LastTimestamp: now,
					InvolvedObject: corev1.ObjectReference{
						Name: "m-0-debug",
					},
				},
				{
					LastTimestamp: now,
					InvolvedObject: corev1.ObjectReference{
						Name: "m-3",
					},
				},
				{
					LastTimestamp: longAgo,
					InvolvedObject: corev1.ObjectReference{
						Name: "m-1-debug",
					},
				},
			},
			expectedCount: 1,
		},
	} {
		if countDebugPods(tt.debugPodNames, tt.events) != tt.expectedCount {
			t.Fatalf("test failed for the set: %v", tt.debugPodNames)
		}
	}
}
