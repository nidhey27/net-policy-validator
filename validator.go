package main

import (
	"context"
	"log"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Controller struct {
	clientset *kubernetes.Clientset
}

func getClientSet(config *rest.Config) (*kubernetes.Clientset, error) {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		// log.Printf("ERROR[ClientSet]: %s\n", err.Error())
		return nil, err
	}

	return clientset, nil
}

func newController(config *rest.Config) *Controller {

	clientset, err := getClientSet(config)

	if err != nil {
		log.Printf("ERROR[]: %s", err.Error())
	}

	c := &Controller{
		clientset: clientset,
	}
	return c
}

func (c *Controller) checkPodLabels(pod corev1.Pod) error {
	log.Printf("INFO[PodLabels]: Checking Pod Labels")

	// podList, err := c.clientset.CoreV1().Pods("").List(context.Background(), metav1.ListOptions{})
	netPolList, err := c.clientset.NetworkingV1().NetworkPolicies("").List(context.Background(), metav1.ListOptions{})

	if err != nil {
		return err
	}

	// Extract labels used in each NetworkPolicy and store them in a set
	netpolLabels := make(map[string]string)
	for _, netpol := range netPolList.Items {
		for key, value := range netpol.Spec.PodSelector.MatchLabels {
			netpolLabels[key] = value
		}
	}
	// Extract labels used in each Pod and store them in a set
	podLabels := make(map[string]string)
	for key, value := range pod.Labels {
		podLabels[key] = value
	}

	// Compare labels used in each NetworkPolicy with the labels of Pod
	for lKey, lVal := range netpolLabels {
		for pKey, pVal := range podLabels {
			if lKey == pKey && lVal == pVal {
				return errors.Errorf("Warning: Label %s:%s used in a NetworkPolicy, Labels should not be edited!! \n", pKey, pVal)
			}
		}
		// if pKey, pVal := podLabels[label]; !ok {
		// 	// Generate a warning message if a NetworkPolicy uses a label that is not present in any Pod
		// 	return errors.Errorf("Warning: Label %s used in a NetworkPolicy, Labels should not be edited!! \n", label)
		// }
	}

	log.Println(netpolLabels, podLabels)
	return nil
}
