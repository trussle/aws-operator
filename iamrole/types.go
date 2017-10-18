package iamrole

import (
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type IamRole struct {
	meta_v1.TypeMeta   `json:",inline"`
	meta_v1.ObjectMeta `json:"metadata"`
	Spec               IamRoleSpec `json:"spec"`
}

type IamRoleSpec struct {
	Name               string `json:"name"`
	PolicyDocument     string `json:"policyDocument"`
	RolePolicyDocument string `json:"rolePolicyDocument"`
}

type IamRoleList struct {
	meta_v1.TypeMeta `json:",inline"`
	meta_v1.ListMeta `json:"metadata"`

	Items []IamRole `json:"items"`
}
