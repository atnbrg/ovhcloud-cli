// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"net/http"

	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
	"github.com/maxatome/tdhttpmock"
	"github.com/ovh/ovhcloud-cli/internal/cmd"
)

//
// LIST TESTS
//

// TestCloudKubeListCmd tests that listing kubes returns the expected output.
func (ms *MockSuite) TestCloudKubeListCmd(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/kube",
		httpmock.NewStringResponder(200, `["kube-12345"]`).Once())

	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/kube/kube-12345",
		httpmock.NewStringResponder(200, `{
			"id": "kube-12345",
			"name": "test-kube",
			"region": "GRA11",
			"plan": "free",
			"version": "1.21.5",
			"status": "INSTALLING",
			"createdAt": "2021-10-12T14:23:45+00:00"
		}`).Once())

	out, err := cmd.Execute("cloud", "kube", "ls", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.String(out, `
┌────────────┬───────────┬────────┬──────┬─────────┬────────────┐
│     id     │   name    │ region │ plan │ version │   status   │
├────────────┼───────────┼────────┼──────┼─────────┼────────────┤
│ kube-12345 │ test-kube │ GRA11  │ free │ 1.21.5  │ INSTALLING │
└────────────┴───────────┴────────┴──────┴─────────┴────────────┘
💡 Use option --json or --yaml to get the raw output with all information`[1:])
}

//
// CREATION CLUSTER WITH CILIUM HUBBLE CUSTOMIZATION TESTS
//

// TestCloudKubeCreateCiliumHubbleEnabled tests that creating a kube with only Cilium Hubble enabled results in a successful creation.
func (ms *MockSuite) TestCloudKubeCreateCiliumHubbleEnabled(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(
		http.MethodPost,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/kube",
		tdhttpmock.JSONBody(td.SuperJSONOf(`{
			"customization": {
				"cilium": {
					"hubble": {
						"enabled": true
					}
				}
			}
		}`)),
		httpmock.NewStringResponder(200, `{
			"id": "kube-99999",
			"name": "test-hubble-kube"
		}`).Once())

	out, err := cmd.Execute(
		"cloud", "kube", "create",
		"--cloud-project", "fakeProjectID",
		"--region", "GRA999",
		"--cilium-hubble-enabled",
	)

	require.CmpNoError(err)
	assert.Contains(out, "created successfully")
}

// TestCloudKubeCreateCiliumHubbleUIEnabledOnly tests that creating a kube with only Cilium Hubble UI enabled results in an error since the UI flag requires all frontend/backend resource flags to be set as well.
func (ms *MockSuite) TestCloudKubeCreateCiliumHubbleUIEnabledOnly(assert, require *td.T) {
	_, err := cmd.Execute(
		"cloud", "kube", "create",
		"--cloud-project", "fakeProjectID",
		"--region", "GRA999",
		"--cilium-hubble-ui-enabled",
	)

	require.CmpError(err)
	assert.Contains(err.Error(), "--cilium-hubble-ui-enabled and all frontend/backend resource flags (limits-cpu, limits-memory, requests-cpu, requests-memory) must all be set together")
}

// TestCloudKubeCreateCiliumHubbleUIWithoutHubbleEnabled tests that creating a kube with Cilium Hubble UI enabled but without Cilium Hubble enabled results in an error since the UI flag requires the Hubble flag to be set as well.
func (ms *MockSuite) TestCloudKubeCreateCiliumHubbleUIWithoutHubbleEnabled(assert, require *td.T) {
	_, err := cmd.Execute(
		"cloud", "kube", "create",
		"--cloud-project", "fakeProjectID",
		"--region", "GRA999",
		"--cilium-hubble-ui-enabled",
		"--cilium-hubble-ui-frontend-limits-cpu=10",
		"--cilium-hubble-ui-frontend-limits-memory=100m",
		"--cilium-hubble-ui-frontend-requests-cpu=10",
		"--cilium-hubble-ui-frontend-requests-memory=200m",
		"--cilium-hubble-ui-backend-limits-cpu=10",
		"--cilium-hubble-ui-backend-limits-memory=200m",
		"--cilium-hubble-ui-backend-requests-cpu=10",
		"--cilium-hubble-ui-backend-requests-memory=200m",
	)

	require.CmpError(err)
	assert.Contains(err.Error(), "--cilium-hubble-enabled must be set together with --cilium-hubble-ui-enabled")
}

