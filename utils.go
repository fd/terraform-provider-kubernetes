package main

import (
	"strconv"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/unversioned"
)

type Setter interface {
	Set(key string, value interface{}) error
}

type ObjectBuilder struct {
	parent Setter
	key    string
	m      map[string]interface{}
}

type ListBuilder struct {
	parent Setter
	key    string
	l      []interface{}
	b      ObjectBuilder
}

func NewObjectBuilder(parent Setter, key string) *ObjectBuilder {
	return &ObjectBuilder{parent: parent, key: key}
}

func NewListBuilder(parent Setter, key string) *ListBuilder {
	return &ListBuilder{parent: parent, key: key}
}

func (b *ObjectBuilder) Set(key string, value interface{}) error {
	if strings.Contains(key, ".") {
		panic("invalid key: " + key)
	}
	if b.m == nil {
		b.m = make(map[string]interface{})
	}
	b.m[key] = value
	return nil
}

func (b *ObjectBuilder) NewList(key string) *ListBuilder {
	return NewListBuilder(b, key)
}

func (b *ObjectBuilder) Touch() {
	if b.m == nil {
		b.m = make(map[string]interface{})
	}
}

func (b *ObjectBuilder) Apply() error {
	if b.m == nil {
		return nil
	}

	if b.key != "" {
		return b.parent.Set(b.key, b.m)
	}

	for k, v := range b.m {
		err := b.parent.Set(k, v)
		if err != nil {
			return err
		}
	}

	return nil
}

func (b *ListBuilder) Set(key string, value interface{}) error {
	return b.b.Set(key, value)
}

func (b *ListBuilder) Touch() {
	b.b.Touch()
}

func (b *ListBuilder) NewList(key string) *ListBuilder {
	return b.b.NewList(key)
}

func (b *ListBuilder) Apply() error {
	b.Next()

	if b.l == nil {
		return nil
	}

	return b.parent.Set(b.key, b.l)
}

func (b *ListBuilder) Next() {
	if b.b.m != nil {
		b.l = append(b.l, b.b.m)
		b.b.m = nil
	}
}

func extractClient(v interface{}) *unversioned.Client {
	return v.(*unversioned.Client)
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
