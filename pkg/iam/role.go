package iam

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/iam"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	AWSIamRoleCRDName       = "awsiamrole"
	AWSIamRoleCRDNamePlural = "awsiamroles"
)

type AWSIamRole struct {
	meta_v1.TypeMeta   `json:",inline"`
	meta_v1.ObjectMeta `json:"metadata"`
	Spec               AWSIamRoleSpec `json:"spec"`
}

type AWSIamRoleSpec struct {
	// Just use the types directly from aws sdk.
	AssumeRolePolicyDocument string `json:"assumeRolePolicyDocument"`
	Description              string `json:"description"`
	Path                     string `json:"path"`
	RoleName                 string `json:"roleName"`

	ManagedPolicies []string `json:"managedPolicies"`
}

type AWSIamRoleList struct {
	meta_v1.TypeMeta `json:",inline"`
	meta_v1.ListMeta `json:"metadata"`

	Items []AWSIamRole `json:"items"`
}

func (c *Controller) AddRole(obj interface{}) {
	role := obj.(*AWSIamRole)

	// Check if our role already exists.
	iroles, err := c.svc.ListRoles(&iam.ListRolesInput{})
	if err != nil {
		fmt.Printf(err.Error())
		return
	}
	input := &iam.CreateRoleInput{
		RoleName:                 &role.Spec.RoleName,
		AssumeRolePolicyDocument: &role.Spec.AssumeRolePolicyDocument,
	}

	if role.Spec.Path != "" {
		input.Path = &role.Spec.Path
	}

	if role.Spec.Description != "" {
		input.Description = &role.Spec.Description
	}

	for _, ir := range iroles.Roles {
		if *ir.RoleName == role.Spec.RoleName {
			fmt.Printf("Skipping due to existing IAM role: %s\n", role.Spec.RoleName)
			return
		}
	}
	_, err = c.svc.CreateRole(input)
	if err != nil {
		fmt.Printf("Failed to create role: %s\n%v", role.Spec.RoleName, err)
		return
	}

	ipolicies, err := c.svc.ListPolicies(&iam.ListPoliciesInput{})
	if err != nil {
		fmt.Printf(err.Error())
		return
	}

	for _, ip := range ipolicies.Policies {
		for _, pa := range role.Spec.ManagedPolicies {
			if *ip.PolicyName == pa {
				_, err = c.svc.AttachRolePolicy(&iam.AttachRolePolicyInput{
					RoleName:  &role.Spec.RoleName,
					PolicyArn: ip.Arn,
				})
				if err != nil {
					fmt.Printf(err.Error())
					return
				}
			}
		}
	}
}

func (c *Controller) DeleteRole(obj interface{}) {
	fmt.Printf("Delete: %s", obj)
}

func (c *Controller) UpdateRole(old, new interface{}) {
	fmt.Printf("Update: %s %s", old, new)
}
