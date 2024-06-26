package main

import (
	"context"
	"flag"
	"fmt"
	"strings"
	"time"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	var (
		kubeconfigPath string
	)

	flag.StringVar(&kubeconfigPath, "kubeconfigPath", "", "Absolute path to the kubeconfig file")
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		panic(err)
	}
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	res := schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
	allDeploy, err := client.
		Resource(res).
		Namespace(apiv1.NamespaceAll).
		List(context.Background(), metav1.ListOptions{FieldSelector: "metadata.namespace!=kube-system"})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Number of deployments retrieved: %d\n", len(allDeploy.Items))

	for _, deploy := range allDeploy.Items {
		// fmt.Println(pod)
		metadata, found, err := unstructured.NestedMap(deploy.Object, "metadata")
		if err != nil {
			fmt.Printf("Metadata found: %t\n", found)
			fmt.Println(err)
			continue
		}
		// fmt.Println(metadata)
		dName := metadata["name"].(string)
		if strings.Contains(strings.ToLower(dName), "database") {
			fmt.Println("Restarting " + dName)
			namespace := metadata["namespace"].(string)
			unstructured.SetNestedField(deploy.Object, time.Now().Format("20060102150405"), "spec", "template", "metadata", "annotations", "kubectl.kubernetes.io/restartedAt")
			_, err := client.
				Resource(res).
				Namespace(namespace).
				Update(
					context.Background(),
					&deploy,
					metav1.UpdateOptions{},
				)

			if err != nil {
				fmt.Println(err)
			}

		}
	}
}