// TestCloudKubeCreateCiliumHubbleUIAndHubbleEnabled tests that creating a kube with both Cilium Hubble and Cilium Hubble UI enabled and all required UI resource flags results in a successful creation.
func (ms *MockSuite) TestCloudKubeCreateCiliumHubbleUIAndHubbleEnabled(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(
		http.MethodPost,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/kube",
		tdhttpmock.JSONBody(td.SuperJSONOf(`{
			"customization": {
				"cilium": {
					"hubble": {
						"enabled": true,
						"ui": {
							"enabled": true,
							"frontendResources": {
								"limits": {
									"cpu": "10",
									"memory": "100m"
								},
								"requests": {
									"cpu": "10",
									"memory": "200m"
								}
							},
							"backendResources": {
								"limits": {
									"cpu": "10",
									"memory": "200m"
								},
								"requests": {
									"cpu": "10",
									"memory": "200m"
								}
							}
						}
					}
				}
			}
		}`)),
		httpmock.NewStringResponder(200, `{
			"id": "kube-99999",
			"name": "test-hubble-ui-kube"
		}`).Once())

	out, err := cmd.Execute(
		"cloud", "kube", "create",
		"--cloud-project", "fakeProjectID",
		"--region", "GRA999",
		"--cilium-hubble-enabled",
		"--cilium-hubble-ui-enabled",
		"--cilium-hubble-ui-frontend-limits-cpu=10",
		"--cilium-hubble-ui-frontend-limits-memory=100m",
		"--cilium-hubble-ui-frontend-requests-cpu=10",
		"--cilium-hubble-ui-frontend-requests-memory=200m",
		"--cilium-hubble-ui-backend-limits-cpu=10",
		"--cilium-hubble-ui-backend-limits-memory=200m",
		"--cilium-hubble-ui-backend-requests-cpu=10",
		"--cilium-hubble-ui-backend-requests-memory=200m",
	)

	require.CmpNoError(err)
	assert.Contains(out, "created successfully")
}

// TestCloudKubeCreateCiliumHubbleRelayEnabledWithoutHubbleEnabled tests that creating a kube with --cilium-hubble-relay-enabled but without --cilium-hubble-enabled results in an error since the relay flag requires the hubble flag.
func (ms *MockSuite) TestCloudKubeCreateCiliumHubbleRelayEnabledWithoutHubbleEnabled(assert, require *td.T) {
	_, err := cmd.Execute(
		"cloud", "kube", "create",
		"--cloud-project", "fakeProjectID",
		"--region", "GRA999",
		"--cilium-hubble-relay-enabled",
	)

	require.CmpError(err)
	assert.Contains(err.Error(), "--cilium-hubble-enabled must be set together with --cilium-hubble-relay-enabled")
}

// TestCloudKubeCreateCiliumHubbleRelayEnabled tests that creating a kube with Cilium Hubble and Hubble Relay enabled results in a successful creation.
func (ms *MockSuite) TestCloudKubeCreateCiliumHubbleRelayEnabled(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(
		http.MethodPost,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/kube",
		tdhttpmock.JSONBody(td.SuperJSONOf(`{
			"customization": {
				"cilium": {
					"hubble": {
						"enabled": true,
						"relay": {
							"enabled": true
						}
					}
				}
			}
		}`)),
		httpmock.NewStringResponder(200, `{
			"id": "kube-99999",
			"name": "test-hubble-relay-kube"
		}`).Once())

	out, err := cmd.Execute(
		"cloud", "kube", "create",
		"--cloud-project", "fakeProjectID",
		"--region", "GRA999",
		"--cilium-hubble-enabled",
		"--cilium-hubble-relay-enabled",
	)

	require.CmpNoError(err)
	assert.Contains(out, "created successfully")
}

//
// CREATION CLUSTER WITH CILIUM CLUSTERMESH CUSTOMIZATION TESTS
//

