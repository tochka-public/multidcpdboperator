package webhook

// Admission Webhook for Core Types
// https://book.kubebuilder.io/reference/webhook-for-core-types.html

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	admissionv1 "k8s.io/api/admission/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/klog/v2"
	multidccrd "k8s.tochka.com/multidc-pdb-operator/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/cluster"

	"net/http"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	//"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// podValidator validates Pods
type PodValidator struct {
	Client   client.Client
	decoder  *admission.Decoder
	Clusters map[string]*cluster.Cluster
}

func (v *PodValidator) getMetaFromV1Eviction(req admission.Request) (metav1.ObjectMeta, error) {
	eviction := &policyv1.Eviction{}
	err := v.decoder.DecodeRaw(req.Object, eviction)
	if err != nil {
		return metav1.ObjectMeta{}, fmt.Errorf("v1 decoder.DecodeRaw: %w", err)
	}

	return eviction.ObjectMeta, nil
}

func (v *PodValidator) getMetaFromV1Beta1Eviction(req admission.Request) (metav1.ObjectMeta, error) {
	eviction := &policyv1beta1.Eviction{}
	err := v.decoder.DecodeRaw(req.Object, eviction)
	if err != nil {
		return metav1.ObjectMeta{}, fmt.Errorf("v1beta1 decoder.DecodeRaw: %w", err)
	}

	return eviction.ObjectMeta, nil
}

// podValidator admits a pod if a specific annotation exists.
// +kubebuilder:webhook:path=/validate-v1-pod,failurePolicy=fail,groups="",resources=pods,verbs=delete,versions=v1,name=mpod.kb.io
func (v *PodValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	if req.Kind.Kind != "Eviction" || req.Operation != admissionv1.Create || req.Resource.Resource != "pods" {
		req.Object.Raw = []byte{}
		klog.Warningf("podValidator req: %+v", req)
		return admission.Allowed("Bad request")
	}

	var metaObject metav1.ObjectMeta
	var metaErr error
	switch req.RequestKind.Version {
	case policyv1.SchemeGroupVersion.Version:
		metaObject, metaErr = v.getMetaFromV1Eviction(req)
	case policyv1beta1.SchemeGroupVersion.Version:
		metaObject, metaErr = v.getMetaFromV1Beta1Eviction(req)
	default:
		req.Object.Raw = []byte{}
		klog.Warningf("eviction.RequestKind %s not supported: %+v", req.RequestKind.Version, req)
		return admission.Errored(http.StatusBadRequest, fmt.Errorf("request kind of version %s not supported", req.RequestKind.Version))
	}
	if metaErr != nil {
		req.Object.Raw = []byte{}
		klog.Error(metaErr, fmt.Sprintf("req: %+v", req))
		return admission.Errored(http.StatusBadRequest, metaErr)
	}

	// klog.Infof("podValidator eviction: %v+", eviction)
	namespace := metaObject.Namespace
	podName := metaObject.Name
	podEviction := &corev1.Pod{}
	if err := v.Client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: podName}, podEviction); err != nil {
		klog.Error(err)
		return admission.Denied(err.Error())
	}

	multidcPdbList := &multidccrd.MultidcPodDisruptionBudgetList{}
	if err := v.Client.List(ctx, multidcPdbList,
		&client.ListOptions{Namespace: req.Namespace}); err != nil {
		klog.Error(err)
		return admission.Denied(err.Error())
	}

	for _, multidcPdb := range multidcPdbList.Items {
		selector := labels.NewSelector()
		for pdbKey, pdbVal := range multidcPdb.Spec.Selector {
			selectorReq, err := labels.NewRequirement(pdbKey, selection.Equals, []string{pdbVal})
			if err != nil {
				klog.Error(err)
				return admission.Denied(err.Error())
			}
			klog.V(4).Infof("add key %s to selector req %+v for %s", pdbKey, selectorReq, multidcPdb.Name)
			selector = selector.Add(*selectorReq)
		}

		if selector.Empty() {
			klog.Warningf("empty selector %v+ for %s: %+v", selector, multidcPdb.Name, multidcPdb.Spec)
			continue
		}
		podEvictionLabels := labels.Set(podEviction.Labels)
		if selector.Matches(podEvictionLabels) {
			if err := v.processPdb(ctx, &multidcPdb.Spec, podEviction); err != nil {
				klog.Error(fmt.Errorf("failed multidcPdb %s pdbSelector %s podEvictionLabels %s processPdb: %s", multidcPdb.Name, selector.String(), podEvictionLabels.String(), err))
				return admission.Denied(err.Error())
			}
		}
	}

	return admission.Allowed(fmt.Sprintf("%s: Ok", podName))
}

