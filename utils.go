package main

import (
	"strconv"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"

	"k8s.io/kubernetes/pkg/api"
	client "k8s.io/kubernetes/pkg/client/unversioned"
)

func extractClient(v interface{}) *client.Client {
	return v.(*client.Client)
}

func readLabels(r *schema.ResourceData, meta *api.ObjectMeta) {
	m := make(map[string]interface{})
	if len(meta.Labels) > 0 {
		for k, v := range meta.Labels {
			m[k] = v
		}
	}
	r.Set("labels", m)
}

func readAnnotations(r *schema.ResourceData, meta *api.ObjectMeta) {
	m := make(map[string]interface{})
	if len(meta.Annotations) > 0 {
		for k, v := range meta.Annotations {
			if k == "terraform.io/owned" {
				continue
			}
			m[k] = v
		}
	}
	r.Set("annotations", m)
}

func writeLabels(r *schema.ResourceData, meta *api.ObjectMeta) {
	meta.Labels = map[string]string{}
	if labels, _ := r.Get("labels").(map[string]interface{}); labels != nil {
		for k, v := range labels {
			if s, ok := v.(string); ok {
				meta.Labels[k] = s
			}
		}
	}
}

func writeAnnotations(r *schema.ResourceData, meta *api.ObjectMeta) {
	meta.Annotations = map[string]string{
		"terraform.io/owned": "true",
	}
	if annotations, _ := r.Get("annotations").(map[string]interface{}); annotations != nil {
		for k, v := range annotations {
			if s, ok := v.(string); ok {
				meta.Annotations[k] = s
			}
		}
	}
}

func join(namespace, name string) string {
	return namespace + "/" + name
}

func split(id string) (string, string) {
	parts := strings.Split(id, "/")
	if len(parts) != 2 {
		panic("invalid id: " + strconv.Quote(id))
	}
	return parts[0], parts[1]
}
