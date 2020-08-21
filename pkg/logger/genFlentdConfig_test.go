package logger

import (
	"bytes"
	"fmt"
	fv1 "github.com/fission/fission/pkg/apis/core/v1"
	"github.com/fission/fission/pkg/crd"
	"io/ioutil"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
	"text/template"
)

func assert(c bool, msg string) {
	if !c {
		fmt.Printf("assert failed: %v", msg)
	}
}

type renderContext struct {
	Logs []map[string]string
}

func createRenderContext() (renderContext, error) {
	context := renderContext{}
	cMap := map[string]string{
		"file":        "file_test",
		"namespace":   "namespace_test",
		"function":    "function_test",
		"kafkaEnable": "true",
		"brokers":     "brokers_test",
		"topic":       "topic_test",
		"interval":    "interval_test",
	}
	context.Logs = append(context.Logs, cMap)
	return context, nil
}

func TestParse(t *testing.T) {
	content, err := ioutil.ReadFile(fmt.Sprintf("%v/%v", "/home/jingtao/repos/fission/pkg/logger/", "template.conf"))
	assert(err == nil, "read file error!")
	tpl := template.New("fluentd")
	tpl, err = tpl.Parse(string(content))
	context, _ := createRenderContext()
	var buf bytes.Buffer
	err = tpl.Execute(&buf, context)
	fmt.Print(buf.String())
}

func TestConfigMap(t *testing.T) {
	_, kc, _, err := crd.MakeFissionClient()
	if err != nil {
		fmt.Printf("%v", err)
		return
	}
	var cfm *v1.ConfigMap
	cfm, err = kc.CoreV1().ConfigMaps(fv1.GlobalSecretConfigMapNS).Get("test", metav1.GetOptions{})
	if err != nil {
		fmt.Printf("%v", err)
	}
	for k, v := range cfm.Data {
		fmt.Printf("%v: %v\n", k, v)
	}
}
