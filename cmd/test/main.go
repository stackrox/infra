package main

import (
	"fmt"
	"os"

	"github.com/argoproj/argo/pkg/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
)

const kubeconfig = "/Users/josh/.kube/config"

func main() {
	if err := mainCmd(); err != nil {
		fmt.Fprintf(os.Stderr, "cli: %s\n", err.Error())
		os.Exit(1)
	}
}

func mainCmd() error {
	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return err
	}

	//// create the clientset
	//clientset, err := kubernetes.NewForConfig(config)
	//if err != nil {
	//	return err
	//}

	//argoclient, err := argoprojv1alpha1.NewForConfig(config)
	//if err != nil {
	//	panic(err)
	//}

	//
	//client, err := dynamic.NewForConfig(config)
	//if err != nil {
	//	panic(err)
	//}
	//
	////clientset.res
	//
	//var workflowtemplatesResource = schema.GroupVersionResource{Group: "argoproj.io", Version: "v1alpha1", Resource: "workflows"}
	//
	//results, err := client.Resource(workflowtemplatesResource).Namespace("default").List(metav1.ListOptions{})
	//if err != nil {
	//	panic(err)
	//}

	//fmt.Printf("COUNT %d\n", len(results.Items))
	//for _, item := range results.Items {
	//	fmt.Printf("FOUND: %+v\n", item.GetName())
	//}
	//
	//
	//result, err := client.Resource(workflowtemplatesResource).Namespace("default").Get("gke-automation-b6kjl", metav1.GetOptions{})
	//fmt.Printf("RESULT: %+v\n", result)

	//result.DeepCopyInto()

	//client := discovery.NewDiscoveryClientForConfigOrDie(config)
	//clientset.RESTClient().APIVersion().WithKind( "v1alpha1")
	//gvr := clientset.RESTClient().APIVersion().WithResource("argoproj.io/v1alpha1")

	//result2 := &v1alpha1.Workflow{}
	//
	//
	//err = argoclient.RESTClient().Get().
	//	Namespace("default").
	//	Resource("workflows").
	//	Name("gke-automation-b6kjl").
	//	//VersionedParams(&metav1.GetOptions{}, scheme.ParameterCodec).
	//	Do().
	//	Into(result2)
	//
	//if err != nil {
	//	panic(err)
	//}
	//
	//fmt.Printf("RESULT: %+v\n", result2)

	wfClientset := versioned.NewForConfigOrDie(config)
	wfClient := wfClientset.ArgoprojV1alpha1().Workflows("default")

	workflow, err := wfClient.Get("gke-automation-b6kjl", metav1.GetOptions{})
	if err != nil {
		panic(err)
	}

	fmt.Printf("WORKFLOW: %+v\n", workflow)
	//
	//if err := util.ResumeWorkflow(wfClient, "gke-automation-b6kjl"); err != nil {
	//	return err
	//}
	//
	//fmt.Println("RESUMED WORKFLOW")
	//
	//wfClient2 := commands.InitWorkflowClient()
	//workflow2, err := wfClient2.Get("gke-automation-b6kjl", metav1.GetOptions{})
	//
	//fmt.Printf("WORKFLOW2: %+v\n", workflow2)

	//err = clientset.RESTClient().Get().
	//Namespace("default").
	//Resource("workflows.v1alpha1.argoproj.io").
	//Name("gke-automation-b6kjl").
	//
	////VersionedParams(&metav1.GetOptions{}, scheme.ParameterCodec).
	//Do().
	//Into(result2)

	return nil
}
