package controllers

import (
	"context"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"

	"encoding/json"

	v1 "k8s.io/api/apps/v1"
	v1types "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

func checkerr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

type Responce struct {
	Labels             map[string]string      `json:"labels"`
	Annotations        map[string]string      `json:"annotations"`
	Status             map[string]interface{} `json:"status"`
	Attachments        map[string]interface{} `json:"attachments"`
	ResyncAfterSeconds float32                `json:"resyncAfterSeconds"`
}

func getConfig(body []byte) any {

	log.Println("Getting raw body")
	request := string(body)

	// Extracting resourse
	start := strings.Index(request, "\"object\"")
	end := strings.LastIndex(request, ",\"attachments\":")
	request = request[start:end]

	// Extracting config from annotation
	start = strings.Index(request, "{\\\"apiVersion\\\"")
	end = strings.LastIndex(request, "\\n\"")
	request = request[start:end]

	request = strings.ReplaceAll(request, "\\\"", "\"")

	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode([]byte(request), nil, nil)
	checkerr(err)

	return obj
}

func addNodeAffinity(nodeLabel string, deployment *v1.Deployment) *v1.Deployment {
	// Add affinity
	var nodeSelectorTerms []v1types.NodeSelectorTerm
	newTerm := v1types.NodeSelectorTerm{
		MatchExpressions: []v1types.NodeSelectorRequirement{
			{
				Key:      nodeLabel,
				Operator: v1types.NodeSelectorOpExists,
			},
		},
	}
	nodeSelectorTerms = append(nodeSelectorTerms, newTerm)

	deployment.Spec.Template.Spec.Affinity = &v1types.Affinity{
		NodeAffinity: &v1types.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &v1types.NodeSelector{
				NodeSelectorTerms: nodeSelectorTerms,
			},
		},
	}

	return deployment
}

func deploymentToJSON(deployment *v1.Deployment) ([]byte, error) {
	deploymentUnstructured, err := runtime.DefaultUnstructuredConverter.ToUnstructured(deployment)
	if err != nil {
		return nil, err
	}

	return json.Marshal(deploymentUnstructured)
}

func ChangeConfig(c *fiber.Ctx) error {

	log.Println("Getting deployment")
	obj := getConfig(c.BodyRaw())
	newDeployment := obj.(*v1.Deployment)

	log.Println("Forming node affinity")
	targetNode := newDeployment.Labels["target-node"] 
	newDeployment = addNodeAffinity(targetNode, newDeployment)

	yamlDeploy, err := deploymentToJSON(newDeployment)
	checkerr(err)

	log.Println("Connecting to cluster")
	config, err := rest.InClusterConfig()
	checkerr(err)
	clientSet, err := kubernetes.NewForConfig(config)
	checkerr(err)
	deploymentClient := clientSet.AppsV1().Deployments("default")

	log.Println("Applying new config")
	_, err = deploymentClient.Patch(context.Background(), newDeployment.Name, types.StrategicMergePatchType, yamlDeploy, metav1.PatchOptions{})
	checkerr(err)
	log.Println("Successfully applied")

	resp := Responce{
		Labels: map[string]string{
			"status": "changed",
		},
		ResyncAfterSeconds: 10,
	}

	jsonResp, err := json.Marshal(resp)
	checkerr(err)

	return c.Send(jsonResp)
}
