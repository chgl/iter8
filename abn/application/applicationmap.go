package application

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/iter8-tools/iter8/abn/k8sclient"
	"github.com/iter8-tools/iter8/base/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

const (
	// secretPostfix is the postfix added to an application name to create a secret name
	/* #nosec */
	secretPostfix string = "-metrics"
	// secretKey is the name of the key in the data field of a kubernetes secret in which the application will be written
	secretKey string = "application.yaml"
	// defaultBatchWriteInterval is the default value of BatchWriteInterval
	defaultBatchWriteInterval time.Duration = time.Duration(60 * time.Second)
	// defaultMaxApplicationDataBytes is the default of maxApplicationDataBytes
	// must be less than the maximum size of a Kubernetes secret (1 MB)
	// The size of an application is proportional to the number of versions and the number of metrics per version.
	// Since only summary metrics are permitted, each is a fixed size
	defaultMaxApplicationDataBytes int = 750000
)

var (
	// Applications is map of app name to Application
	// This is a global variable used to maintain an internal representation of the applications in a cluster
	Applications ThreadSafeApplicationMap
	// BatchWriteInterval is the interval over which changes may batched before being persisted
	BatchWriteInterval time.Duration
	// maxApplicationDataBytes is the maximum number of bytes allowed in an application (as YAML converted to []byte)
	// this limit prevents trying to persist an application that is too large (Kubernetes secrets have a 1 MB size limit)
	maxApplicationDataBytes int
)

// initialize global variables
func init() {
	Applications = ThreadSafeApplicationMap{
		apps:           map[string]*Application{},
		mutexes:        map[string]*sync.RWMutex{},
		lastWriteTimes: map[string]*time.Time{},
	}
	BatchWriteInterval = defaultBatchWriteInterval
	maxApplicationDataBytes = defaultMaxApplicationDataBytes // a secret's maximum size is 1MB
}

// ThreadSafeApplicationMap is type to control thread safety of operations on an application map
type ThreadSafeApplicationMap struct {
	// mutex is used to mediate read/write of the application map (ie, Get vs Add/Clear)
	mutex sync.RWMutex
	apps  map[string]*Application
	// mutexes mediate read/write of individual applications within the map
	mutexes        map[string]*sync.RWMutex
	lastWriteTimes map[string]*time.Time
}

// RLock lock application for reading
func (m *ThreadSafeApplicationMap) RLock(application string) {
	m.mutexes[application].RLock()
}

// RUnlock undoes a single RLock call
func (m *ThreadSafeApplicationMap) RUnlock(application string) {
	m.mutexes[application].RUnlock()
}

// Lock lock application for writing
func (m *ThreadSafeApplicationMap) Lock(application string) {
	m.mutexes[application].Lock()
}

// Unlock unlocks application
func (m *ThreadSafeApplicationMap) Unlock(application string) {
	m.mutexes[application].Unlock()
}

// Put adds an application into the application map if it is not already there
// Returns the application that is/was there
func (m *ThreadSafeApplicationMap) Put(a *Application) *Application {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	current, ok := m.apps[a.Name]
	if ok {
		return current
	}
	m.apps[a.Name] = a
	m.mutexes[a.Name] = &sync.RWMutex{}
	return a
}

// Get application object from in memory map
func (m *ThreadSafeApplicationMap) Get(application string) (*Application, error) {
	// if available in the in-memory map, return it
	m.mutex.RLock()
	a, ok := m.apps[application]
	m.mutex.RUnlock()
	if ok {
		return a, nil
	}

	return nil, errors.New(application + " not in memory")
}

// Read the application object if necessary from persistent storage (a secret)
func (m *ThreadSafeApplicationMap) Read(application string) (*Application, error) {
	a, err := m.Get(application)
	if a != nil {
		return a, err
	}

	// otherwise, read from persistent store (secret)
	// if no secret, create new object
	a, err = m.readFromSecret(application)

	// and add to the in memory map
	a = m.Put(a)

	return a, err
}

