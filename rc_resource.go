package main

import (
	"github.com/hashicorp/terraform/helper/schema"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/errors"
)

var portResourceSpec = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"host_port": {
			Type:     schema.TypeInt,
			Optional: true,
		},
		"container_port": {
			Type:     schema.TypeInt,
			Required: true,
		},
		"protocol": {
			Type:     schema.TypeString,
			Default:  "TCP",
			Optional: true,
		},
		"host_ip": {
			Type:     schema.TypeString,
			Optional: true,
		},
	},
}

var envResourceSpec = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Required: true,
		},
		"value": {
			Type:     schema.TypeInt,
			Required: true,
		},
	},
}

var volumeMountResourceSpec = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Required: true,
		},
		"mount_path": {
			Type:     schema.TypeString,
			Required: true,
		},
		"read_only": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},
	},
}

var containerResourceSpec = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Required: true,
		},
		"image": {
			Type:     schema.TypeString,
			Required: true,
		},
		"image_pull_policy": {
			Type:     schema.TypeString,
			Required: true,
		},
		"command": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
		"args": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
		"working_dir": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"ports": {
			Type:     schema.TypeList,
			Optional: true,
			Elem:     portResourceSpec,
		},
		"env": {
			Type:     schema.TypeList,
			Optional: true,
			Elem:     envResourceSpec,
		},
		"volume_mounts": {
			Type:     schema.TypeList,
			Optional: true,
			Elem:     volumeMountResourceSpec,
		},
	},
}

func rcResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"namespace": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "default",
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "Opaque",
			},
			"data": {
				Type:     schema.TypeSet,
				Optional: true,
				Set:      secretsSetFunc,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"value": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
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
			namespace := r.Get("namespace").(string)
			name := r.Get("name").(string)

			item := &api.Secret{}
			item.Name = name
			item.Type = api.SecretType(r.Get("type").(string))

			writeLabels(r, &item.ObjectMeta)
			writeAnnotations(r, &item.ObjectMeta)
			writeSecretData(r, item)

			item, err := client.Secrets(namespace).Create(item)
			if err != nil {
				return err
			}

			r.SetId(join(namespace, name))
			return nil
		},
		Read: func(r *schema.ResourceData, v interface{}) error {
			client := extractClient(v)
			namespace, name := split(r.Id())

			item, err := client.Secrets(namespace).Get(name)
			if err != nil {
				return err
			}

			readLabels(r, &item.ObjectMeta)
			readAnnotations(r, &item.ObjectMeta)
			readSecretData(r, item)

			r.Set("type", string(item.Type))
			r.Set("name", item.ObjectMeta.Name)
			return nil
		},
		Update: func(r *schema.ResourceData, v interface{}) error {

			client := extractClient(v)
			namespace, name := split(r.Id())

			item, err := client.Secrets(namespace).Get(name)
			if err != nil {
				return err
			}

			item.Type = api.SecretType(r.Get("type").(string))
			writeLabels(r, &item.ObjectMeta)
			writeAnnotations(r, &item.ObjectMeta)
			writeSecretData(r, item)

			_, err = client.Secrets(namespace).Update(item)
			if err != nil {
				return err
			}

			return nil
		},
		Delete: func(r *schema.ResourceData, v interface{}) error {
			client := extractClient(v)
			namespace, name := split(r.Id())

			return client.Secrets(namespace).Delete(name)
		},
		Exists: func(r *schema.ResourceData, v interface{}) (bool, error) {
			client := extractClient(v)
			namespace, name := split(r.Id())

			_, err := client.Secrets(namespace).Get(name)
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
