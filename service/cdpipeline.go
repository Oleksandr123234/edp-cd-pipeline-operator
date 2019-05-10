package service

import (
	edpv1alpha1 "cd-pipeline-handler-controller/pkg/apis/edp/v1alpha1"
	jenkinsClient "cd-pipeline-handler-controller/pkg/jenkins"
	ClientSet "cd-pipeline-handler-controller/pkg/openshift"
	"errors"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"time"
)

const (
	StatusInit       = "initialized"
	StatusFailed     = "failed"
	StatusFinished   = "created"
	StatusInProgress = "in progress"
)

func CreateCDPipeline(cr *edpv1alpha1.CDPipeline) error {
	if cr.Status.Status != StatusInit {
		log.Printf("CD Pipeline %v is not in init status. Skipped", cr.Spec.Name)
		return errors.New(fmt.Sprintf("CD Pipeline %v is not in init status. Skipped", cr.Spec.Name))
	}

	setCdPipelineStatusFields(cr, StatusInProgress, time.Now())

	clientSet := ClientSet.CreateOpenshiftClients()

	jenkinsUrl := fmt.Sprintf("http://jenkins.%s:8080", cr.Namespace)
	jenkinsToken, jenkinsUsername, err := getJenkinsCreds(clientSet, cr.Namespace)
	if err != nil {
		rollbackCdPipeline(cr)
		return err
	}

	jenkins, err := jenkinsClient.Init(jenkinsUrl, jenkinsUsername, jenkinsToken)
	if err != nil {
		rollbackCdPipeline(cr)
		return err
	}

	_, err = jenkins.CreateFolder(cr.Name + "-cd-pipeline")
	if err != nil {
		rollbackCdPipeline(cr)
		return err
	}

	setCdPipelineStatusFields(cr, StatusFinished, time.Now())
	log.Printf("CD pipeline has been created. Status: %v", StatusFinished)
	return nil
}

func getJenkinsCreds(clientSet *ClientSet.ClientSet, namespace string) (string, string, error) {
	log.Printf("Start recieving credentials for Jenkins in namespace %v", namespace)
	jenkinsTokenSecret, err := clientSet.CoreClient.Secrets(namespace).Get("jenkins-token", metav1.GetOptions{})
	if err != nil {
		errorMsg := fmt.Sprint(err)
		log.Println(errorMsg)
		return "", "", errors.New(errorMsg)
	}

	log.Printf("Credentials for Jenkins in namespace %v has been recieved", namespace)

	return string(jenkinsTokenSecret.Data["token"]), string(jenkinsTokenSecret.Data["username"]), nil
}

func rollbackCdPipeline(cr *edpv1alpha1.CDPipeline) {
	setCdPipelineStatusFields(cr, StatusFailed, time.Now())
}

func setCdPipelineStatusFields(cr *edpv1alpha1.CDPipeline, status string, time time.Time) {
	cr.Status.Status = status
	cr.Status.LastTimeUpdated = time
	log.Printf("Status for CD pipeline %v has been updated to '%v' at %v.", cr.Spec.Name, status, time)
}