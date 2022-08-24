package cmd

import (
	"context"
	"flag"
	"fmt"
	"github.com/FFFFFaraway/MPI-Operator/cmd/myctl/utils"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
	"strings"
)

type SubmitArgs struct {
	Name         string `yaml:"name"`
	NameSpace    string `yaml:"namespace"`
	NumWorkers   int    `yaml:"numWorkers"`
	GitUrl       string `yaml:"gitUrl"`
	GitRepoName  string `yaml:"gitRepoName"`
	WorkDir      string `yaml:"workDir"`
	Command      string `yaml:"command"`
	PipInstall   bool   `yaml:"pipInstall"`
	GpuPerWorker int    `yaml:"gpuPerWorker"`
}

var (
	submitArgs   SubmitArgs
	restConfig   *rest.Config
	clientConfig clientcmd.ClientConfig
	clientset    *kubernetes.Clientset
)

func createNamespace(client *kubernetes.Clientset, namespace string) error {
	ns := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	_, err := client.CoreV1().Namespaces().Create(context.TODO(), ns, metav1.CreateOptions{})
	return err
}

func ensureNamespace(ns string) error {
	_, err := clientset.CoreV1().Namespaces().Get(context.TODO(), ns, metav1.GetOptions{})
	if err != nil && errors.IsNotFound(err) {
		if err = createNamespace(clientset, ns); err != nil {
			return err
		}
	}
	if err != nil {
		return err
	}
	return nil
}

// submitCmd represents the submit command
var submitCmd = &cobra.Command{
	Use:   "submit [mpijob name]",
	Short: "submit a mpi job",
	Long:  `submit a mpi job`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		submitArgs.Name = args[0]
		if submitArgs.NameSpace == "" {
			fmt.Println("namespace needed")
			return
		}
		if submitArgs.GitUrl == "" {
			fmt.Println("git url needed")
			return
		}
		parts := strings.Split(strings.Trim(submitArgs.GitUrl, "/"), "/")
		submitArgs.GitRepoName = strings.Split(parts[len(parts)-1], ".git")[0]
		if submitArgs.Command == "" {
			fmt.Println("command needed")
			return
		}
		if err := ensureNamespace(submitArgs.NameSpace); err != nil {
			return
		}
		if err := utils.InstallRelease(submitArgs.Name, submitArgs.NameSpace, submitArgs, "../../charts/mpijob"); err != nil {
			fmt.Println("helm install error", err)
			return
		}
	},
}

func initKubeClient() error {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err = kubernetes.NewForConfig(config)
	return err
}

func init() {
	if err := initKubeClient(); err != nil {
		return
	}
	rootCmd.AddCommand(submitCmd)
	submitCmd.Flags().StringVar(&submitArgs.NameSpace, "ns", "", "MPI Job Namespace")
	submitCmd.Flags().IntVarP(&submitArgs.NumWorkers, "numWorkers", "n", 1, "Number of Workers")
	submitCmd.Flags().StringVarP(&submitArgs.GitUrl, "gitUrl", "i", "", "git repo link for sync code")
	submitCmd.Flags().StringVar(&submitArgs.WorkDir, "wd", ".", "working directory under project")
	submitCmd.Flags().StringVarP(&submitArgs.Command, "command", "c", "", "entry point")
	submitCmd.Flags().BoolVar(&submitArgs.PipInstall, "pip", false, "whether needed to run pip install requirements.txt for workers")
	submitCmd.Flags().IntVar(&submitArgs.GpuPerWorker, "gpu", 1, "number of gpu allocated for each workers")
}
