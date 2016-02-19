package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/pborman/uuid"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/errors"
	"k8s.io/kubernetes/pkg/api/resource"
	"k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/util/intstr"
	"k8s.io/kubernetes/pkg/util/wait"
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
			Type:     schema.TypeString,
			Optional: true,
		},
		"value_from": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{

					"field_ref": {
						Type:     schema.TypeList,
						Optional: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"field_path": {
									Type:     schema.TypeString,
									Required: true,
								},
							},
						},
					},

					"config_map_key_ref": {
						Type:     schema.TypeList,
						Optional: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"name": {
									Type:     schema.TypeString,
									Required: true,
								},
								"key": {
									Type:     schema.TypeString,
									Required: true,
								},
							},
						},
					},

					"secret_key_ref": {
						Type:     schema.TypeList,
						Optional: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"name": {
									Type:     schema.TypeString,
									Required: true,
								},
								"key": {
									Type:     schema.TypeString,
									Required: true,
								},
							},
						},
					},
				},
			},
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
		"termination_message_path": {
			Type:     schema.TypeString,
			Optional: true,
			Default:  "/dev/termination-log",
		},
		"command": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"args": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"working_dir": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"port": {
			Type:     schema.TypeList,
			Optional: true,
			Elem:     portResourceSpec,
		},
		"env": {
			Type:     schema.TypeList,
			Optional: true,
			Elem:     envResourceSpec,
		},
		"volume_mount": {
			Type:     schema.TypeList,
			Optional: true,
			Elem:     volumeMountResourceSpec,
		},
		"liveness_probe": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{

					"initial_delay": {
						Type:     schema.TypeInt,
						Optional: true,
					},
					"timeout": {
						Type:     schema.TypeInt,
						Optional: true,
					},
					"period": {
						Type:     schema.TypeInt,
						Optional: true,
						Default:  10,
					},
					"success_threshold": {
						Type:     schema.TypeInt,
						Optional: true,
						Default:  1,
					},
					"failure_threshold": {
						Type:     schema.TypeInt,
						Optional: true,
						Default:  3,
					},

					"exec": {
						Type:     schema.TypeList,
						Optional: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"command": {
									Type:     schema.TypeList,
									Required: true,
									Elem: &schema.Schema{
										Type: schema.TypeString,
									},
								},
							},
						},
					},

					"http_get": {
						Type:     schema.TypeList,
						Optional: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"path": {
									Type:     schema.TypeString,
									Required: true,
								},
								"port": {
									Type:     schema.TypeInt,
									Required: true,
								},
								"host": {
									Type:     schema.TypeString,
									Optional: true,
								},
								"scheme": {
									Type:     schema.TypeString,
									Optional: true,
									Default:  "HTTP",
								},
								"http_header": {
									Type:     schema.TypeList,
									Optional: true,
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
							},
						},
					},

					"tcp_socket": {
						Type:     schema.TypeList,
						Optional: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"port": {
									Type:     schema.TypeInt,
									Required: true,
								},
							},
						},
					},
				},
			},
		},
		"readiness_probe": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{

					"initial_delay": {
						Type:     schema.TypeInt,
						Optional: true,
					},
					"timeout": {
						Type:     schema.TypeInt,
						Optional: true,
					},
					"period": {
						Type:     schema.TypeInt,
						Optional: true,
						Default:  10,
					},
					"success_threshold": {
						Type:     schema.TypeInt,
						Optional: true,
						Default:  1,
					},
					"failure_threshold": {
						Type:     schema.TypeInt,
						Optional: true,
						Default:  3,
					},

					"exec": {
						Type:     schema.TypeList,
						Optional: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"command": {
									Type:     schema.TypeList,
									Required: true,
									Elem: &schema.Schema{
										Type: schema.TypeString,
									},
								},
							},
						},
					},

					"http_get": {
						Type:     schema.TypeList,
						Optional: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"path": {
									Type:     schema.TypeString,
									Required: true,
								},
								"port": {
									Type:     schema.TypeInt,
									Required: true,
								},
								"host": {
									Type:     schema.TypeString,
									Optional: true,
								},
								"scheme": {
									Type:     schema.TypeString,
									Optional: true,
									Default:  "HTTP",
								},
								"http_header": {
									Type:     schema.TypeList,
									Optional: true,
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
							},
						},
					},

					"tcp_socket": {
						Type:     schema.TypeList,
						Optional: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"port": {
									Type:     schema.TypeInt,
									Required: true,
								},
							},
						},
					},
				},
			},
		},
		"resources": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"limits": {
						Type:     schema.TypeList,
						Optional: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"cpu": {
									Type:     schema.TypeString,
									Optional: true,
									ValidateFunc: func(v interface{}, _ string) ([]string, []error) {
										_, err := resource.ParseQuantity(v.(string))
										if err != nil {
											return nil, []error{err}
										}
										return nil, nil
									},
								},
								"memory": {
									Type:     schema.TypeString,
									Optional: true,
									ValidateFunc: func(v interface{}, _ string) ([]string, []error) {
										_, err := resource.ParseQuantity(v.(string))
										if err != nil {
											return nil, []error{err}
										}
										return nil, nil
									},
								},
							},
						},
					},
					"requests": {
						Type:     schema.TypeList,
						Optional: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"cpu": {
									Type:     schema.TypeString,
									Optional: true,
									ValidateFunc: func(v interface{}, _ string) ([]string, []error) {
										_, err := resource.ParseQuantity(v.(string))
										if err != nil {
											return nil, []error{err}
										}
										return nil, nil
									},
								},
								"memory": {
									Type:     schema.TypeString,
									Optional: true,
									ValidateFunc: func(v interface{}, _ string) ([]string, []error) {
										_, err := resource.ParseQuantity(v.(string))
										if err != nil {
											return nil, []error{err}
										}
										return nil, nil
									},
								},
							},
						},
					},
				},
			},
		},
	},
}

