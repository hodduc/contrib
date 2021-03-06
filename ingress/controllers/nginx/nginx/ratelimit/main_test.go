/*
Copyright 2015 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package ratelimit

import (
	"testing"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/util/intstr"
)

func buildIngress() *extensions.Ingress {
	defaultBackend := extensions.IngressBackend{
		ServiceName: "default-backend",
		ServicePort: intstr.FromInt(80),
	}

	return &extensions.Ingress{
		ObjectMeta: api.ObjectMeta{
			Name:      "foo",
			Namespace: api.NamespaceDefault,
		},
		Spec: extensions.IngressSpec{
			Backend: &extensions.IngressBackend{
				ServiceName: "default-backend",
				ServicePort: intstr.FromInt(80),
			},
			Rules: []extensions.IngressRule{
				{
					Host: "foo.bar.com",
					IngressRuleValue: extensions.IngressRuleValue{
						HTTP: &extensions.HTTPIngressRuleValue{
							Paths: []extensions.HTTPIngressPath{
								{
									Path:    "/foo",
									Backend: defaultBackend,
								},
							},
						},
					},
				},
			},
		},
	}
}

func TestAnnotations(t *testing.T) {
	ing := buildIngress()

	lip := ingAnnotations(ing.GetAnnotations()).limitIp()
	if lip != 0 {
		t.Error("Expected 0 in limit by ip but %v was returned", lip)
	}

	lrps := ingAnnotations(ing.GetAnnotations()).limitRps()
	if lrps != 0 {
		t.Error("Expected 0 in limit by rps but %v was returend", lrps)
	}

	data := map[string]string{}
	data[limitIp] = "5"
	data[limitRps] = "100"
	ing.SetAnnotations(data)

	lip = ingAnnotations(ing.GetAnnotations()).limitIp()
	if lip != 5 {
		t.Error("Expected %v in limit by ip but %v was returend", lip)
	}

	lrps = ingAnnotations(ing.GetAnnotations()).limitRps()
	if lrps != 100 {
		t.Error("Expected 100 in limit by rps but %v was returend", lrps)
	}
}

func TestWithoutAnnotations(t *testing.T) {
	ing := buildIngress()
	_, err := ParseAnnotations(ing)
	if err == nil {
		t.Error("Expected error with ingress without annotations")
	}
}

func TestBadRateLimiting(t *testing.T) {
	ing := buildIngress()

	data := map[string]string{}
	data[limitIp] = "0"
	data[limitRps] = "0"
	ing.SetAnnotations(data)

	_, err := ParseAnnotations(ing)
	if err == nil {
		t.Errorf("Expected error with invalid limits (0)")
	}

	data = map[string]string{}
	data[limitIp] = "5"
	data[limitRps] = "100"
	ing.SetAnnotations(data)

	rateLimit, err := ParseAnnotations(ing)
	if err != nil {
		t.Errorf("Uxpected error: %v", err)
	}

	if rateLimit.Connections.Limit != 5 {
		t.Error("Expected 5 in limit by ip but %v was returend", rateLimit.Connections)
	}

	if rateLimit.RPS.Limit != 100 {
		t.Error("Expected 100 in limit by rps but %v was returend", rateLimit.RPS)
	}
}
