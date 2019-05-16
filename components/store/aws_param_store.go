package store

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/subosito/gotenv"
	"github.com/turnerlabs/cstore/components/catalog"
	"github.com/turnerlabs/cstore/components/cfg"
	"github.com/turnerlabs/cstore/components/cipher"
	"github.com/turnerlabs/cstore/components/contract"
	"github.com/turnerlabs/cstore/components/models"
	"github.com/turnerlabs/cstore/components/prompt"
	"github.com/turnerlabs/cstore/components/setting"
	"github.com/turnerlabs/cstore/components/vault"
)

const (
	configType = "CONFIG"

	cmdRefFormat = "refs"
)

// AWSParameterStore ...
type AWSParameterStore struct {
	Session *session.Session

	context  string
	settings map[string]setting.Setting

	encryptionType string
	credentialType string

	uo cfg.UserOptions

	io models.IO
}

// Name ...
func (s AWSParameterStore) Name() string {
	return "aws-parameter"
}

// Supports ...
func (s AWSParameterStore) Supports(feature string) bool {
	switch feature {
	case VersionFeature:
		return true
	default:
		return false
	}
}

// Description ...
func (s AWSParameterStore) Description() string {
	return `
	detail: https://github.com/turnerlabs/cstore/blob/master/docs/PARAMETER.md
`
}

// Pre ...
func (s *AWSParameterStore) Pre(clog catalog.Catalog, file *catalog.File, access contract.IVault, uo cfg.UserOptions, io models.IO) error {
	s.settings = map[string]setting.Setting{}
	s.context = clog.Context
	s.uo = uo
	s.io = io

	s.credentialType = autoDetect
	s.encryptionType = getEncryptionType(*file)

	(setting.Setting{
		Group:        "AWS",
		Prop:         "REGION",
		Prompt:       uo.Prompt,
		Set:          true,
		DefaultValue: awsDefaultRegion,
		Vault:        vault.EnvVault{},
	}).Get(clog.Context, io)

	//---------------------------------------------
	//- Store authentication and encryption options
	//---------------------------------------------
	if uo.Prompt {
		s.credentialType = strings.ToLower(prompt.GetValFromUser("Authentication", prompt.Options{
			Description:  "OPTIONS\n (P)rofile \n (U)ser",
			DefaultValue: "P"}, io))
	}

	//------------------------------------------
	//- Required auth creds
	//------------------------------------------
	switch s.credentialType {
	case cTypeProfile:
		os.Unsetenv(awsSecretAccessKey)
		os.Unsetenv(awsAccessKeyID)

		(setting.Setting{
			Group:        "AWS",
			Prop:         "PROFILE",
			DefaultValue: os.Getenv(awsProfile),
			Prompt:       uo.Prompt,
			Set:          true,
			Vault:        vault.EnvVault{},
		}).Get(clog.Context, io)

	case cTypeUser:
		os.Unsetenv(awsProfile)

		(setting.Setting{
			Group:  "AWS",
			Prop:   "ACCESS_KEY_ID",
			Prompt: uo.Prompt,
			Set:    true,
			Vault:  access,
			Stage:  vault.EnvVault{},
		}).Get(clog.Context, io)

		(setting.Setting{
			Group:  "AWS",
			Prop:   "SECRET_ACCESS_KEY",
			Prompt: uo.Prompt,
			Set:    true,
			Vault:  access,
			Stage:  vault.EnvVault{},
		}).Get(clog.Context, io)
	}

	//------------------------------------------
	//- Optional encryption
	//------------------------------------------
	if uo.Prompt {
		s.encryptionType = strings.ToLower(prompt.GetValFromUser("Encryption", prompt.Options{
			DefaultValue: strings.ToUpper(s.encryptionType),
			Description:  "OPTIONS\n (C)lient - 16 or 32 character encryption key \n (S)erver - override S3 Bucket KMS Key ID \n (N)one"}, io))
	}

	switch s.encryptionType {
	case eTypeClient:

		s.settings[clientEncryptionToken] = setting.Setting{
			Description:  "Specify a 16 or 32 bit encryption key. Save the key somewhere secure to decrypt the files later.",
			Group:        fmt.Sprintf("CSTORE_%s", strings.ToUpper(s.context)),
			Prop:         fmt.Sprintf("ENCRYPTION_KEY_%s", strings.ToUpper(file.Key())),
			DefaultValue: cipher.GenKey(32),
			Prompt:       uo.Prompt,
			Set:          true,
			Vault:        access,
		}

	case eTypeServer:

		s.settings[serverEncryptionToken] = setting.Setting{
			Description: "Specify the AWS KMS Key ID to use for server side encryption.",
			Group:       "AWS",
			Prop:        "KMS_KEY_ID",
			Prompt:      uo.Prompt,
			Set:         true,
			Vault:       access,
		}

	}

	//------------------------------------------
	//- Open connection to store.
	//------------------------------------------
	sess, err := session.NewSession()
	if err != nil {
		return err
	}

	s.Session = sess

	return err
}

