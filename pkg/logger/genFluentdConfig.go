package logger

import (
	"fmt"
	fv1 "github.com/fission/fission/pkg/apis/core/v1"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	k8sCache "k8s.io/client-go/tools/cache"
	"time"
)

const configPath = "/fluentd/config/config.yaml"

// jingtao-add generate fluentd config by config.
func genFluentdConfig(cfm corev1.ConfigMap) {

}
func MakeFluentdConfigWatcher(zapLogger *zap.Logger, k8sClientSet *kubernetes.Clientset) k8sCache.Controller {
	resyncPeriod := 30 * time.Second
	lw := k8sCache.NewListWatchFromClient(k8sClientSet.CoreV1().RESTClient(), "configmaps", fv1.GlobalSecretConfigMapNS, fields.Everything())
	_, controller := k8sCache.NewInformer(lw, &corev1.ConfigMap{}, resyncPeriod,
		k8sCache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				cfm := obj.(*corev1.ConfigMap)
				fmt.Printf("%v", cfm)
			},
			UpdateFunc: func(_, obj interface{}) {
				cfm := obj.(*corev1.ConfigMap)
				fmt.Printf("%v", cfm)
			},
			DeleteFunc: func(obj interface{}) {
				// todo how to do?
			},
		})
	return controller
}
