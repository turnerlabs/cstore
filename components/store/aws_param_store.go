package store

// import (
// 	"bytes"
// 	"encoding/hex"
// 	"errors"
// 	"fmt"
// 	"strings"
// 	"time"

// 	"github.com/aws/aws-sdk-go/aws/session"
// 	"github.com/aws/aws-sdk-go/service/ssm"
// 	"github.com/subosito/gotenv"
// 	"github.com/turnerlabs/cstore/components/catalog"
// 	"github.com/turnerlabs/cstore/components/cipher"
// 	"github.com/turnerlabs/cstore/components/contract"
// 	"github.com/turnerlabs/cstore/components/models"
// 	"github.com/turnerlabs/cstore/components/token"
// )

// // AWSParameterStore ...
// type AWSParameterStore struct {
// 	Vault   contract.IVault
// 	Session *session.Session
// 	Service *ssm.SSM

// 	CSKey        string
// 	CSEncryption bool

// 	SSKMSKeyID   string
// 	SSEncryption bool

// 	EncryptionEnabled bool

// 	context string

// 	io models.IO
// }

// // Name ...
// func (s AWSParameterStore) Name() string {
// 	return "aws-parameter"
// }

// // Supports ...
// func (s AWSParameterStore) Supports(feature string) bool {
// 	return false
// }

// // Description ...
// func (s AWSParameterStore) Description() string {
// 	return "" //fmt.Sprintf(description, "Only environment variables listed in a .env file can be stored in the AWS Parameter Store.", awsAccessKeyID, awsSecretAccessKey, awsAccessKeyID, awsSecretAccessKey, awsDefaultProfile)
// }

// // Pre ...
// func (s *AWSParameterStore) Pre(context string, file *catalog.File, cv contract.IVault, promptUser bool, io models.IO) error {
// 	s.context = context
// 	s.io = io
// 	// sess, clientKey, KMSKeyID, err := setupAWS(contextID, file, cv, promptUser)

// 	// s.Session = sess
// 	// s.Service = ssm.New(sess)

// 	// if len(clientKey) > 0 {
// 	// 	s.CSKey = clientKey
// 	// 	s.CSEncryption = true
// 	// 	s.EncryptionEnabled = true
// 	// }

// 	// if len(KMSKeyID) > 0 {
// 	// 	s.SSKMSKeyID = KMSKeyID
// 	// 	s.SSEncryption = true
// 	// 	s.EncryptionEnabled = true
// 	// }

// 	return nil
// }

// // Push ...
// func (s AWSParameterStore) Push(file *catalog.File, fileData []byte, version string) (map[string]string, error) {

// 	contextKey := file.ContextKey(s.context)

// 	if !file.IsEnv {
// 		return map[string]string{}, fmt.Errorf("cannot store file %s", file.Path)
// 	}

// 	pairs := gotenv.Parse(bytes.NewReader(fileData))

// 	overwrite := true
// 	paramType := "String"

// 	input := ssm.PutParameterInput{
// 		Overwrite: &overwrite,
// 		Type:      &paramType,
// 	}

// 	if s.SSEncryption {
// 		paramType = "SecureString"
// 		input.KeyId = &s.SSKMSKeyID
// 	}

// 	storedParams, err := getStoreParams(s.Service, file.Data, s.SSEncryption, contextKey)
// 	if err != nil {
// 		return nil, err
// 	}

// 	params := []string{}
// 	for name, value := range pairs {
// 		v := value

// 		if s.CSEncryption {
// 			b, err := cipher.Encrypt(s.CSKey, []byte(value))
// 			if err != nil {
// 				return nil, err
// 			}
// 			v = hex.EncodeToString(b)
// 		}

// 		keyID := buildParamKey(contextKey, name)

// 		input.Name = &keyID
// 		input.Value = &v

// 		_, err := s.Service.PutParameter(&input)
// 		if err != nil {
// 			return nil, err
// 		}

// 		params = append(params, keyID)
// 	}

// 	for name := range storedParams {
// 		if !isParamIn(params, name) {

// 			input := ssm.DeleteParameterInput{
// 				Name: &name,
// 			}

// 			_, err := s.Service.DeleteParameter(&input)
// 			if err != nil {
// 				return nil, err
// 			}
// 		}
// 	}

// 	return toMap(params), nil
// }

// // Pull ...
// func (s AWSParameterStore) Pull(file *catalog.File, version string) ([]byte, error) {

// 	contextKey := file.ContextKey(s.context)

// 	storedParams, err := getStoreParams(s.Service, file.Data, s.SSEncryption, contextKey)
// 	if err != nil {
// 		return nil, err
// 	}

// 	if len(storedParams) == 0 {
// 		return []byte{}, nil
// 	}

