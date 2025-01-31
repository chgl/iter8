package core

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	abnapp "github.com/iter8-tools/iter8/abn/application"
	pb "github.com/iter8-tools/iter8/abn/grpc"
	"github.com/iter8-tools/iter8/abn/k8sclient"
	"github.com/iter8-tools/iter8/base/summarymetrics"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"helm.sh/helm/v3/pkg/cli"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

type Scenario struct {
	// parameters to lookup
	application string
	user        string
	// expected results for lookup scenarios
	errorSubstring string
	track          string
	// additional expected results for writemetric scenarios
	metric string
	value  string
}

func TestLookup(t *testing.T) {
	testcases := map[string]Scenario{
		"no application": {application: "default/noapp", user: "foobar", errorSubstring: "application default/noapp not found", track: ""},
		"no user":        {application: "default/application", user: "", errorSubstring: "no user session provided", track: ""},
		"valid":          {application: "default/application", user: "user", errorSubstring: "", track: "candidate"},
	}

	for label, scenario := range testcases {
		t.Run(label, func(t *testing.T) {
			client, teardown := setup(t)
			defer teardown()
			testLookup(t, client, scenario)
		})
	}
}

func testLookup(t *testing.T, client *pb.ABNClient, scenario Scenario) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s, err := (*client).Lookup(
		ctx,
		&pb.Application{
			Name: scenario.application,
			User: scenario.user,
		},
	)

	if scenario.errorSubstring != "" {
		assert.Error(t, err)
		assert.ErrorContains(t, err, scenario.errorSubstring)
	} else {
		assert.NoError(t, err)

		assert.NotNil(t, s)
		assert.Equal(t, scenario.track, s.GetTrack())
	}
}

func TestGetApplicationData(t *testing.T) {
	client, teardown := setup(t)
	defer teardown()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s, err := (*client).GetApplicationData(
		ctx,
		&pb.ApplicationRequest{
			Application: "default/application",
		},
	)

	assert.NoError(t, err)
	assert.NotNil(t, s)
	jsonStr := s.GetApplicationJson()
	assert.Equal(t, "{\"name\":\"default/application\",\"tracks\":{\"candidate\":\"v2\"},\"versions\":{\"v1\":{\"metrics\":{\"metric1\":[1,45,45,45,2025]}},\"v2\":{\"metrics\":{}}}}", jsonStr)
}

func TestWriteMetric(t *testing.T) {
	testcases := map[string]Scenario{
		"no application": {application: "", user: "user", errorSubstring: "track not mapped", track: "", metric: "", value: "76"},
		"no user":        {application: "default/application", user: "", errorSubstring: "no user session provided", track: "", metric: "", value: "76"},
		"invalid value":  {application: "default/application", user: "user", errorSubstring: "strconv.ParseFloat: parsing \"abc\": invalid syntax", track: "", metric: "", value: "abc"},
		"valid":          {application: "default/application", user: "user", errorSubstring: "", track: "candidate", metric: "metric1", value: "76"},
	}

	for label, scenario := range testcases {
		t.Run(label, func(t *testing.T) {
			client, teardown := setup(t)
			defer teardown()
			abnapp.BatchWriteInterval = time.Duration(0)
			testWriteMetric(t, client, scenario)
		})
	}
}

func testWriteMetric(t *testing.T, client *pb.ABNClient, scenario Scenario) {
	// get current count of metric
	var oldCount uint32
	var a *abnapp.Application
	a, err := abnapp.Applications.Get(scenario.application)
	if scenario.application == "" {
		assert.ErrorContains(t, err, "not in memory")
		return
	}
	assert.NotNil(t, a)
	abnapp.Applications.RLock(a.Name)
	if scenario.metric != "" {
		m := getMetric(a, scenario.track, scenario.metric)
		assert.NotNil(t, m)
		oldCount = m.Count()
	}
	abnapp.Applications.RUnlock(a.Name)

	// call gRPC service WriteMetric()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = (*client).WriteMetric(
		ctx,
		&pb.MetricValue{
			Name:        scenario.metric,
			Value:       scenario.value,
			Application: scenario.application,
			User:        scenario.user,
		},
	)

	if scenario.errorSubstring != "" {
		assert.ErrorContains(t, err, scenario.errorSubstring)
	} else {
		assert.NoError(t, err) // never any errors
	}

	// verify that metric count has increased by 1
	a, _ = abnapp.Applications.Get(scenario.application)
	assert.NotNil(t, a)
	abnapp.Applications.RLock(a.Name)
	if scenario.metric != "" {
		m := getMetric(a, scenario.track, scenario.metric)
		assert.NotNil(t, m)
		assert.Equal(t, oldCount+1, m.Count())
	}
	abnapp.Applications.RUnlock(a.Name)
}

