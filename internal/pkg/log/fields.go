/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package log

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Log Fields.
const (
	FieldUserLogLevel        = "userLogLevel"
	FieldID                  = "id"
	FieldName                = "name"
	FieldCommand             = "command"
	FieldHTTPStatus          = "httpStatus"
	FieldResponseBody        = "responseBody"
	FieldTopic               = "topic"
	FieldAdditionalMessage   = "additionalMessage"
	FieldHostURL             = "hostURL"
	FieldToken               = "token"
	FieldTotalRequests       = "totalRequests"
	FieldResponses           = "responses"
	FieldConcurrencyRequests = "concurrencyRequests"
	FieldWorkers             = "workers"
	FieldPath                = "path"
	FieldURL                 = "url"
	FieldJSON                = "json"
	FieldJSONResolution      = "jsonResolution"
	FieldSleep               = "sleep"
	FieldDuration            = "duration"
	FieldEvent               = "event"
	FieldDockerComposeCmd    = "dockerComposeCmd"
	FieldCertPoolSize        = "certPoolSize"
	FieldIDToken             = "idToken"
	FieldVPToken             = "vpToken"
	FieldTxID                = "transactionID"
	FieldPresDefID           = "presDefinitionID"
	FieldState               = "state"
	FieldProfileID           = "profileID"
)

// ObjectMarshaller uses reflection to marshal an object's fields.
type ObjectMarshaller struct {
	key string
	obj interface{}
}

// NewObjectMarshaller returns a new ObjectMarshaller.
func NewObjectMarshaller(key string, obj interface{}) *ObjectMarshaller {
	return &ObjectMarshaller{key: key, obj: obj}
}

// MarshalLogObject marshals the object's fields.
func (m *ObjectMarshaller) MarshalLogObject(e zapcore.ObjectEncoder) error {
	return e.AddReflected(m.key, m.obj)
}

// WithError sets the error field.
func WithError(err error) zap.Field {
	return zap.Error(err)
}

// WithUserLogLevel sets the user log level field.
func WithUserLogLevel(userLogLevel string) zap.Field {
	return zap.String(FieldUserLogLevel, userLogLevel)
}

// WithID sets the id field.
func WithID(id string) zap.Field {
	return zap.String(FieldID, id)
}

// WithName sets the name field.
func WithName(name string) zap.Field {
	return zap.String(FieldName, name)
}

// WithCommand sets the command field.
func WithCommand(command string) zap.Field {
	return zap.String(FieldCommand, command)
}

// WithHTTPStatus sets the http-status field.
func WithHTTPStatus(value int) zap.Field {
	return zap.Int(FieldHTTPStatus, value)
}

// WithResponseBody sets the response body field.
func WithResponseBody(value []byte) zap.Field {
	return zap.String(FieldResponseBody, string(value))
}

// WithTopic sets the topic field.
func WithTopic(value string) zap.Field {
	return zap.String(FieldTopic, value)
}

// WithAdditionalMessage sets the additional message field.
func WithAdditionalMessage(msg string) zap.Field {
	return zap.String(FieldAdditionalMessage, msg)
}

// WithHostURL sets the hostURL field.
func WithHostURL(hostURL string) zap.Field {
	return zap.String(FieldHostURL, hostURL)
}

// WithToken sets the token field.
func WithToken(token string) zap.Field {
	return zap.String(FieldToken, token)
}

// WithTotalRequests sets the total requests field.
func WithTotalRequests(totalRequests int) zap.Field {
	return zap.Int(FieldTotalRequests, totalRequests)
}

// WithResponses sets the responses field.
func WithResponses(responses int) zap.Field {
	return zap.Int(FieldResponses, responses)
}

// WithConcurrencyRequests sets the concurrency requests field.
func WithConcurrencyRequests(concurrencyReq int) zap.Field {
	return zap.Int(FieldConcurrencyRequests, concurrencyReq)
}

// WithWorkers sets the workers field.
func WithWorkers(workers int) zap.Field {
	return zap.Int(FieldWorkers, workers)
}

// WithPath sets the path field.
func WithPath(path string) zap.Field {
	return zap.String(FieldPath, path)
}

// WithURL sets the url field.
func WithURL(url string) zap.Field {
	return zap.String(FieldURL, url)
}

// WithJSON sets the json field.
func WithJSON(json string) zap.Field {
	return zap.String(FieldJSON, json)
}

// WithJSONResolution sets the Json resolution field.
func WithJSONResolution(jsonResolution string) zap.Field {
	return zap.String(FieldJSONResolution, jsonResolution)
}

// WithSleep sets the sleep field.
func WithSleep(sleep time.Duration) zap.Field {
	return zap.Duration(FieldSleep, sleep)
}

// WithDuration sets the duration field.
func WithDuration(value time.Duration) zap.Field {
	return zap.Duration(FieldDuration, value)
}

// WithEvent sets the event field.
func WithEvent(event interface{}) zap.Field {
	return zap.Inline(NewObjectMarshaller(FieldEvent, event))
}

// WithDockerComposeCmd sets the docker compose command field.
func WithDockerComposeCmd(cmd string) zap.Field {
	return zap.String(FieldDockerComposeCmd, cmd)
}

// WithCertPoolSize sets the certificate pool size field.
func WithCertPoolSize(poolSize int) zap.Field {
	return zap.Int(FieldCertPoolSize, poolSize)
}

// WithIDToken sets the id token field.
func WithIDToken(idToken string) zap.Field {
	return zap.String(FieldIDToken, idToken)
}

// WithVPToken sets the vp token field.
func WithVPToken(vpToken string) zap.Field {
	return zap.String(FieldVPToken, vpToken)
}

// WithTxID sets the transaction id field.
func WithTxID(txID string) zap.Field {
	return zap.String(FieldTxID, txID)
}

// WithPresDefID sets the presentation definition id field.
func WithPresDefID(id string) zap.Field {
	return zap.String(FieldPresDefID, id)
}

// WithState sets the state field.
func WithState(state string) zap.Field {
	return zap.String(FieldState, state)
}

// WithProfileID sets the presentation definition id field.
func WithProfileID(id string) zap.Field {
	return zap.String(FieldProfileID, id)
}