// TestCloudKubeCreateOnlyCiliumClusterMeshEnabledWithoutClusterID tests that creating a kube with only Cilium ClusterMesh enabled and without a Cluster ID results in an error since the Cluster ID is required when enabling ClusterMesh.
func (ms *MockSuite) TestCloudKubeCreateOnlyCiliumClusterMeshEnabledWithoutClusterID(assert, require *td.T) {
	_, err := cmd.Execute(
		"cloud", "kube", "create",
		"--cloud-project", "fakeProjectID",
		"--region", "GRA999",
		"--cilium-cluster-mesh-enabled",
	)

	require.CmpError(err)
	assert.Contains(err.Error(), "--cilium-cluster-id must be set when setting any other Cilium ClusterMesh is enabled")
}

// TestCloudKubeCreateOnlyCiliumClusterID tests that creating a kube with only Cilium Cluster ID set and without Cilium ClusterMesh enabled results in an error since ClusterMesh must be enabled when setting a Cluster ID.
func (ms *MockSuite) TestCloudKubeCreateOnlyCiliumClusterID(assert, require *td.T) {
	_, err := cmd.Execute(
		"cloud", "kube", "create",
		"--cloud-project", "fakeProjectID",
		"--cilium-cluster-id=2",
		"--region", "GRA999",
	)

	require.CmpError(err)
	assert.Contains(err.Error(), "set --cilium-cluster-mesh-enabled to enable ClusterMesh when setting --cilium-cluster-id")
}

// TestCloudKubeCreateOnlyCiliumClusterMeshEnabled tests that creating a kube with only Cilium ClusterMesh enabled results in an error since all ClusterMesh API server flags must be set when enabling ClusterMesh.
func (ms *MockSuite) TestCloudKubeCreateOnlyCiliumClusterMeshEnabled(assert, require *td.T) {
	_, err := cmd.Execute(
		"cloud", "kube", "create",
		"--cloud-project", "fakeProjectID",
		"--cilium-cluster-id=2",
		"--region", "GRA999",
		"--cilium-cluster-mesh-enabled",
	)

	require.CmpError(err)
	assert.Contains(err.Error(), "--cilium-cluster-mesh-enabled, --cilium-cluster-mesh-apiserver-service-type, and --cilium-cluster-mesh-apiserver-node-port must all be set together")
}

// TestCloudKubeCreateCiliumClusterMeshWithAllOptions tests that creating a kube with all ClusterMesh options set results in a successful creation.
func (ms *MockSuite) TestCloudKubeCreateCiliumClusterMeshWithAllOptions(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(
		http.MethodPost,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/kube",
		tdhttpmock.JSONBody(td.SuperJSONOf(`{
			"customization": {
				"cilium": {
					"clusterId": 123,
					"clusterMesh": {
						"enabled": true,
						"apiserver": {
							"serviceType": "NodePort",
							"nodePort": 30000
						}
					}
				}
			}
		}`)),
		httpmock.NewStringResponder(200, `{
			"id": "kube-99999",
			"name": "test-clustermesh-kube"
		}`).Once())

	out, err := cmd.Execute(
		"cloud", "kube", "create",
		"--cloud-project", "fakeProjectID",
		"--region", "GRA999",
		"--cilium-cluster-id=123",
		"--cilium-cluster-mesh-enabled",
		"--cilium-cluster-mesh-apiserver-service-type=NodePort",
		"--cilium-cluster-mesh-apiserver-node-port=30000",
	)

	require.CmpNoError(err)
	assert.Contains(out, "created successfully")
}

// TestCloudKubeCreateCiliumClusterMeshInvalidServiceType tests that creating a kube with an invalid --cilium-cluster-mesh-apiserver-service-type value results in an error since the only allowed values are LoadBalancer, NodePort, and ClusterIP.
func (ms *MockSuite) TestCloudKubeCreateCiliumClusterMeshInvalidServiceType(assert, require *td.T) {
	_, err := cmd.Execute(
		"cloud", "kube", "create",
		"--cloud-project", "fakeProjectID",
		"--region", "GRA999",
		"--cilium-cluster-id=2",
		"--cilium-cluster-mesh-enabled",
		"--cilium-cluster-mesh-apiserver-service-type=InvalidType",
		"--cilium-cluster-mesh-apiserver-node-port=30000",
	)

	require.CmpError(err)
	assert.Contains(err.Error(), "--cilium-cluster-mesh-apiserver-service-type must be one of: LoadBalancer, NodePort, ClusterIP")
}

