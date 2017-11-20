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
	PolicyName     string `json:"policyName"`
	PolicyDocument string `json:"policyDocument"`
	Path           string `json:"path"`
	Description    string `json:"description"`
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

	input := &iam.CreatePolicyInput{
		PolicyName:     &policy.Spec.PolicyName,
		PolicyDocument: &policy.Spec.PolicyDocument,
	}

	if policy.Spec.Path != "" {
		input.Path = &policy.Spec.Path
	}

	if policy.Spec.Description != "" {
		input.Description = &policy.Spec.Description
	}

	for _, ip := range ipolicies.Policies {
		if *ip.PolicyName == policy.Spec.PolicyName {
			fmt.Printf("Skipping policy due to existing policy: %s\n%v", policy.Spec.PolicyName, err)
			return
		}
	}
	_, err = c.svc.CreatePolicy(input)
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