func replicationControllerResource() *schema.Resource {
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
			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
				Computed: true,
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

			"replicas": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"template": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"labels": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem: &schema.Schema{
								Type:     schema.TypeString,
								Required: true,
							},
						},
						"node_selector": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem: &schema.Schema{
								Type:     schema.TypeString,
								Required: true,
							},
						},
						"volume": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{

									"name": {
										Type:     schema.TypeString,
										Required: true,
									},

									"host_path": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"path": {
													Type:     schema.TypeString,
													Required: true,
												},
											},
										},
									},

									"empty_dir": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"medium": {
													Type:     schema.TypeString,
													Required: true,
												},
											},
										},
									},

									"gce_persistent_disk": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"pd_name": {
													Type:     schema.TypeString,
													Required: true,
												},
												"fs_type": {
													Type:     schema.TypeString,
													Required: true,
												},
												"partition": {
													Type:     schema.TypeInt,
													Required: true,
												},
												"read_only": {
													Type:     schema.TypeBool,
													Required: true,
												},
											},
										},
									},

									"aws_elastic_block_store": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"volume_id": {
													Type:     schema.TypeString,
													Required: true,
												},
												"fs_type": {
													Type:     schema.TypeString,
													Required: true,
												},
												"partition": {
													Type:     schema.TypeInt,
													Required: true,
												},
												"read_only": {
													Type:     schema.TypeBool,
													Required: true,
												},
											},
										},
									},

									"git_repo": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"repository": {
													Type:     schema.TypeString,
													Required: true,
												},
												"revision": {
													Type:     schema.TypeString,
													Required: true,
												},
												"directory": {
													Type:     schema.TypeString,
													Required: true,
												},
											},
										},
									},

									"secret": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"secret_name": {
													Type:     schema.TypeString,
													Required: true,
												},
											},
										},
									},

									// TODO:
									// NFS *NFSVolumeSource
									// ISCSI *ISCSIVolumeSource
									// Glusterfs *GlusterfsVolumeSource
									// PersistentVolumeClaim *PersistentVolumeClaimVolumeSource
									// RBD *RBDVolumeSource
									// FlexVolume *FlexVolumeSource
									// Cinder *CinderVolumeSource
									// CephFS *CephFSVolumeSource
									// Flocker *FlockerVolumeSource
									// DownwardAPI *DownwardAPIVolumeSource
									// FC *FCVolumeSource
									// AzureFile *AzureFileVolumeSource
									// ConfigMap *ConfigMapVolumeSource

								},
							},
						},
						"image_pull_secret": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
						"container": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     containerResourceSpec,
						},
						"restart_policy": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"service_account_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"node_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"dns_policy": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"termination_grace_period": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
						"active_deadline": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
			},
		},

		Read:   resourceControllerRead,
		Create: resourceControllerCreate,
		Update: resourceControllerUpdate,
		Delete: resourceControllerDelete,
		Exists: resourceControllerExists,
	}
}