// 	var buffer bytes.Buffer
// 	//lastModKey := ""

// 	for key, value := range storedParams {
// 		//lastModKey = key

// 		name := key[strings.LastIndex(key, "/")+1 : len(key)]
// 		v := value

// 		if s.CSEncryption {
// 			b, err := hex.DecodeString(value)
// 			if err != nil {
// 				return nil, err
// 			}

// 			b, err = cipher.Decrypt(s.CSKey, b)
// 			if err != nil {
// 				return b, err
// 			}

// 			v = string(b)
// 		}

// 		buffer.WriteString(fmt.Sprintf("%s=%s\n", name, v))
// 	}

// 	// modified, err := getLastModified(s.Service, lastModKey)
// 	// if err != nil {
// 	// 	return nil, err
// 	// }

// 	return buffer.Bytes(), nil
// }

// // Purge ...
// func (s AWSParameterStore) Purge(file *catalog.File, version string) error {

// 	for key := range file.Data {
// 		input := ssm.DeleteParameterInput{
// 			Name: &key,
// 		}

// 		_, err := s.Service.DeleteParameter(&input)
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }

// // GetTokenValues ...
// func (s AWSParameterStore) GetTokenValues(tokens map[string]token.Token, contextID string) (map[string]token.Token, error) {
// 	return map[string]token.Token{}, nil
// }

// // SaveTokenValues ...
// func (s AWSParameterStore) SaveTokenValues(tokens map[string]token.Token, contextID string) (map[string]token.Token, error) {
// 	return map[string]token.Token{}, nil
// }

// func getStoreParams(svc *ssm.SSM, data map[string]string, ssEncryption bool, contextKey string) (map[string]string, error) {
// 	dataParams := []*string{}
// 	for dataParam := range data {
// 		dataParams = append(dataParams, &dataParam)
// 	}

// 	if len(dataParams) == 0 {
// 		return map[string]string{}, nil
// 	}

// 	storedParams := map[string]string{}

// 	// AWS Golang SDK limits requests to a ten param limit.
// 	for start := 0; start <= len(dataParams); start += 10 {
// 		end := start + 9
// 		if end > len(dataParams)-1 {
// 			end = len(dataParams)
// 		}

// 		subset := dataParams[start:end]
// 		sp, err := getParams(svc, subset, ssEncryption, contextKey)
// 		if err != nil {
// 			return nil, err
// 		}

// 		for key, value := range sp {
// 			storedParams[key] = value
// 		}
// 	}

// 	return storedParams, nil
// }

// // Since all the params are pushed every time, it does not matter
// // which one is used for the last modified date time. If this store
// // intelligently pushes params, this method will likely need to get
// // the most recently edited params date time.
// func getLastModified(svc *ssm.SSM, key string) (time.Time, error) {

// 	name := "Name"

// 	filter := ssm.ParametersFilter{
// 		Key:    &name,
// 		Values: []*string{&key},
// 	}

// 	input := ssm.DescribeParametersInput{
// 		Filters: []*ssm.ParametersFilter{&filter},
// 	}

// 	output, err := svc.DescribeParameters(&input)
// 	if err != nil {
// 		return time.Time{}, err
// 	}

// 	if len(output.Parameters) == 0 {
// 		return time.Time{}, errors.New("Failed to get last modified time.")
// 	}

// 	return *output.Parameters[0].LastModifiedDate, nil
// }

// func getParams(svc *ssm.SSM, names []*string, ssEncryption bool, contextKey string) (stored map[string]string, err error) {

// 	input := ssm.GetParametersInput{
// 		Names:          names,
// 		WithDecryption: &ssEncryption,
// 	}

// 	output, err := svc.GetParameters(&input)
// 	if err != nil {
// 		return nil, err
// 	}

// 	storedParams := map[string]string{}
// 	for _, value := range output.Parameters {
// 		storedParams[*value.Name] = *value.Value
// 	}

// 	return storedParams, nil
// }

// func toMap(params []string) (data map[string]string) {
// 	for _, value := range params {
// 		data[value] = envVarType
// 	}

// 	return data
// }

// func isParamIn(params []string, param string) bool {
// 	for _, value := range params {
// 		if value == param {
// 			return true
// 		}
// 	}
// 	return false
// }

// func buildParamKey(key, name string) string {
// 	return fmt.Sprintf("%s/%s", buildParamPath(key), name)
// }

// func buildParamPath(key string) string {
// 	return fmt.Sprintf("/cstore/%s", key)
// }

// func init() {
// 	//--------------------------------
// 	//- Disabled until converted to v2
// 	//--------------------------------
// 	//s := new(AWSParameterStore)
// 	//stores[s.Name()] = s
// }
