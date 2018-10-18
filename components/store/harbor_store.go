package store

// import (
// 	"bytes"
// 	"encoding/json"
// 	"errors"
// 	"fmt"
// 	"net/http"
// 	"time"

// 	"github.com/subosito/gotenv"
// 	"github.com/turnerlabs/cstore/components/catalog"
// 	"github.com/turnerlabs/cstore/components/contract"
// 	"github.com/turnerlabs/cstore/components/models"
// 	"github.com/turnerlabs/cstore/components/prompt"
// 	"github.com/turnerlabs/cstore/components/token"
// 	harborauth "github.com/turnerlabs/harbor-auth-client"
// )

// const (
// 	authURL = "http://auth.services.dmtio.net"
// 	shipURL = "http://shipit.services.dmtio.net"

// 	tokenToken = "HARBOR_TOKEN"
// 	userToken  = "HARBOR_USER"
// 	passToken  = "HARBOR_PASS"

// 	shipmentToken  = "HARBOR_SHIPMENT"
// 	containerToken = "HARBOR_CONTAINER"
// 	envToken       = "HARBOR_ENV"

// 	modifiedToken  = "CSTORE_MODIFIED"
// 	modifiedLayout = "2006-01-02 15:04:05.999999999 -0700 MST"

// 	envTypeBasic    = "basic"
// 	envTypeDiscover = "discover"
// 	envTypeHidden   = "hidden"
// )

// // HarborStore ...
// type HarborStore struct {
// 	Vault contract.IVault

// 	Auth     HarborAuth
// 	Shipment HarborShipment

// 	io models.IO
// }

// // HarborAuth ...
// type HarborAuth struct {
// 	User  string
// 	Token string
// }

// // HarborShipment ...
// type HarborShipment struct {
// 	Name      string
// 	Container string
// 	Env       string
// }

// // Name ...
// func (s HarborStore) Name() string {
// 	return "harbor"
// }

// // Supports ...
// func (s HarborStore) Supports(feature string) bool {
// 	return false
// }

// // Description ...
// func (s HarborStore) Description() string {
// 	return `Environment variables listed in a .env file can be stored in Harbor at the shipment container level.

// 	When pushing a .env file, a user will be prompted for NT credentails. When the temporary access token expires, the user will be prompted for credentials again.

// 	A shipment, environment, and container are required when using this store to identify which container will store the environment variables.
// `
// }

// // Pre ...
// func (s *HarborStore) Pre(contextID string, file catalog.File, cv contract.IVault, promptUser bool, io models.IO) error {
// 	s.io = io

// 	client, err := harborauth.NewAuthClient(authURL)
// 	if err != nil {
// 		return err
// 	}

// 	s.Shipment = HarborShipment{}
// 	s.Auth = HarborAuth{}

// 	isAuth := false

// 	// Argonauts Login ID
// 	if value, err := cv.Get(contextID, userToken, ""); err == nil {
// 		s.Auth.User = value
// 	}

// 	if value, err := cv.Get(contextID, tokenToken, ""); err == nil {
// 		s.Auth.Token = value
// 	}

// 	if len(s.Auth.Token) > 0 && len(s.Auth.User) > 0 {
// 		isAuth, _ = client.IsAuthenticated(s.Auth.User, s.Auth.Token)
// 	}

// 	if !isAuth {
// 		// Argonauts Login ID
// 		s.Auth.User = prompt.GetValFromUser(userToken, prompt.Options{}, io)
// 		pass := prompt.GetValFromUser(passToken, prompt.Options{HideInput: true}, s.io)

// 		token, success, err := client.Login(s.Auth.User, pass)
// 		if err != nil || !success {
// 			return err
// 		}

// 		err = cv.Set(contextID, userToken, "", s.Auth.User)
// 		if err != nil {
// 			return err
// 		}

// 		err = cv.Set(contextID, tokenToken, "", token)
// 		if err != nil {
// 			return err
// 		}

// 		s.Auth.Token = token
// 	}

// 	if shipment, found := file.Data[shipmentToken]; found {
// 		s.Shipment.Name = shipment
// 	} else {
// 		s.Shipment.Name = prompt.GetValFromUser(shipmentToken, prompt.Options{}, s.io)
// 	}