// Push ...
func (s AWSParameterStore) Push(file *catalog.File, fileData []byte, version string) error {

	if !file.SupportsConfig() {
		return fmt.Errorf("store does not support file type: %s", file.Type)
	}

	input := ssm.PutParameterInput{
		Overwrite: aws.Bool(true),
		Type:      aws.String(ssm.ParameterTypeString),
	}

	//------------------------------------------
	//- Get encryption keys
	//------------------------------------------
	key, serverEncryption := s.settings[serverEncryptionToken]
	if serverEncryption {
		value, err := key.Get(s.context, s.io)
		if err != nil {
			return err
		}

		input.KeyId = &value
		input.Type = aws.String(ssm.ParameterTypeSecureString)
	}

	clientEncryptionKey := ""
	if key, clientEncryption := s.settings[clientEncryptionToken]; clientEncryption {
		value, err := key.Get(s.context, s.io)
		if err != nil {
			return err
		}
		clientEncryptionKey = value

		file.AddData(map[string]string{
			fileDataEncryptionKey: clientEncryptionToken,
		})
	}

	//------------------------------------------
	//- Push configuration
	//------------------------------------------
	newParams := gotenv.Parse(bytes.NewReader(fileData))

	svc := ssm.New(s.Session)

	storedParams, err := getStoredParams(s.context, file.Path, version, svc)
	if err != nil {
		return err
	}

	for name, value := range newParams {
		n := name
		v := value

		remoteKey := buildRemoteKey(s.context, file.Path, n, version)

		if noChange(remoteKey, value, storedParams) {
			continue
		}

		if len(clientEncryptionKey) > 0 {
			b, err := cipher.Encrypt(clientEncryptionKey, []byte(value))
			if err != nil {
				return err
			}
			v = hex.EncodeToString(b)
		}

		v = formatValue(v)

		input.Name = &remoteKey
		input.Value = &v

		_, err := svc.PutParameter(&input)
		if err != nil {
			fmt.Fprintf(s.io.UserOutput, "parameter: %s", remoteKey)
			return err
		}
	}

	//------------------------------------------
	//- Delete no longer relavant stored params
	//------------------------------------------
	for _, remoteParam := range storedParams {

		param := strings.Replace(*remoteParam.Name, buildRemotePath(s.context, file.Path, version)+"/", "", 1)

		if !isParamIn(param, newParams) {
			if _, err := svc.DeleteParameter(&ssm.DeleteParameterInput{
				Name: remoteParam.Name,
			}); err != nil {
				fmt.Fprintf(s.io.UserOutput, "parameter: %s", *remoteParam.Name)
				return err
			}
		}
	}

	return nil
}