func (v *PodValidator) getDeploymentName(ctx context.Context, podName string, namespace string, replicaSetName string) (string, error) {
	replicaSetObj := &appsv1.ReplicaSet{}
	if err := v.Client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: replicaSetName}, replicaSetObj); err != nil {
		klog.Error(err)
		return "", fmt.Errorf("%s: missing ReplicaSet %s", podName, replicaSetName)
	}
	deploymentObj := &appsv1.Deployment{}
	for _, ownerReference := range replicaSetObj.OwnerReferences {
		if ownerReference.Kind == "Deployment" {
			if err := v.Client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: ownerReference.Name}, deploymentObj); err != nil {
				klog.Error(err)
				return "", fmt.Errorf("%s: missing Deployment %s", podName, ownerReference.Name)
			}
			return deploymentObj.Name, nil
		}
	}
	return "", fmt.Errorf("%s: missing Deployment for ReplicaSet %s", podName, replicaSetName)
}

func (v *PodValidator) processPdb(ctx context.Context, pdbSpec *multidccrd.MultidcPodDisruptionBudgetSpec, pod *corev1.Pod) error {
	minAvailable := 0
	var err error
	if pdbSpec.MinAvailable != "" {
		minAvailable, err = strconv.Atoi(pdbSpec.MinAvailable)
		if err != nil {
			return fmt.Errorf("can`t parse MinAvailable %s", err)
		}
	}

	if !podIsReady(pod) {
		return nil
	}

	evictionSelector := labels.Set(pdbSpec.Selector)

	if minAvailable > 0 {
		readyReplicas := int32(0)
		for k, c := range v.Clusters {
			listObj := &corev1.PodList{}
			if err := (*c).GetClient().List(ctx, listObj,
				&client.ListOptions{Namespace: pod.Namespace, LabelSelector: evictionSelector.AsSelector()}); err != nil {
				klog.Error(err)
				return fmt.Errorf("missing pods %s on %s", evictionSelector.String(), k)
			}
			for _, itemPod := range listObj.Items {
				if podIsReady(&itemPod) {
					readyReplicas += 1
				}
			}
			if readyReplicas > int32(minAvailable) {
				break
			}
		}
		if readyReplicas <= int32(minAvailable) {
			return fmt.Errorf("for %s: minAvailable %d, got %d available", pod.Name, minAvailable, readyReplicas)
		}
	}

	maxUnavailable := 0
	if pdbSpec.MaxUnavailable != "" {
		maxUnavailable, err = strconv.Atoi(pdbSpec.MaxUnavailable)
		if err != nil {
			return fmt.Errorf("can`t parse MaxUnavailable %s", err)
		}
	}

	if maxUnavailable > 0 {
		var unavailableReplicas int32
		var ownerRefKind string
		var ownerRefName string
		var ownerRefDeploymentName string

		for _, ownerRef := range pod.OwnerReferences {
			if strings.HasSuffix(ownerRef.Kind, "Set") {
				ownerRefKind = ownerRef.Kind
				ownerRefName = ownerRef.Name
			}
		}

		for k, c := range v.Clusters {
			if ownerRefKind == "ReplicaSet" {
				if ownerRefDeploymentName == "" {
					ownerRefDeploymentName, err = v.getDeploymentName(ctx, pod.Name, pod.Namespace, ownerRefName)
					if err != nil {
						return err
					}
				}
				deploymentObj := &appsv1.Deployment{}
				if err := (*c).GetClient().Get(ctx, types.NamespacedName{Namespace: pod.Namespace, Name: ownerRefDeploymentName}, deploymentObj); err != nil {
					errMsg := fmt.Errorf("%s: missing %s %s on %s err: %v", pod.Name, deploymentObj.Name, ownerRefDeploymentName, k, err)
					klog.Warningf(errMsg.Error())
					if strings.HasSuffix(err.Error(), "not found") {
						continue
					} else {
						return errMsg
					}
				}
				unavailableReplicas += deploymentObj.Status.UnavailableReplicas
			} else if ownerRefKind == "StatefulSet" {
				kindObj := &appsv1.StatefulSet{}
				if err := (*c).GetClient().Get(ctx, types.NamespacedName{Namespace: pod.Namespace, Name: ownerRefName}, kindObj); err != nil {
					errMsg := fmt.Errorf("%s: missing %s %s on %s err: %v", pod.Name, kindObj.Name, ownerRefDeploymentName, k, err)
					klog.Warningf(errMsg.Error())
					if strings.HasSuffix(err.Error(), "not found") {
						continue
					} else {
						return errMsg
					}
				}
				unavailableReplicas += kindObj.Status.CurrentReplicas - kindObj.Status.ReadyReplicas
			}
		}
		if unavailableReplicas >= int32(maxUnavailable) {
			return fmt.Errorf("for %s: maxUnavailable %d, got %d unavailable", pod.Name, maxUnavailable, unavailableReplicas)
		}
	}

	return nil
}

func podIsReady(pod *corev1.Pod) bool {
	for _, containerStatus := range pod.Status.ContainerStatuses {
		if !containerStatus.Ready {
			return false
		}
	}
	return true
}

// podValidator implements admission.DecoderInjector.
// A decoder will be automatically injected.

// InjectDecoder injects the decoder.
func (v *PodValidator) InjectDecoder(d *admission.Decoder) error {
	v.decoder = d
	return nil
}