// 	if container, found := file.Data[containerToken]; found {
// 		s.Shipment.Container = container
// 	} else {
// 		s.Shipment.Container = prompt.GetValFromUser(containerToken, prompt.Options{}, s.io)
// 	}

// 	if env, found := file.Data[envToken]; found {
// 		s.Shipment.Env = env
// 	} else {
// 		s.Shipment.Env = prompt.GetValFromUser(envToken, prompt.Options{}, s.io)
// 	}

// 	return nil
// }

// // Push ...
// func (s HarborStore) Push(file catalog.File, fileData []byte, version string) (map[string]string, error) {

// 	if !file.IsEnv {
// 		return map[string]string{}, fmt.Errorf("cannot store file %s", file.Path)
// 	}

// 	data := map[string]string{
// 		shipmentToken:  s.Shipment.Name,
// 		containerToken: s.Shipment.Container,
// 		envToken:       s.Shipment.Env,
// 	}

// 	localKeys := gotenv.Parse(bytes.NewReader(fileData))
// 	localKeys[modifiedToken] = time.Now().UTC().String()

// 	url := buildURL(s.Shipment)

// 	for key, value := range localKeys {

// 		prefixedKey := addEnvVarPrefix(key)

// 		keyType := envTypeHidden
// 		if storedKeyType, found := file.Data[prefixedKey]; found {
// 			if storedKeyType != envVarType {
// 				keyType = storedKeyType
// 			}
// 		}

// 		p := pair{
// 			Name:  key,
// 			Value: value,
// 			Type:  keyType,
// 		}

// 		if err := createKey(p, url, s.Auth); err != nil {
// 			if err := updateKey(p, url, s.Auth); err != nil {
// 				return data, err
// 			}
// 		}

// 		data[prefixedKey] = keyType
// 	}

// 	harborKeys, err := getHarborKeys(s.Shipment, s.Auth)
// 	if err != nil {
// 		return data, err
// 	}

// 	for key := range harborKeys {
// 		prefixedKey := addEnvVarPrefix(key)

// 		if _, found := file.Data[prefixedKey]; found {
// 			if _, found := localKeys[key]; !found {
// 				if err := deleteKey(key, url, s.Auth); err != nil {
// 					return data, err
// 				}
// 			}
// 		}
// 	}

// 	return data, nil
// }

// // Pull ...
// func (s HarborStore) Pull(file catalog.File, version string) ([]byte, error) {

// 	keys, err := getHarborKeys(s.Shipment, s.Auth)
// 	if err != nil {
// 		return []byte{}, err
// 	}

// 	var buffer bytes.Buffer

// 	for key, contents := range keys {
// 		if key == modifiedToken {
// 			continue
// 		}

// 		if _, found := file.Data[addEnvVarPrefix(key)]; found {
// 			buffer.WriteString(fmt.Sprintf("%s=%s\n", key, contents.value))
// 		}
// 	}

// 	// if modified, found := keys[modifiedToken]; found {
// 	// 	m, err := time.Parse(modifiedLayout, modified.value)
// 	// 	if err == nil {
// 	// 		attr.LastModified = m
// 	// 	}
// 	// }

// 	return buffer.Bytes(), nil
// }

// // Purge ...
// func (s HarborStore) Purge(file catalog.File, version string) error {

// 	url := buildURL(s.Shipment)

// 	for key, value := range file.Data {
// 		if isEnvVarType(value) {
// 			if err := deleteKey(key, url, s.Auth); err != nil {
// 				return err
// 			}
// 		}
// 	}

// 	return nil
// }

// // GetTokenValues ...
// func (s HarborStore) GetTokenValues(tokens map[string]token.Token, contextID string) (map[string]token.Token, error) {
// 	return map[string]token.Token{}, nil
// }

// // SaveTokenValues ...
// func (s HarborStore) SaveTokenValues(tokens map[string]token.Token, contextID string) (map[string]token.Token, error) {
// 	return map[string]token.Token{}, nil
// }

// func isEnvVarType(envVarType string) bool {
// 	switch envVarType {
// 	case envTypeBasic:
// 		return true
// 	case envTypeDiscover:
// 		return true
// 	case envTypeHidden:
// 		return true
// 	default:
// 		return false
// 	}
// }