func setup(t *testing.T) (*pb.ABNClient, func()) {
	k8sclient.Client = *k8sclient.NewFakeKubeClient(cli.New())
	// populate watcher.Applications with test applications
	abnapp.Applications.Clear()
	a, err := yamlToApplication("default/application", "../../testdata", "abninputs/readtest.yaml")
	assert.NoError(t, err)
	abnapp.Applications.Put(a)
	err = ensureSecretCreated(a.Name)
	assert.NoError(t, err)

	// 49152-65535 are recommended ports; we use a random one for testing
	/* #nosec */
	port := rand.Intn(65535-49152) + 49152

	// start server
	lis, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	assert.NoError(t, err)

	serverOptions := []grpc.ServerOption{}
	grpcServer := grpc.NewServer(serverOptions...)
	pb.RegisterABNServer(grpcServer, newServer())
	go func() {
		_ = grpcServer.Serve(lis)
	}()

	// setup client
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	conn, err := grpc.Dial(lis.Addr().String(), opts...)
	assert.NoError(t, err)

	c := pb.NewABNClient(conn)

	// return client and teardown function to clean up
	return &c, func() {
		grpcServer.Stop()
		_ = lis.Close()
		_ = conn.Close()
	}
}

func yamlToApplication(name, folder, file string) (*abnapp.Application, error) {
	byteArray, err := readYamlFromFile(folder, file)
	if err != nil {
		return nil, err
	}

	return byteArrayToApplication(name, byteArray)
}

func readYamlFromFile(folder, file string) ([]byte, error) {
	_, filename, _, _ := runtime.Caller(1) // one step up the call stack
	fname := filepath.Clean(filepath.Join(filepath.Dir(filename), folder, file))
	return os.ReadFile(fname)
}

func byteArrayToApplication(name string, data []byte) (*abnapp.Application, error) {
	a := &abnapp.Application{}
	err := yaml.Unmarshal(data, a)
	if err != nil {
		return abnapp.NewApplication(name), nil
	}
	a.Name = name

	// Initialize versions if not already initialized
	if a.Versions == nil {
		a.Versions = abnapp.Versions{}
	}
	for _, v := range a.Versions {
		if v.Metrics == nil {
			v.Metrics = map[string]*summarymetrics.SummaryMetric{}
		}
	}

	return a, nil
}

func getMetric(a *abnapp.Application, track, metric string) *summarymetrics.SummaryMetric {
	version, ok := a.Tracks[track]
	if !ok {
		return nil
	}
	v, _ := a.GetVersion(version, true)
	m, _ := v.GetMetric(metric, true)
	return m
}

func ensureSecretCreated(application string) error {
	namespace := namespaceFromKey(application)
	name := secretNameFromKey(application)
	_, err := k8sclient.Client.Typed().CoreV1().Secrets(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		}
		_, err = k8sclient.Client.Typed().CoreV1().Secrets(namespace).Create(context.Background(), secret, metav1.CreateOptions{})
		return err
	}
	return nil
}

// nameFromKey returns the name from a key of the form "namespace/name"
func nameFromKey(applicationKey string) string {
	_, n := splitApplicationKey(applicationKey)
	return n
}

// secretNameFromKey returns the name of the secret used to persist an applicatiob
func secretNameFromKey(applicationKey string) string {
	return nameFromKey(applicationKey) + "-metrics"
}

// namespaceFromKey returns the namespace from a key of the form "namespace/name"
func namespaceFromKey(applicationKey string) string {
	ns, _ := splitApplicationKey(applicationKey)
	return ns
}

// splitApplicationKey is a utility function that returns the name and namespace from a key of the form "namespace/name"
func splitApplicationKey(applicationKey string) (string, string) {
	var name, namespace string
	names := strings.Split(applicationKey, "/")
	if len(names) > 1 {
		namespace, name = names[0], names[1]
	} else {
		namespace, name = "default", names[0]
	}

	return namespace, name
}
