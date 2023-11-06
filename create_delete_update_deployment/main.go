package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/client-go/util/retry"
)

func main() {
	// get kubeconfig
	kubeconfig := Get_kubeconfig()
	fmt.Printf("kubeconfig: %v\n", *kubeconfig)
	// get clientset
	clientset := getClientset(kubeconfig)
	deploymentsClient := clientset.AppsV1().Deployments(apiv1.NamespaceDefault)
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "demo-deployment"},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(2),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "demo",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "demo",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:            "nginx",
							Image:           "docker.io/library/nginx:latest",
							ImagePullPolicy: apiv1.PullIfNotPresent,
						},
					},
				},
			},
		},
	}

	// create deployment
	fmt.Println("Creating a deployment....")
	result, err := deploymentsClient.Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil {
		panic("Create deployment failed. " + err.Error())
	}
	fmt.Printf("Created deployment %v\n", result.GetObjectMeta().GetName())

	// update deployment
	Prompt()
	fmt.Println("Updating deployment...")
	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		result, err := deploymentsClient.Get(context.TODO(), "demo-deployment", metav1.GetOptions{})
		if err != nil {
			panic(err.Error())
		}
		result.Spec.Replicas = int32Ptr(1)
		_, err = deploymentsClient.Update(context.TODO(), result, metav1.UpdateOptions{})
		return err
	})
	if err != nil {
		panic(fmt.Errorf("update failed %v", err.Error()))
	}
	fmt.Println("Updated deployment...")

	// list deployment
	Prompt()
	fmt.Printf("List deployment in namespace %v\n", apiv1.NamespaceDefault)
	deployList, err := deploymentsClient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	for _, result := range deployList.Items {
		fmt.Printf("*** %s (%d replicas)\n", result.Name, *result.Spec.Replicas)
	}

	// delete deployment
	Prompt()
	fmt.Println("Deleting deployment...")
	dp := metav1.DeletePropagationForeground
	if err = deploymentsClient.Delete(context.TODO(), "demo-deployment", metav1.DeleteOptions{
		PropagationPolicy: &dp,
	}); err != nil {
		panic(err.Error())
	}
	fmt.Println("Deleted deployment")
}
func Get_kubeconfig() *string {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()
	return kubeconfig
}
func getClientset(kubeconfig *string) *kubernetes.Clientset {
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return clientset
}
func int32Ptr(i int32) *int32 {
	return &i
}
func Prompt() {
	fmt.Println("-> Press Enter key to continue.")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		break
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
}