func resourceControllerRead(r *schema.ResourceData, v interface{}) error {
	client := extractClient(v)
	namespace, id := split(r.Id())

	item, err := client.ReplicationControllers(namespace).Get(id)
	if err != nil {
		return err
	}

	delete(item.ObjectMeta.Annotations, "kubectl.kubernetes.io/original-replicas")

	readLabels(r, &item.ObjectMeta)
	readAnnotations(r, &item.ObjectMeta)

	root := NewObjectBuilder(r, "")
	t := root.NewList("template")

	root.Set("replicas", item.Spec.Replicas)

	if tmpl := item.Spec.Template; tmpl != nil {

		t.Set("restart_policy", string(tmpl.Spec.RestartPolicy))
		t.Set("dns_policy", string(tmpl.Spec.DNSPolicy))
		t.Set("service_account_name", tmpl.Spec.ServiceAccountName)
		t.Set("node_name", tmpl.Spec.NodeName)

		if tmpl.Spec.TerminationGracePeriodSeconds != nil {
			t.Set("termination_grace_period", int(*tmpl.Spec.TerminationGracePeriodSeconds))
		}

		if tmpl.Spec.ActiveDeadlineSeconds != nil {
			t.Set("active_deadline", int(*tmpl.Spec.ActiveDeadlineSeconds))
		}

		if tmpl.Spec.NodeSelector != nil {
			var nodeSelector = map[string]interface{}{}
			for k, v := range tmpl.Spec.NodeSelector {
				nodeSelector[k] = v
			}
			t.Set("node_selector", nodeSelector)
		}

		if tmpl.ObjectMeta.Labels != nil {
			var labels = map[string]interface{}{}
			for k, v := range tmpl.ObjectMeta.Labels {
				if k == "deployment" {
					continue
				}
				labels[k] = v
			}
			t.Set("labels", labels)
		}

		vol := t.NewList("volume")
		for _, volume := range tmpl.Spec.Volumes {
			vol.Next()
			vol.Set("name", volume.Name)

			if volume.HostPath != nil {
				x := vol.NewList("host_path")
				x.Set("path", volume.HostPath.Path)
				x.Apply()
			}

			if volume.EmptyDir != nil {
				x := vol.NewList("empty_dir")
				x.Set("medium", string(volume.EmptyDir.Medium))
				x.Apply()
			}

			if volume.GCEPersistentDisk != nil {
				x := vol.NewList("gce_persistent_disk")
				x.Set("pd_name", volume.GCEPersistentDisk.PDName)
				x.Set("fs_type", volume.GCEPersistentDisk.FSType)
				x.Set("partition", volume.GCEPersistentDisk.Partition)
				x.Set("read_only", volume.GCEPersistentDisk.ReadOnly)
				x.Apply()
			}

			if volume.AWSElasticBlockStore != nil {
				x := vol.NewList("aws_elastic_block_store")
				x.Set("volume_id", volume.AWSElasticBlockStore.VolumeID)
				x.Set("fs_type", volume.AWSElasticBlockStore.FSType)
				x.Set("partition", volume.AWSElasticBlockStore.Partition)
				x.Set("read_only", volume.AWSElasticBlockStore.ReadOnly)
				x.Apply()
			}

			if volume.GitRepo != nil {
				x := vol.NewList("git_repo")
				x.Set("repository", volume.GitRepo.Repository)
				x.Set("revision", volume.GitRepo.Revision)
				x.Set("directory", volume.GitRepo.Directory)
				x.Apply()
			}

			if volume.Secret != nil {
				x := vol.NewList("secret")
				x.Set("secret_name", volume.Secret.SecretName)
				x.Apply()
			}

			// TODO:
			// NFS *NFSVolumeSource
			// ISCSI *ISCSIVolumeSource
			// Glusterfs *GlusterfsVolumeSource
			// PersistentVolumeClaim *PersistentVolumeClaimVolumeSource
			// RBD *RBDVolumeSource
			// FlexVolume *FlexVolumeSource
			// Cinder *CinderVolumeSource
			// CephFS *CephFSVolumeSource
			// Flocker *FlockerVolumeSource
			// DownwardAPI *DownwardAPIVolumeSource
			// FC *FCVolumeSource
			// AzureFile *AzureFileVolumeSource
			// ConfigMap *ConfigMapVolumeSource
		}
		vol.Apply()

		imagePullSecret := t.NewList("image_pull_secret")
		for _, x := range tmpl.Spec.ImagePullSecrets {
			imagePullSecret.Next()
			imagePullSecret.Set("name", x.Name)
		}
		imagePullSecret.Apply()

		c := t.NewList("container")
		for _, container := range tmpl.Spec.Containers {
			c.Next()
			c.Set("name", container.Name)
			c.Set("image", container.Image)
			c.Set("image_pull_policy", string(container.ImagePullPolicy))
			c.Set("termination_message_path", container.TerminationMessagePath)
			c.Set("working_dir", container.WorkingDir)

			if container.Command != nil {
				c.Set("command", container.Command)
			}

			if container.Args != nil {
				c.Set("args", container.Args)
			}

			if container.Ports != nil {
				port := c.NewList("port")
				for _, v := range container.Ports {
					port.Next()
					port.Set("name", v.Name)
					port.Set("host_port", v.HostPort)
					port.Set("host_ip", v.HostIP)
					port.Set("container_port", v.ContainerPort)
					port.Set("protocol", string(v.Protocol))
				}
				port.Apply()
			}

			if container.Env != nil {
				env := c.NewList("env")
				for _, x := range container.Env {
					env.Next()
					env.Set("name", x.Name)
					if x.ValueFrom == nil {
						env.Set("value", x.Value)
					} else {
						valueFrom := env.NewList("value_from")
						if x.ValueFrom.FieldRef != nil {
							fieldRef := valueFrom.NewList("field_ref")
							fieldRef.Set("field_path", x.ValueFrom.FieldRef.FieldPath)
							fieldRef.Apply()
						}
						if x.ValueFrom.ConfigMapKeyRef != nil {
							configMapKeyRef := valueFrom.NewList("config_map_key_ref")
							configMapKeyRef.Set("name", x.ValueFrom.ConfigMapKeyRef.Name)
							configMapKeyRef.Set("key", x.ValueFrom.ConfigMapKeyRef.Key)
							configMapKeyRef.Apply()
						}
						if x.ValueFrom.SecretKeyRef != nil {
							secretKeyRef := valueFrom.NewList("secret_key_ref")
							secretKeyRef.Set("name", x.ValueFrom.SecretKeyRef.Name)
							secretKeyRef.Set("key", x.ValueFrom.SecretKeyRef.Key)
							secretKeyRef.Apply()
						}
						valueFrom.Apply()
					}
				}
				env.Apply()
			}

			if container.VolumeMounts != nil {
				volumeMount := c.NewList("volume_mount")
				for _, x := range container.VolumeMounts {
					volumeMount.Next()
					volumeMount.Set("name", x.Name)
					volumeMount.Set("read_only", x.ReadOnly)
					volumeMount.Set("mount_path", x.MountPath)
				}
				volumeMount.Apply()
			}

			resources := c.NewList("resources")
			limits := resources.NewList("limits")
			if x := container.Resources.Limits.Cpu(); x != nil && x.Value() != 0 {
				limits.Set("cpu", x.String())
			}
			if x := container.Resources.Limits.Memory(); x != nil && x.Value() != 0 {
				limits.Set("memory", x.String())
			}
			limits.Apply()

			requests := resources.NewList("requests")
			if x := container.Resources.Requests.Cpu(); x != nil && x.Value() != 0 {
				requests.Set("cpu", x.String())
			}
			if x := container.Resources.Requests.Memory(); x != nil && x.Value() != 0 {
				requests.Set("memory", x.String())
			}
			requests.Apply()
			resources.Apply()

			if x := container.LivenessProbe; x != nil {
				livenessProbe := c.NewList("liveness_probe")
				livenessProbe.Set("initial_delay", x.InitialDelaySeconds)
				livenessProbe.Set("timeout", x.TimeoutSeconds)
				livenessProbe.Set("period", x.PeriodSeconds)
				livenessProbe.Set("success_threshold", x.SuccessThreshold)
				livenessProbe.Set("failure_threshold", x.FailureThreshold)

				if y := x.Exec; y != nil {
					exec := livenessProbe.NewList("exec")
					exec.Set("command", y.Command)
					exec.Apply()
				}
				if y := x.HTTPGet; y != nil {
					httpGet := livenessProbe.NewList("http_get")
					httpGet.Set("path", y.Path)
					httpGet.Set("port", y.Port.IntValue())
					httpGet.Set("host", y.Host)
					httpGet.Set("scheme", string(y.Scheme))
					httpHeader := httpGet.NewList("http_header")
					for _, h := range y.HTTPHeaders {
						httpHeader.Next()
						httpHeader.Set("name", h.Name)
						httpHeader.Set("value", h.Value)
					}
					httpHeader.Apply()
					httpGet.Apply()
				}
				if y := x.TCPSocket; y != nil {
					tcpSocket := livenessProbe.NewList("tcp_socket")
					tcpSocket.Set("port", y.Port.IntValue())
					tcpSocket.Apply()
				}

				livenessProbe.Apply()
			}

			if x := container.ReadinessProbe; x != nil {
				readinessProbe := c.NewList("readiness_probe")
				readinessProbe.Set("initial_delay", x.InitialDelaySeconds)
				readinessProbe.Set("timeout", x.TimeoutSeconds)
				readinessProbe.Set("period", x.PeriodSeconds)
				readinessProbe.Set("success_threshold", x.SuccessThreshold)
				readinessProbe.Set("failure_threshold", x.FailureThreshold)

				if y := x.Exec; y != nil {
					exec := readinessProbe.NewList("exec")
					exec.Set("command", y.Command)
					exec.Apply()
				}
				if y := x.HTTPGet; y != nil {
					httpGet := readinessProbe.NewList("http_get")
					httpGet.Set("path", y.Path)
					httpGet.Set("port", y.Port.IntValue())
					httpGet.Set("host", y.Host)
					httpGet.Set("scheme", string(y.Scheme))
					httpHeader := httpGet.NewList("http_header")
					for _, h := range y.HTTPHeaders {
						httpHeader.Next()
						httpHeader.Set("name", h.Name)
						httpHeader.Set("value", h.Value)
					}
					httpHeader.Apply()
					httpGet.Apply()
				}
				if y := x.TCPSocket; y != nil {
					tcpSocket := readinessProbe.NewList("tcp_socket")
					tcpSocket.Set("port", y.Port.IntValue())
					tcpSocket.Apply()
				}

				readinessProbe.Apply()
			}
		}

		c.Apply()
		t.Apply()
		err = root.Apply()
		if err != nil {
			panic(err)
		}
	} else {
		r.Set("template", nil)
	}

	return nil
}