// type pair struct {
// 	Name  string `json:"name"`
// 	Value string `json:"value"`
// 	Type  string `json:"type"`
// }

// func createKey(p pair, url string, auth HarborAuth) error {
// 	client := &http.Client{}

// 	b, err := json.Marshal(p)
// 	if err != nil {
// 		return err
// 	}

// 	url = fmt.Sprintf("%s/envVars", url)

// 	r := bytes.NewReader(b)

// 	req, err := http.NewRequest("POST", url, r)
// 	req.Header.Add("x-token", auth.Token)
// 	req.Header.Add("x-username", auth.User)
// 	req.Header.Add("Content-Type", "application/json")

// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return err
// 	}

// 	if resp.StatusCode != http.StatusCreated {
// 		return errors.New(resp.Status)
// 	}

// 	return nil
// }

// func updateKey(p pair, url string, auth HarborAuth) error {

// 	client := &http.Client{}

// 	b, err := json.Marshal(p)
// 	if err != nil {
// 		return err
// 	}

// 	url = fmt.Sprintf("%s/envVar/%s", url, p.Name)

// 	r := bytes.NewReader(b)

// 	req, err := http.NewRequest("PUT", url, r)
// 	req.Header.Add("x-token", auth.Token)
// 	req.Header.Add("x-username", auth.User)
// 	req.Header.Add("Content-Type", "application/json")

// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return err
// 	}

// 	if resp.StatusCode != http.StatusOK {
// 		return errors.New(resp.Status)
// 	}

// 	return nil
// }

// func deleteKey(key, url string, auth HarborAuth) error {

// 	client := &http.Client{}

// 	url = fmt.Sprintf("%s/envVar/%s", url, key)

// 	req, err := http.NewRequest("DELETE", url, nil)
// 	req.Header.Add("x-token", auth.Token)
// 	req.Header.Add("x-username", auth.User)
// 	req.Header.Add("Content-Type", "application/json")

// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return err
// 	}

// 	if resp.StatusCode != http.StatusOK {
// 		return errors.New(resp.Status)
// 	}

// 	return nil
// }

// type harborKey struct {
// 	value string
// 	vType string
// }

// func getHarborKeys(shipment HarborShipment, auth HarborAuth) (map[string]harborKey, error) {

// 	client := &http.Client{}

// 	url := fmt.Sprintf("%s/v1/shipment/%s/environment/%s", shipURL, shipment.Name, shipment.Env)

// 	req, err := http.NewRequest("GET", url, nil)
// 	req.Header.Add("x-token", auth.Token)
// 	req.Header.Add("x-username", auth.User)
// 	req.Header.Add("Content-Type", "application/json")

// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return nil, err
// 	}

// 	if resp.StatusCode != http.StatusOK {
// 		return nil, errors.New(resp.Status)
// 	}

// 	s := new(HShipment)
// 	if err = json.NewDecoder(resp.Body).Decode(s); err != nil {
// 		return nil, err
// 	}

// 	envVars := map[string]harborKey{}

// 	for _, c := range s.Containers {
// 		if c.Name == shipment.Container {
// 			for _, envVar := range c.EnvVars {
// 				envVars[envVar.Name] = harborKey{
// 					value: envVar.Value,
// 					vType: envVar.Type,
// 				}
// 			}
// 		}
// 	}

// 	return envVars, nil
// }

// // HShipment ...
// type HShipment struct {
// 	Containers []HContainers `json:"containers"`
// }

// // HContainers ...
// type HContainers struct {
// 	Name    string    `json:"name"`
// 	EnvVars []HEnvVar `json:"envVars"`
// }

// // HEnvVar ...
// type HEnvVar struct {
// 	Name  string `json:"name"`
// 	Value string `json:"value"`
// 	Type  string `json:"type"`
// }

// func buildURL(shipment HarborShipment) string {
// 	return fmt.Sprintf("%s/v1/shipment/%s/environment/%s/container/%s", shipURL, shipment.Name, shipment.Env, shipment.Container)
// }

// func init() {
// 	//--------------------------------
// 	//- Disabled until converted to v2
// 	//--------------------------------
// 	//s := new(HarborStore)
// 	//stores[s.Name()] = s
// }
