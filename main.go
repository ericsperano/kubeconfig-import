package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"sync"

	"gopkg.in/yaml.v2"
)

type ClusterInfo struct {
	CertificateAuthorityData string `yaml:"certificate-authorithy-data"`
	Server                   string `yaml:"server"`
}

type ClusterConfig struct {
	Cluster ClusterInfo `yaml:"cluster"`
	Name    string      `yaml:"name"`
}

type ContextInfo struct {
	Cluster string `yaml:"cluster"`
	User    string `yaml:"user"`
}

type ContextConfig struct {
	Context ContextInfo `yaml:"context"`
	Name    string      `yaml:"name"`
}

type UserInfo struct {
	ClientCertificateData string `yaml:"client-certificate-data"`
	ClientKeyData         string `yaml:"client-key-data"`
}

type UserConfig struct {
	User UserInfo `yaml:"user"`
	Name string   `yaml:"name"`
}

type KubeConfig struct {
	APIVersion     string                      `yaml:"apiVersion"`
	CurrentContext string                      `yaml:"current-context"`
	Kind           string                      `yaml:"kind"`
	Preferences    map[interface{}]interface{} `yaml:"preferences"`
	Clusters       []ClusterConfig             `yaml:"clusters"`
	Contexts       []ContextConfig             `yaml:"contexts"`
	Users          []UserConfig                `yaml:"users"`
}

func ReadConfigToImport() (*KubeConfig, error) {
	bytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return nil, err
	}
	config := KubeConfig{}

	err = yaml.Unmarshal(bytes, &config)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling config to import: %w", err)
	}
	return &config, nil
}

func GetLocalConfigPath() (string, error) {
	dirname, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return path.Join(dirname, ".kube", "config"), nil
}

func ReadLocalConfig() (*KubeConfig, error) {
	p, err := GetLocalConfigPath()
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	bytes, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	config := KubeConfig{}

	err = yaml.Unmarshal(bytes, &config)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling config to import: %w", err)
	}
	return &config, nil
}

func main() {
	configToImport, err := ReadConfigToImport()
	if err != nil {
		log.Fatal(err)
	}
	localConfig, err := ReadLocalConfig()
	if err != nil {
		log.Fatal(err)
	}
	if len(os.Args) != 2 {
		log.Fatal("Need to specify the name to use for the import instead of default")
	}
	configName := os.Args[1]
	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		found := false
		for _, cluster := range localConfig.Clusters {
			if cluster.Name == configName {
				cluster.Cluster.CertificateAuthorityData = configToImport.Clusters[0].Cluster.CertificateAuthorityData
				cluster.Cluster.Server = configToImport.Clusters[0].Cluster.Server
				found = true
			}
		}
		if !found {
			cluster := ClusterConfig{
				Cluster: ClusterInfo{
					CertificateAuthorityData: configToImport.Clusters[0].Cluster.CertificateAuthorityData,
					Server:                   configToImport.Clusters[0].Cluster.Server,
				},
				Name: configName,
			}
			localConfig.Clusters = append(localConfig.Clusters, cluster)
		}
	}()
	go func() {
		defer wg.Done()
		found := false
		for _, context := range localConfig.Contexts {
			if context.Name == configName {
				context.Context.Cluster = configName
				context.Context.User = configName
				found = true
			}
		}
		if !found {
			context := ContextConfig{
				Context: ContextInfo{
					Cluster: configName,
					User:    configName,
				},
				Name: configName,
			}
			localConfig.Contexts = append(localConfig.Contexts, context)
		}
	}()
	go func() {
		defer wg.Done()
		found := false
		for _, user := range localConfig.Users {
			if user.Name == configName {
				user.User.ClientCertificateData = configToImport.Users[0].User.ClientCertificateData
				user.User.ClientKeyData = configToImport.Users[0].User.ClientKeyData
				found = true
			}
		}
		if !found {
			user := UserConfig{
				User: UserInfo{
					ClientCertificateData: configToImport.Users[0].User.ClientCertificateData,
					ClientKeyData:         configToImport.Users[0].User.ClientKeyData,
				},
				Name: configName,
			}
			localConfig.Users = append(localConfig.Users, user)
		}
	}()
	wg.Wait()
	p, err := GetLocalConfigPath()
	if err != nil {
		log.Fatal(err)
	}
	d, err := yaml.Marshal(localConfig)
	if err != nil {
		log.Fatal(err)
	}
	if err := ioutil.WriteFile(p, d, 0644); err != nil {
		log.Fatal(err)
	}
}