func resourceControllerCreate(r *schema.ResourceData, v interface{}) error {
	client := extractClient(v)
	namespace := r.Get("namespace").(string)
	name := r.Get("name").(string)

	item := &api.ReplicationController{}
	item.Name = name

	err := writeReplicationController(r, item, uuid.New(), -1)
	if err != nil {
		return err
	}

	item, err = client.ReplicationControllers(namespace).Create(item)
	if err != nil {
		return err
	}

	r.SetId(join(namespace, name))
	return resourceControllerRead(r, v)
}

func resourceControllerUpdate(r *schema.ResourceData, v interface{}) error {
	client := extractClient(v)
	namespace, name := split(r.Id())
	rcs := client.ReplicationControllers(namespace)

	item, err := rcs.Get(name)
	if err != nil {
		return err
	}

	originalDeployment := ""
	originalReplicas := -1

	if item.Spec.Template != nil && item.Spec.Template.ObjectMeta.Labels != nil {
		originalDeployment = item.Spec.Template.ObjectMeta.Labels["deployment"]
	}
	if item.ObjectMeta.Annotations != nil {
		x, ok := item.ObjectMeta.Annotations["kubectl.kubernetes.io/original-replicas"]
		if ok {
			originalReplicas, err = strconv.Atoi(x)
			if err != nil {
				return err
			}
		}
		delete(item.ObjectMeta.Annotations, "kubectl.kubernetes.io/original-replicas")
	}
	if originalReplicas < 0 {
		originalReplicas = item.Spec.Replicas
	}

	if !r.HasChange("template") {
		// inplace update
		err := writeReplicationController(r, item, originalDeployment, -1)
		if err != nil {
			return err
		}

		_, err = rcs.Update(item)
		if err != nil {
			return err
		}

		return resourceControllerRead(r, v)
	}

	deployment := uuid.New()
	tmpRcName := name + "-" + deployment

	{ // create tmp RC
		item := &api.ReplicationController{}
		item.Name = tmpRcName

		err := writeReplicationController(r, item, deployment, 0)
		if err != nil {
			return err
		}

		item, err = rcs.Create(item)
		if err != nil {
			return err
		}
	}

	var (
		originalTarget    = 0
		replacementTarget = 1
		scaleReplacement  = true

		originalStep    = originalReplicas
		replacementStep = 0
	)
	if x := r.Get("replicas"); x != nil {
		replacementTarget = x.(int)
		if replacementStep > replacementTarget {
			replacementStep = replacementTarget
		}
	}

	done := false
	for !done {
		var (
			original    *api.ReplicationController
			replacement *api.ReplicationController
		)

		original, err = rcs.Get(name)
		if err != nil {
			return err
		}
		replacement, err = rcs.Get(tmpRcName)
		if err != nil {
			return err
		}

		cond := crossScaled(client, original, replacement, 30*time.Second)
		err := wait.Poll(1*time.Second, 2*time.Minute, cond)
		if err != nil {
			return err
		}

		{ // are we done
			originalIsOnTarget := false
			replacementIsOnTarget := false
			if original.Spec.Replicas == originalTarget && original.Status.Replicas == originalTarget {
				originalIsOnTarget = true
			}
			if replacement.Spec.Replicas == replacementTarget && replacement.Status.Replicas == replacementTarget {
				replacementIsOnTarget = true
			}
			if originalIsOnTarget && replacementIsOnTarget {
				done = true
				break
			}
		}

		if scaleReplacement {
			scaleReplacement = !scaleReplacement
			if replacementStep < replacementTarget {
				replacementStep++

				item, err := rcs.Get(tmpRcName)
				if err != nil {
					return err
				}

				item.Spec.Replicas = replacementStep

				_, err = rcs.Update(item)
				if err != nil {
					return err
				}
			}
		} else {
			scaleReplacement = !scaleReplacement
			if originalStep > 0 {
				originalStep--

				item, err := rcs.Get(name)
				if err != nil {
					return err
				}

				item.Spec.Replicas = originalStep

				_, err = rcs.Update(item)
				if err != nil {
					return err
				}
			}
		}
	}

	{ // delete original RC
		err := rcs.Delete(name)
		if err != nil {
			return err
		}
	}

	{ // create new RC
		item, err := rcs.Get(tmpRcName)
		if err != nil {
			return err
		}

		item.Name = name
		item.ResourceVersion = ""

		_, err = rcs.Create(item)
		if err != nil {
			return err
		}
	}

	{ // delete tmp RC
		err := rcs.Delete(tmpRcName)
		if err != nil {
			return err
		}
	}

	return nil
}

