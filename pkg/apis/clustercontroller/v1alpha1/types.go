package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KrakenCluster describes a krakencluster.
type KrakenCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KrakenClusterSpec   `json:"spec"`
	Status KrakenClusterStatus `json:"status"`
}

type CloudProviderCredentials struct {
	Username   string `json:"username,omitempty"`
	Password   string `json:"password,omitempty"`
	Accesskey  string `json:"accesskey,omitempty"`
	Secretname string `json:"secretname,omitempty"`
}

type CloudProviderName string

const (
	AwsProvider   CloudProviderName = "aws"
	MaasProvider  CloudProviderName = "maas"
	AzureProvider CloudProviderName = "azure"
)

type CloudProviderInfo struct {
	Name        CloudProviderName        `json:"name"`
	Credentials CloudProviderCredentials `json:"credentials,omitempty"`
	Region      string                   `json:"region,omitempty"`
}

type ProvisionerInfo struct {
	Name string `json:"name"`
}

type NodeProperties struct {
	Name        string `json:"name"`
	Size        uint32 `json:"size"`
	Os          string `json:"os"`
	MachineType string `json:"machineType"`
	PublicIPs   bool   `json:"publicIPs,omitempty"`
}

type FabricInfo struct {
	Name string `json:"name"`
}

type ClusterInfo struct {
	ClusterName string           `json:"clusterName"`
	Fabric      FabricInfo       `json:"fabric"`
	NodePools   []NodeProperties `json:"nodePools"`
}

// KrakenClusterSpec is the spec for a KrakenCluster resource
type KrakenClusterSpec struct {
	CustomerID    string            `json:"customerID"`
	CloudProvider CloudProviderInfo `json:"cloudProvider"`
	Provisioner   ProvisionerInfo   `json:"provisioner"`
	Cluster       ClusterInfo       `json:"cluster"`
}

type KrakenClusterState string

const (
	Unknown  KrakenClusterState = ""
	Creating KrakenClusterState = "Creating"
	Created  KrakenClusterState = "Created"
	Deleting KrakenClusterState = "Deleting"
	Deleted  KrakenClusterState = "Deleted"
)

type KrakenClusterStatus struct {
	Status     string             `json:"status"`
	State      KrakenClusterState `json:"state"`
	Kubeconfig string             `json:"kubeconfig"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KrakenClusterList is a list of KrakenCluster resources
type KrakenClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []KrakenCluster `json:"items"`
}
