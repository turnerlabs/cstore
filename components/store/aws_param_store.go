package store

import (
	"bytes"
	"errors"
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
	"github.com/turnerlabs/cstore/components/contract"
	"github.com/turnerlabs/cstore/components/models"
	"github.com/turnerlabs/cstore/components/prompt"
	"github.com/turnerlabs/cstore/components/setting"
	"github.com/turnerlabs/cstore/components/vault"
)

const (
	configType = "CONFIG"

	cmdRefFormat = "refs"

	defaultPSKMSKey = "aws/ssm"
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

// SupportsFeature ...
func (s AWSParameterStore) SupportsFeature(feature string) bool {
	switch feature {
	case VersionFeature:
		return true
	default:
		return false
	}
}

// SupportsFileType ...
func (s AWSParameterStore) SupportsFileType(fileType string) bool {
	switch fileType {
	case EnvFeature:
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
		Silent:       uo.Silent,
		AutoSave:     true,
		DefaultValue: awsDefaultRegion,
		Vault:        vault.EnvVault{},
	}).Get(clog.Context, io)

	//------------------------------------------
	//- Auth Credentials
	//------------------------------------------
	if uo.Prompt {
		s.credentialType = strings.ToLower(prompt.GetValFromUser("Authentication", prompt.Options{
			Description:  "OPTIONS\n (P)rofile \n (U)ser",
			DefaultValue: "P"}, io))
	}

	switch s.credentialType {
	case cTypeProfile:
		os.Unsetenv(awsSecretAccessKey)
		os.Unsetenv(awsAccessKeyID)

		(setting.Setting{
			Group:        "AWS",
			Prop:         "PROFILE",
			DefaultValue: os.Getenv(awsProfile),
			Prompt:       uo.Prompt,
			Silent:       uo.Silent,
			AutoSave:     true,
			Vault:        vault.EnvVault{},
		}).Get(clog.Context, io)

	case cTypeUser:
		os.Unsetenv(awsProfile)

		(setting.Setting{
			Group:    "AWS",
			Prop:     "ACCESS_KEY_ID",
			Prompt:   uo.Prompt,
			Silent:   uo.Silent,
			AutoSave: true,
			Vault:    access,
		}).Get(clog.Context, io)

		(setting.Setting{
			Group:    "AWS",
			Prop:     "SECRET_ACCESS_KEY",
			Prompt:   uo.Prompt,
			Silent:   uo.Silent,
			AutoSave: true,
			Vault:    access,
		}).Get(clog.Context, io)
	}

	//------------------------------------------
	//- Encryption
	//------------------------------------------
	s.settings[serverEncryptionToken] = setting.Setting{
		Description:  "KMS Key ID is used by Parameter Store to encrypt and decrypt secrets. Any role or user accessing a secret must also have access to the KMS key. When pushing updates, the default setting will preserve existing KMS keys. The aws/ssm key is the default Systems Manager KMS key.",
		Group:        "AWS",
		Prop:         "STORE_KMS_KEY_ID",
		DefaultValue: clog.GetAnyDataBy("AWS_STORE_KMS_KEY_ID", defaultPSKMSKey),
		Prompt:       uo.Prompt,
		Silent:       uo.Silent,
		AutoSave:     false,
		Vault:        file,
	}

	//------------------------------------------
	//- Open Connection
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

	if len(fileData) == 0 {
		return errors.New("empty file")
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

		if value != defaultPSKMSKey {
			input.KeyId = &value
		}

		input.Type = aws.String(ssm.ParameterTypeSecureString)
	}

	//------------------------------------------
	//- Push configuration
	//------------------------------------------
	newParams := gotenv.Parse(bytes.NewReader(fileData))
	if len(newParams) == 0 {
		return errors.New("failed to parse environment variables")
	}

	svc := ssm.New(s.Session)

	storedParams, err := getStoredParamsWithMetaData(s.context, file.Path, version, svc)
	if err != nil {
		return err
	}

	for name, value := range newParams {
		remoteKey := buildRemoteKey(s.context, file.Path, name, version)

		newParam := param{
			name:  remoteKey,
			value: value,
			pType: *input.Type,
		}

		if input.KeyId == nil {
			newParam.keyID = defaultPSKMSKey
		} else {
			newParam.keyID = *input.KeyId
		}

		if noChange(newParam, storedParams) {
			continue
		}

		v := formatValue(newParam.value)

		input.Name = &newParam.name
		input.Value = &v

		_, err := svc.PutParameter(&input)
		if err != nil {
			fmt.Fprintf(s.io.UserOutput, "parameter: %s", remoteKey)
			return err
		}
	}

	//------------------------------------------
	//- Delete removed params
	//------------------------------------------
	for _, remoteParam := range storedParams {

		param := strings.Replace(remoteParam.name, buildRemotePath(s.context, file.Path, version)+"/", "", 1)

		if !isParamIn(param, newParams) {
			if _, err := svc.DeleteParameter(&ssm.DeleteParameterInput{
				Name: aws.String(remoteParam.name),
			}); err != nil {
				fmt.Fprintf(s.io.UserOutput, "parameter: %s", remoteParam.name)
				return err
			}
		}
	}

	return nil
}

