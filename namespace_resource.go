package main

import (
	"github.com/hashicorp/terraform/helper/schema"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/errors"
)

func namespaceResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type:     schema.TypeString,
					Required: true,
				},
			},
			"annotations": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type:     schema.TypeString,
					Required: true,
				},
			},
		},
		Create: func(r *schema.ResourceData, v interface{}) error {
			client := extractClient(v)
			name := r.Get("name").(string)

			item := &api.Namespace{}
			item.ObjectMeta.Name = name

			writeLabels(r, &item.ObjectMeta)
			writeAnnotations(r, &item.ObjectMeta)

			item, err := client.Namespaces().Create(item)
			if err != nil {
				return err
			}

			r.SetId(name)
			return nil
		},
		Read: func(r *schema.ResourceData, v interface{}) error {
			client := extractClient(v)
			name := r.Id()

			item, err := client.Namespaces().Get(name)
			if err != nil {
				return err
			}

			readLabels(r, &item.ObjectMeta)
			readAnnotations(r, &item.ObjectMeta)

			r.Set("name", item.ObjectMeta.Name)
			return nil
		},
		Update: func(r *schema.ResourceData, v interface{}) error {

			if !r.HasChange("labels") && !r.HasChange("annotations") {
				return nil
			}

			client := extractClient(v)
			name := r.Id()

			item, err := client.Namespaces().Get(name)
			if err != nil {
				return err
			}

			writeLabels(r, &item.ObjectMeta)
			writeAnnotations(r, &item.ObjectMeta)

			_, err = client.Namespaces().Update(item)
			if err != nil {
				return err
			}

			return nil
		},
		Delete: func(r *schema.ResourceData, v interface{}) error {
			client := extractClient(v)
			name := r.Id()

			return client.Namespaces().Delete(name)
		},
		Exists: func(r *schema.ResourceData, v interface{}) (bool, error) {
			client := extractClient(v)
			name := r.Id()

			_, err := client.Namespaces().Get(name)
			if errors.IsNotFound(err) {
				return false, nil
			}
			if err != nil {
				return false, err
			}
			return true, nil
		},
	}
}