func resourceControllerDelete(r *schema.ResourceData, v interface{}) error {
	client := extractClient(v)
	namespace, name := split(r.Id())

	return client.ReplicationControllers(namespace).Delete(name)
}

func resourceControllerExists(r *schema.ResourceData, v interface{}) (bool, error) {
	client := extractClient(v)
	namespace, name := split(r.Id())

	_, err := client.ReplicationControllers(namespace).Get(name)
	if errors.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func writeReplicationController(
	r *schema.ResourceData,
	item *api.ReplicationController,
	deployment string,
	replicas int,
) error {

	writeLabels(r, &item.ObjectMeta)
	writeAnnotations(r, &item.ObjectMeta)

	if x := r.Get("replicas"); x != nil {
		item.Spec.Replicas = r.Get("replicas").(int)
		item.ObjectMeta.Annotations["kubectl.kubernetes.io/original-replicas"] = strconv.Itoa(item.Spec.Replicas)
	} else {
		item.Spec.Replicas = 1
		item.ObjectMeta.Annotations["kubectl.kubernetes.io/original-replicas"] = strconv.Itoa(item.Spec.Replicas)
	}
	if replicas >= 0 {
		item.Spec.Replicas = replicas
	}

	item.Spec.Template = &api.PodTemplateSpec{}
	err := writePodTemplateSpec(r, item.Spec.Template)
	if err != nil {
		return err
	}

	if deployment == "" {
		return fmt.Errorf("deployment must not be blank")
	}
	if item.Spec.Template.ObjectMeta.Labels == nil {
		item.Spec.Template.ObjectMeta.Labels = make(map[string]string)
	}
	item.Spec.Template.ObjectMeta.Labels["deployment"] = deployment
	item.Spec.Selector = item.Spec.Template.ObjectMeta.Labels

	return nil
}

func writePodTemplateSpec(r *schema.ResourceData, item *api.PodTemplateSpec) error {
	var template map[string]interface{}
	if x, ok := extractSingleMap(r.Get("template")); ok {
		template = x
	}
	if template == nil {
		return fmt.Errorf("missing template")
	}

	if x, ok := extractSingleMap(template["labels"]); ok && x != nil {
		item.Labels = map[string]string{}
		for k, v := range x {
			item.Labels[k] = v.(string)
		}
	}

	if x, ok := extractSingleMap(template["node_selector"]); ok && x != nil {
		item.Spec.NodeSelector = map[string]string{}
		for k, v := range x {
			item.Spec.NodeSelector[k] = v.(string)
		}
	}

	if x, ok := template["volume"].([]interface{}); ok && x != nil {
		for _, i := range x {
			volume := api.Volume{}
			writePodVolume(i.(map[string]interface{}), &volume)
			item.Spec.Volumes = append(item.Spec.Volumes, volume)
		}
	}

	if x, ok := template["image_pull_secret"].([]interface{}); ok && x != nil {
		for _, i := range x {
			ref := api.LocalObjectReference{}
			writePodImagePullSecret(i.(map[string]interface{}), &ref)
			item.Spec.ImagePullSecrets = append(item.Spec.ImagePullSecrets, ref)
		}
	}

	if x, ok := template["container"].([]interface{}); ok && x != nil {
		for _, i := range x {
			ref := api.Container{}
			err := writePodContainer(i.(map[string]interface{}), &ref)
			if err != nil {
				return err
			}
			item.Spec.Containers = append(item.Spec.Containers, ref)
		}
	}

	if x, ok := template["restart_policy"].(string); ok {
		item.Spec.RestartPolicy = api.RestartPolicy(x)
	}

	if x, ok := template["dns_policy"].(string); ok {
		item.Spec.DNSPolicy = api.DNSPolicy(x)
	}

	if x, ok := template["service_account_name"].(string); ok {
		item.Spec.ServiceAccountName = x
	}

	if x, ok := template["node_name"].(string); ok {
		item.Spec.NodeName = x
	}

	if x, ok := template["termination_grace_period"].(int); ok && x > 0 {
		l := int64(x)
		item.Spec.TerminationGracePeriodSeconds = &l
	}

	if x, ok := template["active_deadline"].(int); ok && x > 0 {
		l := int64(x)
		item.Spec.ActiveDeadlineSeconds = &l
	}

	return nil
}

func writePodVolume(m map[string]interface{}, item *api.Volume) {

	if x, ok := m["name"].(string); ok {
		item.Name = x
	}

	if n, ok := extractSingleMap(m["host_path"]); ok {
		item.HostPath = &api.HostPathVolumeSource{}
		if x, ok := n["path"].(string); ok {
			item.HostPath.Path = x
		}
	}

	if n, ok := extractSingleMap(m["empty_dir"]); ok {
		item.EmptyDir = &api.EmptyDirVolumeSource{}
		if x, ok := n["medium"].(string); ok {
			item.EmptyDir.Medium = api.StorageMedium(x)
		}
	}

	if n, ok := extractSingleMap(m["gce_persistent_disk"]); ok {
		item.GCEPersistentDisk = &api.GCEPersistentDiskVolumeSource{}
		if x, ok := n["pd_name"].(string); ok {
			item.GCEPersistentDisk.PDName = x
		}
		if x, ok := n["fs_type"].(string); ok {
			item.GCEPersistentDisk.FSType = x
		}
		if x, ok := n["partition"].(int); ok {
			item.GCEPersistentDisk.Partition = x
		}
		if x, ok := n["read_only"].(bool); ok {
			item.GCEPersistentDisk.ReadOnly = x
		}
	}

	if n, ok := extractSingleMap(m["aws_elastic_block_store"]); ok {
		item.AWSElasticBlockStore = &api.AWSElasticBlockStoreVolumeSource{}
		if x, ok := n["volume_id"].(string); ok {
			item.AWSElasticBlockStore.VolumeID = x
		}
		if x, ok := n["fs_type"].(string); ok {
			item.AWSElasticBlockStore.FSType = x
		}
		if x, ok := n["partition"].(int); ok {
			item.AWSElasticBlockStore.Partition = x
		}
		if x, ok := n["read_only"].(bool); ok {
			item.AWSElasticBlockStore.ReadOnly = x
		}
	}

	if n, ok := extractSingleMap(m["git_repo"]); ok {
		item.GitRepo = &api.GitRepoVolumeSource{}
		if x, ok := n["repository"].(string); ok {
			item.GitRepo.Repository = x
		}
		if x, ok := n["revision"].(string); ok {
			item.GitRepo.Revision = x
		}
		if x, ok := n["directory"].(string); ok {
			item.GitRepo.Directory = x
		}
	}

	if n, ok := extractSingleMap(m["secret"]); ok {
		item.Secret = &api.SecretVolumeSource{}
		if x, ok := n["secret_name"].(string); ok {
			item.Secret.SecretName = x
		}
	}

	// TODO:
	// NFS *NFSVolumeSource
	// ISCSI *ISCSIVolumeSource
	// Glusterfs *GlusterfsVolumeSource
	// PersistentVolumeClaim *PersistentVolumeClaimVolumeSource
	// RBD *RBDVolumeSource
	// FlexVolume *FlexVolumeSource
	// Cinder *CinderVolumeSource
	// CephFS *CephFSVolumeSource
	// Flocker *FlockerVolumeSource
	// DownwardAPI *DownwardAPIVolumeSource
	// FC *FCVolumeSource
	// AzureFile *AzureFileVolumeSource
	// ConfigMap *ConfigMapVolumeSource
}

func writePodImagePullSecret(m map[string]interface{}, item *api.LocalObjectReference) {

	if x, ok := m["name"].(string); ok {
		item.Name = x
	}

}

func writePodContainer(m map[string]interface{}, item *api.Container) error {

	if x, ok := m["name"].(string); ok {
		item.Name = x
	}

	if x, ok := m["image"].(string); ok {
		item.Image = x
	}

	if x, ok := m["image_pull_policy"].(string); ok {
		item.ImagePullPolicy = api.PullPolicy(x)
	}

	if x, ok := m["termination_message_path"].(string); ok {
		item.TerminationMessagePath = x
	}

	if x, ok := m["working_dir"].(string); ok {
		item.WorkingDir = x
	}

	if x, ok := m["command"].([]interface{}); ok {
		for _, y := range x {
			item.Command = append(item.Command, y.(string))
		}
	}

	if x, ok := m["args"].([]interface{}); ok {
		for _, y := range x {
			item.Args = append(item.Args, y.(string))
		}
	}

	if x, ok := m["port"].([]interface{}); ok {
		for _, y := range x {
			ref := api.ContainerPort{}
			writeContainerPort(y.(map[string]interface{}), &ref)
			item.Ports = append(item.Ports, ref)
		}
	}

	if x, ok := m["env"].([]interface{}); ok {
		for _, y := range x {
			ref := api.EnvVar{}
			writeEnvVar(y.(map[string]interface{}), &ref)
			item.Env = append(item.Env, ref)
		}
	}

	if x, ok := m["volume_mount"].([]interface{}); ok {
		for _, y := range x {
			ref := api.VolumeMount{}
			writeVolumeMount(y.(map[string]interface{}), &ref)
			item.VolumeMounts = append(item.VolumeMounts, ref)
		}
	}

	if n, ok := extractSingleMap(m["liveness_probe"]); ok {
		item.LivenessProbe = &api.Probe{}
		writeProbe(n, item.LivenessProbe)
	}

	if n, ok := extractSingleMap(m["readiness_probe"]); ok {
		item.ReadinessProbe = &api.Probe{}
		writeProbe(n, item.ReadinessProbe)
	}

	if n, ok := extractSingleMap(m["resources"]); ok {
		if o, ok := extractSingleMap(n["limits"]); ok {
			item.Resources.Limits = make(api.ResourceList)
			if x, ok := o["cpu"].(string); ok && x != "" {
				q, err := resource.ParseQuantity(x)
				if err != nil {
					return fmt.Errorf("%s for %q", err, x)
				}
				item.Resources.Limits[api.ResourceCPU] = *q
			}
			if x, ok := o["memory"].(string); ok && x != "" {
				q, err := resource.ParseQuantity(x)
				if err != nil {
					return fmt.Errorf("%s for %q", err, x)
				}
				item.Resources.Limits[api.ResourceMemory] = *q
			}
		}
		if o, ok := extractSingleMap(n["requests"]); ok {
			item.Resources.Requests = make(api.ResourceList)
			if x, ok := o["cpu"].(string); ok && x != "" {
				q, err := resource.ParseQuantity(x)
				if err != nil {
					return fmt.Errorf("%s for %q", err, x)
				}
				item.Resources.Requests[api.ResourceCPU] = *q
			}
			if x, ok := o["memory"].(string); ok && x != "" {
				q, err := resource.ParseQuantity(x)
				if err != nil {
					return fmt.Errorf("%s for %q", err, x)
				}
				item.Resources.Requests[api.ResourceMemory] = *q
			}
		}
	}

	return nil
}

func writeContainerPort(m map[string]interface{}, item *api.ContainerPort) {
	if x, ok := m["name"].(string); ok {
		item.Name = x
	}
	if x, ok := m["host_port"].(int); ok {
		item.HostPort = x
	}
	if x, ok := m["host_ip"].(string); ok {
		item.HostIP = x
	}
	if x, ok := m["container_port"].(int); ok {
		item.ContainerPort = x
	}
	if x, ok := m["protocol"].(string); ok {
		item.Protocol = api.Protocol(x)
	}
}

func writeEnvVar(m map[string]interface{}, item *api.EnvVar) {
	if x, ok := m["name"].(string); ok {
		item.Name = x
	}
	if x, ok := m["value"].(string); ok {
		item.Value = x
	}

	if n, ok := extractSingleMap(m["value_from"]); ok {
		item.Value = ""
		item.ValueFrom = &api.EnvVarSource{}

		if o, ok := extractSingleMap(n["field_ref"]); ok {
			item.ValueFrom.FieldRef = &api.ObjectFieldSelector{}
			if x, ok := o["field_path"].(string); ok {
				item.ValueFrom.FieldRef.FieldPath = x
			}
		}

		if o, ok := extractSingleMap(n["config_map_key_ref"]); ok {
			item.ValueFrom.ConfigMapKeyRef = &api.ConfigMapKeySelector{}
			if x, ok := o["name"].(string); ok {
				item.ValueFrom.ConfigMapKeyRef.Name = x
			}
			if x, ok := o["key"].(string); ok {
				item.ValueFrom.ConfigMapKeyRef.Key = x
			}
		}

		if o, ok := extractSingleMap(n["secret_key_ref"]); ok {
			item.ValueFrom.SecretKeyRef = &api.SecretKeySelector{}
			if x, ok := o["name"].(string); ok {
				item.ValueFrom.SecretKeyRef.Name = x
			}
			if x, ok := o["key"].(string); ok {
				item.ValueFrom.SecretKeyRef.Key = x
			}
		}
	}
}

func writeVolumeMount(m map[string]interface{}, item *api.VolumeMount) {
	if x, ok := m["name"].(string); ok {
		item.Name = x
	}
	if x, ok := m["read_only"].(bool); ok {
		item.ReadOnly = x
	}
	if x, ok := m["mount_path"].(string); ok {
		item.MountPath = x
	}
}

func extractSingleMap(v interface{}) (map[string]interface{}, bool) {
	if m, ok := v.(map[string]interface{}); ok {
		return m, true
	}
	if l, ok := v.([]interface{}); ok {
		if len(l) != 1 {
			return nil, false
		}
		return extractSingleMap(l[0])
	}
	return nil, false
}

func writeProbe(m map[string]interface{}, item *api.Probe) {
	if x, ok := m["initial_delay"].(int); ok {
		item.InitialDelaySeconds = x
	}
	if x, ok := m["timeout"].(int); ok {
		item.TimeoutSeconds = x
	}
	if x, ok := m["period"].(int); ok {
		item.PeriodSeconds = x
	}
	if x, ok := m["success_threshold"].(int); ok {
		item.SuccessThreshold = x
	}
	if x, ok := m["failure_threshold"].(int); ok {
		item.FailureThreshold = x
	}

	if n, ok := extractSingleMap(m["exec"]); ok {
		item.Exec = &api.ExecAction{}
		if l, ok := n["command"].([]interface{}); ok {
			for _, x := range l {
				item.Exec.Command = append(item.Exec.Command, x.(string))
			}
		}
	}

	if n, ok := extractSingleMap(m["http_get"]); ok {
		item.HTTPGet = &api.HTTPGetAction{}
		if x, ok := n["path"].(string); ok {
			item.HTTPGet.Path = x
		}
		if x, ok := n["port"].(int); ok {
			item.HTTPGet.Port = intstr.FromInt(x)
		}
		if x, ok := n["host"].(string); ok {
			item.HTTPGet.Host = x
		}
		if x, ok := n["scheme"].(string); ok {
			item.HTTPGet.Scheme = api.URIScheme(x)
		}
		if l, ok := n["http_header"].([]interface{}); ok {
			for _, x := range l {
				if y, ok := extractSingleMap(x); ok {
					h := api.HTTPHeader{}
					if x, ok := y["name"].(string); ok {
						h.Name = x
					}
					if x, ok := y["value"].(string); ok {
						h.Value = x
					}
					item.HTTPGet.HTTPHeaders = append(item.HTTPGet.HTTPHeaders, h)
				}
			}
		}
	}

	if n, ok := extractSingleMap(m["tcp_socket"]); ok {
		item.TCPSocket = &api.TCPSocketAction{}
		if x, ok := n["port"].(int); ok {
			item.TCPSocket.Port = intstr.FromInt(x)
		}
	}
}

func crossScaled(c unversioned.Interface, oldRC, newRC *api.ReplicationController, settleDuration time.Duration) wait.ConditionFunc {
	oldRCReady := unversioned.ControllerHasDesiredReplicas(c, oldRC)
	newRCReady := unversioned.ControllerHasDesiredReplicas(c, newRC)
	oldRCPodsReady := desiredPodsAreReady(c, oldRC, settleDuration)
	newRCPodsReady := desiredPodsAreReady(c, newRC, settleDuration)
	return func() (done bool, err error) {

		if ok, err := oldRCReady(); err != nil || !ok {
			return ok, err
		}

		if ok, err := newRCReady(); err != nil || !ok {
			return ok, err
		}
		if ok, err := oldRCPodsReady(); err != nil || !ok {
			return ok, err
		}

		if ok, err := newRCPodsReady(); err != nil || !ok {
			return ok, err
		}

		return true, nil
	}
}

func scaled(c unversioned.Interface, rc *api.ReplicationController, settleDuration time.Duration) wait.ConditionFunc {
	rcReady := unversioned.ControllerHasDesiredReplicas(c, rc)
	rcPodsReady := desiredPodsAreReady(c, rc, settleDuration)
	return func() (done bool, err error) {

		if ok, err := rcReady(); err != nil || !ok {
			return ok, err
		}

		if ok, err := rcPodsReady(); err != nil || !ok {
			return ok, err
		}

		return true, nil
	}
}

func desiredPodsAreReady(c unversioned.Interface, rc *api.ReplicationController, settleDuration time.Duration) wait.ConditionFunc {
	return func() (done bool, err error) {
		selector := labels.Set(rc.Spec.Selector).AsSelector()
		options := api.ListOptions{LabelSelector: selector}
		pods, err := c.Pods(rc.Namespace).List(options)
		if err != nil {
			return false, err
		}

		var ready = 0
		var nonReady = 0

		for _, pod := range pods.Items {
			if !api.IsPodReady(&pod) {
				nonReady++
				continue
			}
			if len(pod.Status.ContainerStatuses) == 0 {
				continue
			}

			now := time.Now()
			var readyContainers = 0
			for _, c := range pod.Status.ContainerStatuses {
				if !c.Ready {
					continue
				}
				if c.State.Running == nil {
					continue
				}
				if c.State.Running.StartedAt.After(now.Add(-settleDuration)) {
					continue
				}
				readyContainers++
			}
			if readyContainers != len(pod.Spec.Containers) {
				continue
			}

			ready++
		}

		return ready == rc.Spec.Replicas && nonReady == 0, nil
	}
}
