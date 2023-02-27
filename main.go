package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/spf13/pflag"
	admv1beta1 "k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apiserver/pkg/server"
	"k8s.io/apiserver/pkg/server/options"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/component-base/cli/globalflag"
)

const (
	valCon = "NetPol-controller"
)

var config *rest.Config

type Options struct {
	SecureServingOptions options.SecureServingOptions
}

// Method on Option struct to add flags
func (o *Options) AddFlagSet(fs *pflag.FlagSet) {
	o.SecureServingOptions.AddFlags(fs)
}

type Config struct {
	SecureServingInfo *server.SecureServingInfo
}

func (o *Options) Config() *Config {
	if err := o.SecureServingOptions.MaybeDefaultWithSelfSignedCerts("0.0.0.0", nil, nil); err != nil {
		panic(err)
	}
	c := &Config{}

	o.SecureServingOptions.ApplyTo(&c.SecureServingInfo)
	return c
}

// Func to return default options
func NewDefaultOptions() *Options {
	o := &Options{
		SecureServingOptions: *options.NewSecureServingOptions(),
	}

	o.SecureServingOptions.BindPort = 8443
	// If we will not provide crt and key file then this will generate those files for us
	o.SecureServingOptions.ServerCert.PairName = valCon

	return o
}

func init() {
	// ######## KUBECONFIG FOR CLIENT SET ########
	var err error
	var kubeconfig *string

	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	// flag.Parse()
	config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Printf("Building config from flags failed, %s, trying to build inclusterconfig", err.Error())
		config, err = rest.InClusterConfig()
		if err != nil {
			log.Printf("ERROR[Building config]: %s\n", err.Error())

		}
	}

	// ######## KUBECONFIG FOR CLIENT SET ########
}

func main() {
	options := NewDefaultOptions()
	fs := pflag.NewFlagSet(valCon, pflag.ExitOnError)

	// Add -help flag for our app
	globalflag.AddGlobalFlags(fs, valCon)

	// Adding
	options.AddFlagSet(fs)

	// Parse flag set
	err := fs.Parse(os.Args)

	if err != nil {
		log.Printf("ERROR: %s", err.Error())
	}

	c := options.Config()

	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(ServeCRValidation))

	stopCh := server.SetupSignalHandler()

	_, ch, err := c.SecureServingInfo.Serve(mux, 30*time.Second, stopCh)

	if err != nil {
		log.Printf("ERROR: %s", err.Error())
	} else {
		<-ch
	}

}

var (
	scheme = runtime.NewScheme()
	codecs = serializer.NewCodecFactory(scheme)
)

func ServeCRValidation(w http.ResponseWriter, r *http.Request) {
	fmt.Println("ServeCRValidation was called")
	body, err := ioutil.ReadAll(r.Body)

	// Read body and get instance of admissionReview object
	decoder := codecs.UniversalDeserializer()

	// get GVk for admissionReview Object
	gvk := admv1beta1.SchemeGroupVersion.WithKind("AdmisisonReview")

	// Var of type admission review
	var admissionReview admv1beta1.AdmissionReview
	_, _, err = decoder.Decode(body, &gvk, &admissionReview)

	if err != nil {
		log.Printf("ERROR[Converting Req Body to Admission Type] %s\n", err.Error())
	}

	// convert cr spec from admission review object
	gvk_cr := admv1beta1.SchemeGroupVersion.WithKind("Pod")
	var br corev1.Pod
	_, _, err = decoder.Decode(admissionReview.Request.Object.Raw, &gvk_cr, &br)

	if err != nil {
		log.Printf("ERROR[Converting Admission Req Raw Obj to POD Type] %s\n", err.Error())
	}

	fmt.Printf("POD that we have is %+v\n", br)
	c := newController(config)

	allow, err := validateRequest(c)
	var resp admv1beta1.AdmissionResponse
	if !allow || err != nil {
		resp = admv1beta1.AdmissionResponse{
			UID:     admissionReview.Request.UID,
			Allowed: allow,
			Result: &v1.Status{
				Message: "",
			},
		}
	} else {
		resp = admv1beta1.AdmissionResponse{
			UID:     admissionReview.Request.UID,
			Allowed: allow,
			Result: &v1.Status{
				Message: "",
			},
		}
	}
	admissionReview.Response = &resp

	res, err := json.Marshal(admissionReview)

	if err != nil {
		log.Printf("error %s, while converting response to byte slice", err.Error())
	}

	_, err = w.Write(res)

	if err != nil {
		log.Printf("error %s, writing respnse to responsewriter", err.Error())
	}
}

func validateRequest(c *Controller) (bool, error) {
	c.checkPodLabels()
	return false, nil
}
