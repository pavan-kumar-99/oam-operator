/*
Copyright 2021 Pavan.

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

package controllers

import (
	"context"
	"fmt"
	"strconv"
	"time"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	autoscale "k8s.io/api/autoscaling/v1"
	apiv1 "k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	networkv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	appsv1beta1 "oam-operator/api/v1beta1"
	aws "oam-operator/cloudprovider/aws"
)

var (
	successCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name:      "success_count_total",
		Help:      "Counter of success actions made.",
	})

)

// ApplicationReconciler reconciles a Application object
type ApplicationReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=apps.oam.cfcn.io,resources=applications,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps.oam.cfcn.io,resources=applications/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps.oam.cfcn.io,resources=applications/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Application object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *ApplicationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logs := log.FromContext(ctx)
	var app appsv1beta1.Application
	var hpa autoscale.HorizontalPodAutoscaler
	var svc corev1.Service
	var deploy appsv1.Deployment
	var ing networkv1.Ingress
	if err := r.Get(ctx, req.NamespacedName, &app); err != nil {
		logs.Info("Application Not found", "Error", err)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	finalizers := app.GetFinalizers()
	if app.GetDeletionTimestamp().IsZero() {
		if !containsString(finalizers, "finalizer.app") {
			app.SetFinalizers(append(finalizers, "finalizer.app"))
			fmt.Println("Finalizer Added")
			if err := r.Update(ctx, &app); err != nil {
				return ctrl.Result{}, fmt.Errorf("failed executing finalizer: %w", err)
			}
		}
	} else {

		if containsString(finalizers, "finalizer.app") {
			custom := types.NamespacedName{
				Name:      req.Name,
				Namespace: req.Namespace,
			}
			if err := r.Get(ctx, custom, &hpa); err == nil {
				logs.Info("Found HPA", "HPA Details", custom)
				requestCount.Inc()
				if err := r.Delete(ctx, &hpa); err != nil {
					return ctrl.Result{}, fmt.Errorf("Unable to Delete HPA: %w", err)
				}
			}
			if err := r.Get(ctx, custom, &svc); err == nil {
				logs.Info("Found Service", "Service Details", custom)
				requestCount.Inc()
				if err := r.Delete(ctx, &svc); err != nil {
					return ctrl.Result{}, fmt.Errorf("Unable to Delete Service: %w", err)
				}
			}
			if err := r.Get(ctx, custom, &deploy); err == nil {
				logs.Info("Found Deployment", "Deployment Details", custom)
				requestCount.Inc()
				if err := r.Delete(ctx, &deploy); err != nil {
					return ctrl.Result{}, fmt.Errorf("Unable to Delete Deployment: %w", err)
				}
			}
			if err := r.Get(ctx, custom, &ing); err == nil {
				logs.Info("Found Ingress", "Ingress Details", custom)
				requestCount.Inc()
				if err := r.Delete(ctx, &ing); err != nil {
					return ctrl.Result{}, fmt.Errorf("Unable to Delete Ingress: %w", err)
				}
			}
			shouldCreateS3, err := strconv.ParseBool(app.Spec.Cloud.Aws.S3)
			if err != nil {
				shouldCreateS3 = false
			}
			if shouldCreateS3 {
				logs.Info("S3 is enabled")
				aws.DeleteS3(req.Name + req.Namespace)
			}
			logs.Info("Executing finalizer in app")

			app.SetFinalizers(removeString(finalizers, "finalizer.app"))
			if err := r.Update(ctx, &app); err != nil {
				return ctrl.Result{}, fmt.Errorf("unable to remove finalizer from obj: %w", err)
			}
		}
		return ctrl.Result{}, nil
	}
	shouldCreateS3, err := strconv.ParseBool(app.Spec.Cloud.Aws.S3)
	if err != nil {
		shouldCreateS3 = false
	}
	if aws.ListS3(req.Name + req.Namespace) {
		logs.Info("Bucket already exists")
	} else {
		if shouldCreateS3 {
			logs.Info("S3 is enabled")
			aws.CreateS3(req.Name + req.Namespace)
			app.Status.S3BucketName = req.Name + req.Namespace
			if err := r.Status().Update(ctx, &app); err != nil {
				logs.Error(err, "unable to update App status")
				return ctrl.Result{}, err
			}
			r.recorder.Event(&app, corev1.EventTypeNormal, "Created", "Created S3 Bucket")
		}
	}

	if err := r.Get(ctx, req.NamespacedName, &hpa); err == nil {
		logs.Info("Found HPA", "HPA Details", req.NamespacedName)
		requestCount.Inc()
	} else {
		r.CreateHpa(ctx, req, app, logs)
	}
	if err := r.Get(ctx, req.NamespacedName, &svc); err == nil {
		logs.Info("Found Service", "Service Details", req.NamespacedName)
		requestCount.Inc()
	} else {
		r.CreateService(ctx, req, app, logs)
	}
	if err := r.Get(ctx, req.NamespacedName, &deploy); err == nil {
		logs.Info("Found Deployment", "Deployment Details", req.NamespacedName)
		requestCount.Inc()
	} else {
		r.CreateDeploy(ctx, req, app, logs)
	}
	if err := r.Get(ctx, req.NamespacedName, &ing); err == nil {
		logs.Info("Found Ingress", "Ingress Details", req.NamespacedName)
		requestCount.Inc()
	} else {
		r.CreateIngress(ctx, req, app, logs)
	}
	return ctrl.Result{}, nil
}

func (r *ApplicationReconciler) CreateHpa(ctx context.Context, req ctrl.Request, app appsv1beta1.Application, log logr.Logger) (ctrl.Result, error) {
	var min int32 = 1
	var max int32 = 2
	var targetCPU int32 = 1
	hpa := &autoscale.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{Name: req.Name, Namespace: req.Namespace},
		Spec: autoscale.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autoscale.CrossVersionObjectReference{
				Kind:       "Deployment",
				Name:       req.Name,
				APIVersion: "apps/v1",
			},
			MinReplicas:                    &min,
			MaxReplicas:                    max,
			TargetCPUUtilizationPercentage: &targetCPU,
		},
	}
	if err := r.Create(ctx, hpa); err != nil {
		log.Error(err, "unable to create HPA for Application", "Application", hpa)
		return ctrl.Result{}, err
	}
	r.recorder.Event(&app, corev1.EventTypeNormal, "Created", "Created HPA")
	return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
}

func (r *ApplicationReconciler) CreateIngress(ctx context.Context, req ctrl.Request, app appsv1beta1.Application, log logr.Logger) (ctrl.Result, error) {
	pathType := "Exact"
	ing := &networkv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name,
			Namespace: req.Namespace,
		},
		Spec: networkv1.IngressSpec{
			DefaultBackend: &networkv1.IngressBackend{
				Service: &networkv1.IngressServiceBackend{
					Name: req.Name,
					Port: networkv1.ServiceBackendPort{
						Name: "http",
					},
				},
			},
			Rules: []networkv1.IngressRule{
				{
					IngressRuleValue: networkv1.IngressRuleValue{
						HTTP: &networkv1.HTTPIngressRuleValue{
							Paths: []networkv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: (*networkv1.PathType)(&pathType),
									Backend: networkv1.IngressBackend{
										Service: &networkv1.IngressServiceBackend{
											Name: req.Name,
											Port: networkv1.ServiceBackendPort{
												Name: "http",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	if err := r.Create(ctx, ing); err != nil {
		log.Error(err, "unable to create Ingress for Application", "Application", ing)
		return ctrl.Result{}, err
	}
	r.recorder.Event(&app, corev1.EventTypeNormal, "Created", "Created Ingress")
	return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
}

func (r *ApplicationReconciler) CreateDeploy(ctx context.Context, req ctrl.Request, app appsv1beta1.Application, log logr.Logger) (ctrl.Result, error) {
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name,
			Namespace: req.Namespace,
			Annotations: map[string]string{
				"kv-inject":  "true",
				"init-only":  "false",
				"consul-url": "https://consul-prod.pavan.com",
				"Port":       "9200",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(2),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": req.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": req.Name,
					},
				},
				Spec: corev1.PodSpec{
					InitContainers: []corev1.Container{
						{
							Name:    "echo-init-1",
							Image:   "busybox",
							Command: []string{"echo", "In Init Container"},
							Resources: corev1.ResourceRequirements{
								Limits: map[corev1.ResourceName]resource.Quantity{
									"cpu":    resource.MustParse("1"),
									"memory": resource.MustParse("1Gi"),
								},
								Requests: map[corev1.ResourceName]resource.Quantity{
									"cpu":    resource.MustParse("100m"),
									"memory": resource.MustParse("90Mi"),
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:  "main-container",
							Image: "nginx",
							Resources: corev1.ResourceRequirements{
								Limits: map[corev1.ResourceName]resource.Quantity{
									"cpu":    resource.MustParse("1"),
									"memory": resource.MustParse("1Gi"),
								},
								Requests: map[corev1.ResourceName]resource.Quantity{
									"cpu":    resource.MustParse("100m"),
									"memory": resource.MustParse("90Mi"),
								},
							},
						},
					},
				},
			},
		},
	}
	if err := r.Create(ctx, deployment); err != nil {
		log.Error(err, "unable to create Deployment for Application", "Application", deployment)
		return ctrl.Result{}, err
	}
	r.recorder.Event(&app, corev1.EventTypeNormal, "Created", "Created Deployment")
	return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
}

func int32Ptr(i int32) *int32 { return &i }

func (r *ApplicationReconciler) CreateService(ctx context.Context, req ctrl.Request, app appsv1beta1.Application, log logr.Logger) (ctrl.Result, error) {
	svc := &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name,
			Namespace: req.Namespace,
		},
		Spec: apiv1.ServiceSpec{
			Ports: []apiv1.ServicePort{
				{
					Name: "http",
					Port: int32(80),
					TargetPort: intstr.IntOrString{
						IntVal: 80,
					},
				},
			},
			Selector: map[string]string{
				"app": req.Name,
			},
			Type: "ClusterIP",
		},
	}
	if err := r.Create(ctx, svc); err != nil {
		log.Error(err, "unable to create Service for Application", "Application", svc)
		return ctrl.Result{}, err
	}
	log.Info("created Service for Application", "Application", svc)
	r.recorder.Event(&app, corev1.EventTypeNormal, "Created", "Created Service")
	return ctrl.Result{RequeueAfter: 10 * time.Second}, nil

}

func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}

// SetupWithManager sets up the controller with the Manager.
func (r *ApplicationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	prometheus.MustRegister(successCount)
	r.recorder = mgr.GetEventRecorderFor("Application")
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1beta1.Application{}).
		Complete(r)
}