// readFromSecret reads the application from persistent storage (a Kubernetes secret)
// - the secret name/namespace is the same as the application
// - if no application is present in the persistent storage, a new object is created
func (m *ThreadSafeApplicationMap) readFromSecret(application string) (*Application, error) {
	secretNamespace := namespaceFromKey(application)
	secretName := nameFromKey(application) + secretPostfix

	newApplication := NewApplication(application)

	// read secret from cluster; extract appData
	secret, err := k8sclient.Client.Typed().CoreV1().Secrets(secretNamespace).Get(context.Background(), secretName, metav1.GetOptions{})
	if err != nil {
		log.Logger.Debug("no secret backing " + application)
		return newApplication, err
	}

	// read data from secret (is a yaml file)
	rawData, ok := secret.Data[secretKey]
	if !ok {
		log.Logger.Debug("key missing in backing secret")
		return newApplication, errors.New("secret does not contain expected key: " + secretKey)
	}

	a := &Application{}
	err = yaml.Unmarshal(rawData, a)
	if err != nil {
		log.Logger.Debug("unmarshal failure")
		return newApplication, nil
	}

	// set name
	a.Name = application

	// set last write time to read time; it was written in the past
	now := time.Now()
	m.lastWriteTimes[a.Name] = &now

	return a, nil
}

// Write writes the Application to persistent storage (a Kubernetes secret)
func (m *ThreadSafeApplicationMap) Write(a *Application) error {
	log.Logger.Tracef("Write called with %#v", a)
	defer log.Logger.Trace("Write completed")
	var secret *corev1.Secret

	// marshal to byte array
	// note that this uses a custom MarshalJSON that removes untracked
	// versions if the application data is too large
	rawData, err := yaml.Marshal(a)
	if err != nil {
		return err
	}

	secretNamespace := namespaceFromKey(a.Name)
	secretName := secretNameFromKey(a.Name)

	// get the current secret; it will have been created as part of install
	secret, err = k8sclient.Client.Typed().CoreV1().Secrets(secretNamespace).Get(context.Background(), secretName, metav1.GetOptions{})
	if err != nil {
		log.Logger.Error("secret does not exist; no metrics can be recorded")
		return err
	}

	if secret.Data == nil {
		secret.Data = map[string][]byte{}
	}

	secret.Data[secretKey] = rawData
	if secret.StringData != nil {
		secret.StringData[secretKey] = string(rawData)
	}

	// update the secret
	// TBD do we need to merge what we have?
	_, err = k8sclient.Client.Typed().CoreV1().Secrets(secretNamespace).Update(
		context.Background(),
		secret,
		metav1.UpdateOptions{},
	)
	if err != nil {
		log.Logger.WithError(err).Warn("unable to persist application")
		return err
	}

	// update last write time for application
	now := time.Now()
	m.lastWriteTimes[a.Name] = &now
	return nil
}

// BatchedWrite writes the Application to persistent storage only if the previous write
// was more than BatchWriteInterval ago.
func (m *ThreadSafeApplicationMap) BatchedWrite(a *Application) error {
	log.Logger.Tracef("BatchedWrite called")
	defer log.Logger.Trace("BatchedWrite completed")

	now := time.Now()
	lastWrite, ok := m.lastWriteTimes[a.Name]
	if !ok || lastWrite == nil {
		// no record of the application ever being written; write it now
		err := m.Write(a)
		if err != nil {
			log.Logger.Error(err)
			return err
		}
	} else {
		if now.Sub(*m.lastWriteTimes[a.Name]) > BatchWriteInterval {
			err := m.Write(a)
			if err != nil {
				log.Logger.Error(err)
				return err
			}
		}
	}

	// it was written too recently; wait until another write call
	return nil
}

func deleteUntrackedVersions(a *Application) {
	toDelete := []string{}
	for version := range a.Versions {
		track := ""
		for _, ver := range a.Tracks {
			if ver == version {
				track = ver
				break
			}
		}
		if track == "" {
			toDelete = append(toDelete, version)
		}
	}

	for _, version := range toDelete {
		delete(a.Versions, version)
	}
}

// nameFromKey returns the name from a key of the form "namespace/name"
func nameFromKey(applicationKey string) string {
	_, n := splitApplicationKey(applicationKey)
	return n
}

// secretNameFromKey returns the name of the secret used to persist an applicatiob
func secretNameFromKey(applicationKey string) string {
	return nameFromKey(applicationKey) + secretPostfix
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
