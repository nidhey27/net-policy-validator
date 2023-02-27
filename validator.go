package main

import (
	"context"
	"log"

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

func (c *Controller) checkPodLabels() error {

	netPolList, err := c.clientset.NetworkingV1().NetworkPolicies("").List(context.Background(), metav1.ListOptions{})

	if err != nil {
		return err
	}

	log.Println(netPolList)
	
	return nil
}
