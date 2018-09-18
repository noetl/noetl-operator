package noetlspark

import (
	"context"
	"log"
	"reflect"

	computev1beta1 "deloo/pkg/apis/compute/v1beta1"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// Add creates a new NoetlSpark Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
// USER ACTION REQUIRED: update cmd/manager/main.go to call this compute.Add(mgr) to install this Controller
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileNoetlSpark{Client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("noetlspark-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to NoetlSpark
	err = c.Watch(&source.Kind{Type: &computev1beta1.NoetlSpark{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Uncomment watch a Job created by NoetlSpark - change this for objects you create
	err = c.Watch(&source.Kind{Type: &batchv1.Job{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &computev1beta1.NoetlSpark{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileNoetlSpark{}

// ReconcileNoetlSpark reconciles a NoetlSpark object
type ReconcileNoetlSpark struct {
	client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a NoetlSpark object and makes changes based on the state read
// and what is in the NoetlSpark.Spec
// Automatically generate RBAC rules to allow the Controller to read and write Jobs
// +kubebuilder:rbac:groups=apps,resources=jobss,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=compute.noetl.com,resources=noetlsparks,verbs=get;list;watch;create;update;patch;delete
func (r *ReconcileNoetlSpark) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the NoetlSpark instance
	instance := &computev1beta1.NoetlSpark{}
	err := r.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	for _, job := range instance.Spec.Spark {
		if err = deployBatchJob(job, r, instance); err != nil {
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}

func deployBatchJob(job computev1beta1.Job, r *ReconcileNoetlSpark, instance *computev1beta1.NoetlSpark) error {
	batch := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      job.Name + "-job",
			Namespace: instance.Namespace,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"job": job.Name + "-job"}},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    job.Name,
							Image:   "art-hq.intranet.qualys.com:5001/datalake/spark-runner:0.1.2",
							Command: []string{"/opt/spark-runner/main", "spark-submit", "-config", job.Confs[0]},
						},
					},
					RestartPolicy: "Never",
				},
			},
		},
	}
	if err := controllerutil.SetControllerReference(instance, batch, r.scheme); err != nil {
		return err
	}

	// Check if the Job already exists
	found := &batchv1.Job{}
	err := r.Get(context.TODO(), types.NamespacedName{Name: batch.Name, Namespace: batch.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		log.Printf("Creating Batch Job %s/%s\n", batch.Namespace, batch.Name)
		err = r.Create(context.TODO(), batch)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	// Update the found object and write the result back if there are any changes
	if !reflect.DeepEqual(batch.Status.Conditions, found.Status.Conditions) {
		found.Spec = batch.Spec

		if len(found.Status.Conditions) > 0 {
			jobStatus := found.Status.Conditions[0]
			if jobStatus.Type == "Complete" {
				log.Printf("Spark batch job %s has completed.\n", batch.Name)
			}
		}

	}

	return nil
}
