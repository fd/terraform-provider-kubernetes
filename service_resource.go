package main

import (
	"github.com/hashicorp/terraform/helper/schema"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/errors"
	"k8s.io/kubernetes/pkg/util/intstr"
)

func serviceResource() *schema.Resource {
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
				Default:  "ClusterIP",
			},
			"selector": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type:     schema.TypeString,
					Required: true,
				},
			},
			"cluster_ip": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"load_balancer_ip": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"session_affinity": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "None",
			},
			"external_ips": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"port": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"protocol": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "TCP",
						},
						"port": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"target_port": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
						"node_port": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
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

			"load_balancer_ingress": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"hostname": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "TCP",
						},
					},
				},
			},
		},
		Create: func(r *schema.ResourceData, v interface{}) error {
			client := extractClient(v)
			namespace := r.Get("namespace").(string)
			name := r.Get("name").(string)

			item := &api.Service{}
			item.Name = name
			item.Spec.Type = api.ServiceType(r.Get("type").(string))
			item.Spec.SessionAffinity = api.ServiceAffinity(r.Get("session_affinity").(string))

			if v, ok := r.GetOk("cluster_ip"); ok {
				item.Spec.ClusterIP = v.(string)
			}
			if v, ok := r.GetOk("load_balancer_ip"); ok {
				item.Spec.LoadBalancerIP = v.(string)
			}

			writeExternalIPs(r, &item.Spec)
			writePorts(r, &item.Spec)
			writeSelectors(r, &item.Spec)
			writeLabels(r, &item.ObjectMeta)
			writeAnnotations(r, &item.ObjectMeta)

			item, err := client.Services(namespace).Create(item)
			if err != nil {
				return err
			}

			r.SetId(join(namespace, name))
			r.Set("load_balancer_ip", string(item.Spec.LoadBalancerIP))
			r.Set("cluster_ip", string(item.Spec.ClusterIP))
			readPorts(r, &item.Spec)
			readLoadBalancerIngressIPs(r, &item.Status)

			return nil
		},
		Read: func(r *schema.ResourceData, v interface{}) error {
			client := extractClient(v)
			namespace, name := split(r.Id())

			item, err := client.Services(namespace).Get(name)
			if err != nil {
				return err
			}

			readExternalIPs(r, &item.Spec)
			readPorts(r, &item.Spec)
			readSelectors(r, &item.Spec)
			readLabels(r, &item.ObjectMeta)
			readAnnotations(r, &item.ObjectMeta)
			readLoadBalancerIngressIPs(r, &item.Status)

			r.Set("type", string(item.Spec.Type))
			r.Set("session_affinity", string(item.Spec.SessionAffinity))
			r.Set("load_balancer_ip", string(item.Spec.LoadBalancerIP))
			r.Set("cluster_ip", string(item.Spec.ClusterIP))

			return nil
		},
		Update: func(r *schema.ResourceData, v interface{}) error {

			client := extractClient(v)
			namespace, name := split(r.Id())

			item, err := client.Services(namespace).Get(name)
			if err != nil {
				return err
			}

			item.Spec.Type = api.ServiceType(r.Get("type").(string))
			item.Spec.SessionAffinity = api.ServiceAffinity(r.Get("session_affinity").(string))

			if v, ok := r.GetOk("cluster_ip"); ok {
				item.Spec.ClusterIP = v.(string)
			}
			if v, ok := r.GetOk("load_balancer_ip"); ok {
				item.Spec.LoadBalancerIP = v.(string)
			}

			writeExternalIPs(r, &item.Spec)
			writePorts(r, &item.Spec)
			writeSelectors(r, &item.Spec)
			writeLabels(r, &item.ObjectMeta)
			writeAnnotations(r, &item.ObjectMeta)

			item, err = client.Services(namespace).Update(item)
			if err != nil {
				return err
			}

			r.Set("load_balancer_ip", string(item.Spec.LoadBalancerIP))
			r.Set("cluster_ip", string(item.Spec.ClusterIP))
			readPorts(r, &item.Spec)
			readLoadBalancerIngressIPs(r, &item.Status)

			return nil
		},
		Delete: func(r *schema.ResourceData, v interface{}) error {
			client := extractClient(v)
			namespace, name := split(r.Id())

			return client.Services(namespace).Delete(name)
		},
		Exists: func(r *schema.ResourceData, v interface{}) (bool, error) {
			client := extractClient(v)
			namespace, name := split(r.Id())

			_, err := client.Services(namespace).Get(name)
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

func readSelectors(r *schema.ResourceData, spec *api.ServiceSpec) {
	m := make(map[string]interface{})
	if len(spec.Selector) > 0 {
		for k, v := range spec.Selector {
			m[k] = v
		}
	}
	r.Set("selector", m)
}

func writeSelectors(r *schema.ResourceData, spec *api.ServiceSpec) {
	spec.Selector = map[string]string{}
	if selector, _ := r.Get("selector").(map[string]interface{}); selector != nil {
		for k, v := range selector {
			if s, ok := v.(string); ok {
				spec.Selector[k] = s
			}
		}
	}
}

func readExternalIPs(r *schema.ResourceData, spec *api.ServiceSpec) {
	var m []interface{}
	if len(spec.ExternalIPs) > 0 {
		for _, v := range spec.ExternalIPs {
			m = append(m, v)
		}
	}
	r.Set("external_ips", m)
}

func readLoadBalancerIngressIPs(r *schema.ResourceData, stat *api.ServiceStatus) {
	var m []interface{}
	if len(stat.LoadBalancer.Ingress) > 0 {
		for _, v := range stat.LoadBalancer.Ingress {
			m = append(m, map[string]interface{}{
				"ip":       v.IP,
				"hostname": v.Hostname,
			})
		}
	}
	r.Set("load_balancer_ingress", m)
}

func writeExternalIPs(r *schema.ResourceData, spec *api.ServiceSpec) {
	spec.ExternalIPs = nil
	if l, _ := r.Get("external_ips").([]interface{}); l != nil {
		for _, v := range l {
			if s, ok := v.(string); ok {
				spec.ExternalIPs = append(spec.ExternalIPs, s)
			}
		}
	}
}

func readPorts(r *schema.ResourceData, spec *api.ServiceSpec) {
	var m []interface{}
	if len(spec.Ports) > 0 {
		for _, v := range spec.Ports {
			m = append(m, map[string]interface{}{
				"name":        v.Name,
				"protocol":    string(v.Protocol),
				"port":        v.Port,
				"target_port": v.TargetPort.IntValue(),
				"node_port":   v.NodePort,
			})
		}
	}
	r.Set("port", m)
}

func writePorts(r *schema.ResourceData, spec *api.ServiceSpec) {
	spec.Ports = nil
	if l, _ := r.Get("port").([]interface{}); l != nil {
		for _, v := range l {
			m := v.(map[string]interface{})

			var (
				name, _       = m["name"].(string)
				protocol, _   = m["protocol"].(string)
				port, _       = m["port"].(int)
				targetPort, _ = m["target_port"].(int)
				nodePort, _   = m["node_port"].(int)
			)

			spec.Ports = append(spec.Ports, api.ServicePort{
				Name:       name,
				Protocol:   api.Protocol(protocol),
				Port:       port,
				TargetPort: intstr.FromInt(targetPort),
				NodePort:   nodePort,
			})

		}
	}
}