// Pull ...
func (s AWSParameterStore) Pull(file *catalog.File, version string) ([]byte, contract.Attributes, error) {

	clientEncryptionKey := ""
	if key, clientEncryption := s.settings[clientEncryptionToken]; clientEncryption {
		value, err := key.Get(s.context, s.io)
		if err != nil {
			return []byte{}, contract.Attributes{}, err
		}
		clientEncryptionKey = value
	}

	svc := ssm.New(s.Session)

	storedParams, err := getStoredParams(s.context, file.Path, version, svc)
	if err != nil {
		return []byte{}, contract.Attributes{}, err
	}

	if len(storedParams) == 0 {
		return []byte{}, contract.Attributes{}, nil
	}

	var buffer bytes.Buffer

	for key, value := range toMap(storedParams) {
		name := key[strings.LastIndex(key, "/")+1 : len(key)]
		v := value

		if len(clientEncryptionKey) > 0 {
			b, err := hex.DecodeString(value)
			if err != nil {
				return []byte{}, contract.Attributes{}, err
			}

			b, err = cipher.Decrypt(clientEncryptionKey, b)
			if err != nil {
				return b, contract.Attributes{}, err
			}

			v = string(b)
		}

		if s.uo.StoreCommand == cmdRefFormat {
			buffer.WriteString(fmt.Sprintf("%s=%s\n", name, buildRemoteKey(s.context, file.Path, name, version)))
		} else {
			buffer.WriteString(fmt.Sprintf("%s=%s\n", name, v))
		}
	}

	return buffer.Bytes(), contract.Attributes{
		LastModified: lastModified(storedParams),
	}, nil
}

// Purge ...
func (s AWSParameterStore) Purge(file *catalog.File, version string) error {

	svc := ssm.New(s.Session)

	storedParams, err := getStoredParams(s.context, file.Path, version, svc)
	if err != nil {
		return err
	}

	for _, p := range storedParams {
		if _, err := svc.DeleteParameter(&ssm.DeleteParameterInput{
			Name: p.Name,
		}); err != nil {
			fmt.Fprintf(s.io.UserOutput, "parameter: %s", *p.Name)
			return err
		}
	}

	return nil
}

// Changed ...
func (s AWSParameterStore) Changed(file *catalog.File, fileData []byte, version string) (time.Time, error) {
	config := gotenv.Parse(bytes.NewReader(fileData))

	clientEncryptionKey := ""
	if key, clientEncryption := s.settings[clientEncryptionToken]; clientEncryption {
		value, err := key.Get(s.context, s.io)
		if err != nil {
			return time.Time{}, err
		}
		clientEncryptionKey = value
	}

	svc := ssm.New(s.Session)

	storedParams, err := getStoredParams(s.context, file.Path, version, svc)

	if err != nil {

		return time.Time{}, err
	}

	changedParams := []*ssm.Parameter{}
	for _, p := range storedParams {

		for name, value := range config {
			remoteKey := buildRemoteKey(s.context, file.Path, name, version)

			decryptedValue := *p.Value

			if len(clientEncryptionKey) > 0 {
				b, err := hex.DecodeString(value)
				if err != nil {
					return time.Time{}, err
				}

				b, err = cipher.Decrypt(clientEncryptionKey, b)
				if err != nil {
					return time.Time{}, err
				}

				decryptedValue = string(b)
			}

			if remoteKey == *p.Name && value != decryptedValue {
				changedParams = append(changedParams, p)
			}
		}
	}

	return lastModified(changedParams), nil
}

func lastModified(params []*ssm.Parameter) time.Time {
	mostRecentlyModified := time.Time{}
	for _, sp := range params {
		if mostRecentlyModified.Before(*sp.LastModifiedDate) {
			mostRecentlyModified = *sp.LastModifiedDate
		}
	}

	return mostRecentlyModified
}

func listStoredParams(svc *ssm.SSM, startsWith string) ([]*ssm.ParameterMetadata, error) {
	params, _ := describeParams(svc, startsWith, "", []*ssm.ParameterMetadata{})

	return params, nil
}

