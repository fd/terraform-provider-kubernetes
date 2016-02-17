package main

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/plugin"
	"github.com/hashicorp/terraform/terraform"

	client "k8s.io/kubernetes/pkg/client/unversioned"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: Provider,
	})
}

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"host": {
				Type:     schema.TypeString,
				Required: true,
			},
			"username": {
				Type:     schema.TypeString,
				Required: true,
			},
			"password": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"kubernetes_namespace":              namespaceResource(),
			"kubernetes_secret":                 secretsResource(),
			"kubernetes_service":                serviceResource(),
			"kubernetes_replication_controller": replicationControllerResource(),
		},
		ConfigureFunc: func(r *schema.ResourceData) (interface{}, error) {

			config := &client.Config{
				Insecure: true,
				Host:     "https://" + r.Get("host").(string),
				Username: r.Get("username").(string),
				Password: r.Get("password").(string),
			}

			client, err := client.New(config)
			if err != nil {
				return nil, err
			}

			return client, nil

		},
	}
}