// TestCloudKubeCreateCiliumClusterIDOutOfRange tests that creating a kube with --cilium-cluster-id=256 results in an error since the possible value is between 1 and 255 (uint8).
func (ms *MockSuite) TestCloudKubeCreateCiliumClusterIDOutOfRange(assert, require *td.T) {
	_, err := cmd.Execute(
		"cloud", "kube", "create",
		"--cloud-project", "fakeProjectID",
		"--region", "GRA999",
		"--cilium-cluster-id=256",
		"--cilium-cluster-mesh-enabled",
		"--cilium-cluster-mesh-apiserver-service-type=NodePort",
		"--cilium-cluster-mesh-apiserver-node-port=30000",
	)

	require.CmpError(err)
}

// TestCloudKubeCreateCiliumClusterMeshDisabled tests that creating a kube with --cilium-cluster-mesh-enabled=false results in a successful creation with clusterMesh disabled and no apiserver configuration.
func (ms *MockSuite) TestCloudKubeCreateCiliumClusterMeshDisabled(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(
		http.MethodPost,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/kube",
		tdhttpmock.JSONBody(td.SuperJSONOf(`{
			"customization": {
				"cilium": {
					"clusterMesh": {
						"enabled": false
					}
				}
			}
		}`)),
		httpmock.NewStringResponder(200, `{
			"id": "kube-99999",
			"name": "test-clustermesh-disabled-kube"
		}`).Once())

	out, err := cmd.Execute(
		"cloud", "kube", "create",
		"--cloud-project", "fakeProjectID",
		"--region", "GRA999",
		"--cilium-cluster-mesh-enabled=false",
	)

	require.CmpNoError(err)
	assert.Contains(out, "created successfully")
}

// CREATION CLUSTER WITH IP ALLOCATION POLICY TESTS

// TestCloudKubeCreateWithOnlyPodsIPv4CIDR tests that creating a kube with only --ip-allocation-policy-pods-ipv4-cidr set results in an error since both CIDR flags must be set together.
func (ms *MockSuite) TestCloudKubeCreateWithOnlyPodsIPv4CIDR(assert, require *td.T) {
	_, err := cmd.Execute(
		"cloud", "kube", "create",
		"--cloud-project", "fakeProjectID",
		"--region", "GRA999",
		"--ip-allocation-policy-pods-ipv4-cidr=10.0.0.0/16",
	)

	require.CmpError(err)
	assert.Contains(err.Error(), "both --ip-allocation-policy-pods-ipv4-cidr and --ip-allocation-policy-services-ipv4-cidr must be set together")
}

// TestCloudKubeCreateWithOnlyServicesIPv4CIDR tests that creating a kube with only --ip-allocation-policy-services-ipv4-cidr set results in an error since both CIDR flags must be set together.
func (ms *MockSuite) TestCloudKubeCreateWithOnlyServicesIPv4CIDR(assert, require *td.T) {
	_, err := cmd.Execute(
		"cloud", "kube", "create",
		"--cloud-project", "fakeProjectID",
		"--region", "GRA999",
		"--ip-allocation-policy-services-ipv4-cidr=10.1.0.0/16",
	)

	require.CmpError(err)
	assert.Contains(err.Error(), "both --ip-allocation-policy-pods-ipv4-cidr and --ip-allocation-policy-services-ipv4-cidr must be set together")
}

// TestCloudKubeCreateWithBothIPAllocationPolicyCIDRs tests that creating a kube with both CIDR flags set results in a successful creation.
func (ms *MockSuite) TestCloudKubeCreateWithBothIPAllocationPolicyCIDRs(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(
		http.MethodPost,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/kube",
		tdhttpmock.JSONBody(td.SuperJSONOf(`{
			"ipAllocationPolicy": {
				"podsIpv4Cidr": "10.0.0.0/16",
				"servicesIpv4Cidr": "10.1.0.0/16"
			}
		}`)),
		httpmock.NewStringResponder(200, `{
			"id": "kube-99999",
			"name": "test-cidr-kube"
		}`).Once())

	out, err := cmd.Execute(
		"cloud", "kube", "create",
		"--cloud-project", "fakeProjectID",
		"--region", "GRA999",
		"--ip-allocation-policy-pods-ipv4-cidr=10.0.0.0/16",
		"--ip-allocation-policy-services-ipv4-cidr=10.1.0.0/16",
	)

	require.CmpNoError(err)
	assert.Contains(out, "created successfully")
}
