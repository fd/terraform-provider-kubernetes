package main

import (
	"github.com/hashicorp/terraform/helper/schema"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/errors"
)

var secretsSetFunc = schema.HashResource(&schema.Resource{
	Schema: map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Required: true,
		},
	},
})

func secretsResource() *schema.Resource {

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

func readSecretData(r *schema.ResourceData, s *api.Secret) {
	var m = schema.NewSet(secretsSetFunc, nil)
	if len(s.Data) > 0 {
		for k, v := range s.Data {
			m.Add(map[string]interface{}{
				"name":  k,
				"value": string(v),
			})
		}
	}
	r.Set("data", m)
}

func writeSecretData(r *schema.ResourceData, s *api.Secret) {
	s.Data = map[string][]byte{}
	if set, _ := r.Get("data").(*schema.Set); set != nil {
		for _, p := range set.List() {
			m := p.(map[string]interface{})
			k := m["name"].(string)
			x := m["value"].(string)
			s.Data[k] = []byte(x)
		}
	}
}