// Pull ...
func (s AWSParameterStore) Pull(file *catalog.File, version string) ([]byte, contract.Attributes, error) {

	svc := ssm.New(s.Session)

	storedParams, err := getStoredParams(s.context, file.Path, version, svc)
	if err != nil {
		return []byte{}, contract.Attributes{}, err
	}

	if len(storedParams) == 0 {
		return []byte{}, contract.Attributes{}, errors.New("parameters not found, verify AWS account and credentials")
	}

	var buffer bytes.Buffer

	for key, value := range toMap(storedParams) {
		name := key[strings.LastIndex(key, "/")+1 : len(key)]
		v := value

		if s.uo.StoreCommand == cmdRefFormat {
			buffer.WriteString(fmt.Sprintf("%s=%s\n", name, buildRemoteKey(s.context, file.Path, name, version)))
		} else {
			buffer.WriteString(fmt.Sprintf("%s=%s\n", name, v))
		}
	}

	return buffer.Bytes(), contract.Attributes{}, nil
}

// Purge ...
func (s AWSParameterStore) Purge(file *catalog.File, version string) error {

	svc := ssm.New(s.Session)

	storedParams, err := getStoredParams(s.context, file.Path, version, svc)
	if err != nil {
		return err
	}

	msg := ""
	for _, p := range storedParams {
		msg = fmt.Sprintf("%s  - %s\n", msg, p.name)
	}
	msg = fmt.Sprintf("%s \n  Delete parameters?", msg)

	if !prompt.Confirm(msg, prompt.Danger, s.io) {
		return errors.New("user aborted")
	}

	for _, p := range storedParams {
		if _, err := svc.DeleteParameter(&ssm.DeleteParameterInput{
			Name: aws.String(p.name),
		}); err != nil {
			fmt.Fprintf(s.io.UserOutput, "parameter: %s", p.name)
			return err
		}
	}

	return nil
}

// Changed ...
func (s AWSParameterStore) Changed(file *catalog.File, fileData []byte, version string) (time.Time, error) {
	config := gotenv.Parse(bytes.NewReader(fileData))

	svc := ssm.New(s.Session)

	storedParams, err := getStoredParams(s.context, file.Path, version, svc)
	if err != nil {
		return time.Time{}, err
	}

	changedParams := []param{}
	for _, p := range storedParams {

		for name, value := range config {
			remoteKey := buildRemoteKey(s.context, file.Path, name, version)

			if remoteKey == p.name && value != p.value {
				changedParams = append(changedParams, p)
			}
		}
	}

	return lastModified(changedParams), nil
}

