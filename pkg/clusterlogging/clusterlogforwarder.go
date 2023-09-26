package clusterlogging

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/msg"
	clov1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	goclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// ClusterLogForwarderBuilder provides a struct for clusterlogforwarder object from the
// cluster and a clusterlogforwarder definition.
type ClusterLogForwarderBuilder struct {
	// clusterlogforwarder definition, used to create the clusterlogforwarder object.
	Definition *clov1.ClusterLogForwarder
	// Created clusterlogforwarder object.
	Object *clov1.ClusterLogForwarder
	// api client to interact with the cluster.
	apiClient *clients.Settings
	// errorMsg is processed before clusterlogforwarder object is created.
	errorMsg string
}

// PullClusterLogForwarder retrieves an existing clusterlogforwarder object from the cluster.
func PullClusterLogForwarder(apiClient *clients.Settings, name, namespace string) (*clov1.ClusterLogForwarder, error) {
	glog.V(100).Infof("Pulling existing clusterlogforwarder %s in namespace %s", name, namespace)

	builder := ClusterLogForwarderBuilder{
		apiClient: apiClient,
		Definition: &clov1.ClusterLogForwarder{
			ObjectMeta: metaV1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		},
	}

	if name == "" {
		glog.V(100).Infof("The name of the clusterlogforwarder is empty")

		builder.errorMsg = "clusterlogforwarder 'name' cannot be empty"
	}

	if namespace == "" {
		glog.V(100).Infof("The namespace of the clusterlogforwarder is empty")

		builder.errorMsg = "clusterlogforwarder 'namespace' cannot be empty"
	}

	if !builder.Exists() {
		return nil, fmt.Errorf("clusterlogforwarder object %s doesn't exist in namespace %s", name, namespace)
	}

	return builder.Object, nil
}

// Get returns clusterlogforwarder object if found.
func (builder *ClusterLogForwarderBuilder) Get() (*clov1.ClusterLogForwarder, error) {
	if valid, err := builder.validate(); !valid {
		return nil, err
	}

	glog.V(100).Infof("Getting clusterlogforwarder %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	clusterLogForwarder := &clov1.ClusterLogForwarder{}
	err := builder.apiClient.Get(context.Background(), goclient.ObjectKey{
		Name:      builder.Definition.Name,
		Namespace: builder.Definition.Namespace,
	}, clusterLogForwarder)

	if err != nil {
		return nil, err
	}

	return clusterLogForwarder, err
}

// Create makes a clusterlogforwarder in the cluster and stores the created object in struct.
func (builder *ClusterLogForwarderBuilder) Create() (*ClusterLogForwarderBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Creating the clusterlogforwarder %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	var err error
	if !builder.Exists() {
		err = builder.apiClient.Create(context.TODO(), builder.Definition)
		if err == nil {
			builder.Object = builder.Definition
		}
	}

	return builder, err
}

// Delete removes clusterlogforwarder from a cluster.
func (builder *ClusterLogForwarderBuilder) Delete() error {
	if valid, err := builder.validate(); !valid {
		return err
	}

	glog.V(100).Infof("Deleting the clusterlogforwarder %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	if !builder.Exists() {
		return fmt.Errorf("clusterlogforwarder cannot be deleted because it does not exist")
	}

	err := builder.apiClient.Delete(context.Background(), builder.Definition)

	if err != nil {
		return fmt.Errorf("can not delete clusterlogforwarder: %w", err)
	}

	builder.Object = nil

	return nil
}

// Exists checks whether the given clusterlogforwarder exists.
func (builder *ClusterLogForwarderBuilder) Exists() bool {
	if valid, _ := builder.validate(); !valid {
		return false
	}

	glog.V(100).Infof("Checking if clusterlogforwarder %s exists in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	var err error
	builder.Object, err = builder.Get()

	return err == nil || !k8serrors.IsNotFound(err)
}

// Update renovates the existing clusterlogforwarder object with clusterlogforwarder definition in builder.
func (builder *ClusterLogForwarderBuilder) Update(force bool) (*ClusterLogForwarderBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Info("Updating clusterlogforwarder %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	err := builder.apiClient.Update(context.TODO(), builder.Definition)

	if err != nil {
		if force {
			glog.V(100).Infof("Failed to update the clusterlogforwarder object %s in namespace $s. "+
				"Note: Force flag set, executed delete/create methods instead",
				builder.Definition.Name, builder.Definition.Namespace)

			err := builder.Delete()

			if err != nil {
				glog.V(100).Infof(
					"Failed to update the clusterlogforwarder object %s in namespace $s."+
						"due to error in delete function", builder.Definition.Name, builder.Definition.Namespace)

				return nil, err
			}

			return builder.Create()
		}
	}

	if err == nil {
		builder.Object = builder.Definition
	}

	return builder, err
}

// validate will check that the builder and builder definition are properly initialized before
// accessing any member fields.
func (builder *ClusterLogForwarderBuilder) validate() (bool, error) {
	resourceCRD := "ClusterLogForwarder"

	if builder == nil {
		glog.V(100).Infof("The %s builder is uninitialized", resourceCRD)

		return false, fmt.Errorf("error: received nil %s builder", resourceCRD)
	}

	if builder.Definition == nil {
		glog.V(100).Infof("The %s is undefined", resourceCRD)

		return false, fmt.Errorf(msg.UndefinedCrdObjectErrString(resourceCRD))
	}

	if builder.apiClient == nil {
		glog.V(100).Infof("The %s builder apiclient is nil", resourceCRD)

		return false, fmt.Errorf("%s builder cannot have nil apiClient", resourceCRD)
	}

	return true, nil
}