func describeParams(svc *ssm.SSM, startsWith string, nextToken string, params []*ssm.ParameterMetadata) ([]*ssm.ParameterMetadata, error) {
	filters := []*ssm.ParameterStringFilter{
		&ssm.ParameterStringFilter{
			Key:    aws.String(ssm.ParametersFilterKeyName),
			Option: aws.String("BeginsWith"),
			Values: aws.StringSlice([]string{startsWith}),
		},
	}

	input := &ssm.DescribeParametersInput{
		ParameterFilters: filters,
	}

	if len(nextToken) > 0 {
		input.SetNextToken(nextToken)
	}

	output, err := svc.DescribeParameters(input)
	if err != nil {
		return nil, err
	}

	params = append(params, output.Parameters...)

	if output.NextToken == nil || len(*output.NextToken) == 0 {
		return params, nil
	}

	return describeParams(svc, startsWith, *output.NextToken, params)
}

func formatValue(value string) string {
	const tokenRegexStr = `{{(([\w\d\/-]+))}}`

	var r = regexp.MustCompile(tokenRegexStr)

	matches := r.FindAllStringSubmatch(value, -1)

	if matches == nil {
		return value
	}

	for _, sm := range matches {
		value = strings.Replace(value, sm[0], fmt.Sprintf("<<%s>>", sm[1]), -1)
	}

	return value
}

func unformatValue(value string) string {
	const tokenRegexStr = `<<(([\w\d\/-]+))>>`

	var r = regexp.MustCompile(tokenRegexStr)

	matches := r.FindAllStringSubmatch(value, -1)

	if matches == nil {
		return value
	}

	for _, sm := range matches {
		value = strings.Replace(value, sm[0], fmt.Sprintf("{{%s}}", sm[1]), -1)
	}

	return value
}

func get(params []*string, svc *ssm.SSM) ([]*ssm.Parameter, error) {

	if len(params) == 0 {
		return []*ssm.Parameter{}, nil
	}

	storedParams := []*ssm.Parameter{}

	// AWS Golang SDK request limit: 10
	chuckedParams := []*string{}

	for i := 0; i < len(params); i++ {

		chuckedParams = append(chuckedParams, params[i])

		if len(chuckedParams) == 10 || i == len(params)-1 {

			output, err := svc.GetParameters(&ssm.GetParametersInput{
				Names:          chuckedParams,
				WithDecryption: aws.Bool(true),
			})
			if err != nil {
				return []*ssm.Parameter{}, err
			}

			storedParams = append(storedParams, output.Parameters...)

			chuckedParams = []*string{}
		}
	}

	return storedParams, nil
}

func toMap(params []*ssm.Parameter) map[string]string {
	data := map[string]string{}

	for _, p := range params {
		data[*p.Name] = *p.Value
	}

	return data
}

func isParamIn(param string, params gotenv.Env) bool {
	for name := range params {
		if name == param {
			return true
		}
	}
	return false
}

func noChange(key, value string, params []*ssm.Parameter) bool {
	for _, p := range params {
		if *p.Name == key {
			return *p.Value == value
		}
	}

	return false
}

func buildRemoteKey(context, path, name, version string) string {
	if len(version) > 0 {
		return fmt.Sprintf("/%s/%s/%s/%s", context, version, path, name)
	}
	return fmt.Sprintf("/%s/%s/%s", context, path, name)
}

func buildRemotePath(context, path, version string) string {
	if len(version) > 0 {
		return fmt.Sprintf("/%s/%s/%s", context, version, path)
	}
	return fmt.Sprintf("/%s/%s", context, path)
}

// Returns a snapshot of previously pushed config
func getStoredParams(context, path, version string, svc *ssm.SSM) ([]*ssm.Parameter, error) {

	storedParamData, err := listStoredParams(svc, buildRemotePath(context, path, version))
	if err != nil {
		return nil, err
	}

	params := []*string{}
	for _, p := range storedParamData {
		params = append(params, p.Name)
	}

	storedParams, err := get(params, svc)
	if err != nil {
		return nil, err
	}

	for _, sp := range storedParams {
		sp.Value = aws.String(unformatValue(*sp.Value))
	}

	return storedParams, nil
}

func init() {
	s := new(AWSParameterStore)
	stores[s.Name()] = s
}
