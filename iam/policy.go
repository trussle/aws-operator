package iam

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/iam"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	AWSIamPolicyCRDName       = "awsiampolicy"
	AWSIamPolicyCRDNamePlural = "awsiampolicies"
	AWSIamPolicyCRDGroup      = "trussle.com"
)

type AWSIamPolicy struct {
	meta_v1.TypeMeta   `json:",inline"`
	meta_v1.ObjectMeta `json:"metadata"`
	Spec               AWSIamPolicySpec `json:"spec"`
}

type AWSIamPolicySpec struct {
	// Just use the types directly from aws sdk.
	PolicySpec *iam.CreatePolicyInput `json:"policySpec"`
}

type AWSIamPolicyList struct {
	meta_v1.TypeMeta `json:",inline"`
	meta_v1.ListMeta `json:"metadata"`

	Items []AWSIamPolicy `json:"items"`
}

func (c *Controller) AddPolicy(obj interface{}) {
	policy := obj.(*AWSIamPolicy)

	ipolicies, err := c.svc.ListPolicies(&iam.ListPoliciesInput{})
	if err != nil {
		fmt.Printf(err.Error())
		return
	}
	for _, ip := range ipolicies.Policies {
		if ip.PolicyName == policy.Spec.PolicySpec.PolicyName {
			fmt.Printf("Skipping policy due to existing policy: %s\n%v", policy.Spec.PolicySpec.PolicyName, err)
			return
		}
	}
	_, err := c.svc.CreatePolicy(policy.Spec.PolicySpec)
	if err != nil {
		fmt.Printf(err.Error())
		return
	}

	return
}

func (c *Controller) DeletePolicy(obj interface{}) {
	fmt.Printf("Delete: %s", obj)
}

func (c *Controller) UpdatePolicy(old, new interface{}) {
	fmt.Printf("Update: %s %s", old, new)
}