func lastModified(params []param) time.Time {
	mostRecentlyModified := time.Time{}
	for _, sp := range params {
		if mostRecentlyModified.Before(sp.lastModified) {
			mostRecentlyModified = sp.lastModified
		}
	}

	return mostRecentlyModified
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
		MaxResults:       aws.Int64(50),
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

// The AWS api call for GetParametersByPath has higher rate limits than DescribeParameters. Using this function is more resiliant
// and should be used where possible. However, comparing KMS key ids requires using DescribeParameters; so, in a few cases, the
// call to describeParams is necesary.
func getStoredParamsByPath(svc *ssm.SSM, startsWith string, nextToken string, params []*ssm.Parameter) ([]*ssm.Parameter, error) {

	input := &ssm.GetParametersByPathInput{
		Recursive:      aws.Bool(true),
		Path:           aws.String(startsWith),
		WithDecryption: aws.Bool(true),
		MaxResults:     aws.Int64(10),
	}

	if len(nextToken) > 0 {
		input.SetNextToken(nextToken)
	}

	output, err := svc.GetParametersByPath(input)
	if err != nil {
		return nil, err
	}

	params = append(params, output.Parameters...)

	if output.NextToken == nil || len(*output.NextToken) == 0 {
		return params, nil
	}

	return getStoredParamsByPath(svc, startsWith, *output.NextToken, params)
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

func get(params []string, svc *ssm.SSM) ([]*ssm.Parameter, error) {

	if len(params) == 0 {
		return []*ssm.Parameter{}, nil
	}

	storedParams := []*ssm.Parameter{}

	// AWS Golang SDK request limit: 10
	chuckedParams := []string{}

	for i := 0; i < len(params); i++ {

		chuckedParams = append(chuckedParams, params[i])

		if len(chuckedParams) == 10 || i == len(params)-1 {

			output, err := svc.GetParameters(&ssm.GetParametersInput{
				Names:          aws.StringSlice(chuckedParams),
				WithDecryption: aws.Bool(true),
			})
			if err != nil {
				return []*ssm.Parameter{}, err
			}

			storedParams = append(storedParams, output.Parameters...)

			chuckedParams = []string{}
		}
	}

	return storedParams, nil
}

func toMap(params []param) map[string]string {
	data := map[string]string{}

	for _, p := range params {
		data[p.name] = p.value
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

func noChange(np param, params []param) bool {
	for _, p := range params {
		if p.name == np.name {
			return p.value == np.value && np.pType == p.pType && np.keyID == p.keyID
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

func getStoredParams(context, path, version string, svc *ssm.SSM) ([]param, error) {
	parameters := []param{}

	storedParams, err := getStoredParamsByPath(svc, buildRemotePath(context, path, version), "", []*ssm.Parameter{})
	if err != nil {
		return nil, err
	}

	for _, sp := range storedParams {
		parameters = append(parameters, param{
			name:         *sp.Name,
			value:        unformatValue(*sp.Value),
			pType:        *sp.Type,
			lastModified: *sp.LastModifiedDate,
		})
	}

	return parameters, nil
}

func getStoredParamsWithMetaData(context, path, version string, svc *ssm.SSM) ([]param, error) {

	parameters := []param{}

	storedParamData, err := describeParams(svc, buildRemotePath(context, path, version), "", []*ssm.ParameterMetadata{})
	if err != nil {
		return nil, err
	}

	params := map[string]*ssm.ParameterMetadata{}
	pNames := []string{}

	for _, p := range storedParamData {
		params[*p.Name] = p
		pNames = append(pNames, *p.Name)
	}

	storedParams, err := get(pNames, svc)
	if err != nil {
		return nil, err
	}

	for _, sp := range storedParams {

		p := param{
			name:         *sp.Name,
			value:        unformatValue(*sp.Value),
			pType:        *sp.Type,
			lastModified: *sp.LastModifiedDate,
		}

		if params[*sp.Name].KeyId != nil {
			p.keyID = strings.Replace(*params[*sp.Name].KeyId, "alias/", "", 1)
		}

		parameters = append(parameters, p)
	}

	return parameters, nil
}

type param struct {
	name  string
	value string

	keyID string
	pType string

	lastModified time.Time
}

func init() {
	s := new(AWSParameterStore)
	stores[s.Name()] = s
}
