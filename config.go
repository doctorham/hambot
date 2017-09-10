package main

import (
	"bytes"
	"encoding/json"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	awssession "github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// HamConfig contains settings used by the hamagrams website.
type HamConfig struct {
	Prompt string `json:"prompt"`
}

func (config *HamConfig) marshal() (data []byte, err error) {
	data, err = json.Marshal(config)
	if err != nil {
		return
	}

	data = append([]byte("config="), data...)
	data = append(data, byte(';'))
	return
}

// Upload uploads the configuration to the configured S3 bucket.
func (config *HamConfig) Upload(
	session *Session,
	then func(),
	catch func(error),
	finally func(),
) {
	defer func() {
		session.Callbacks <- finally
	}()

	onError := func(err error) {
		session.Callbacks <- func() {
			catch(err)
		}
	}

	var data []byte
	var err error
	if data, err = config.marshal(); err != nil {
		onError(err)
		return
	}

	var awsSession *awssession.Session
	awsSession, err = awssession.NewSession(&aws.Config{
		Region: aws.String(Settings.AwsRegion),
		Credentials: credentials.NewStaticCredentials(
			Settings.AwsAccessKey, Settings.AwsSecretAccessKey, ""),
	})
	if err != nil {
		onError(err)
		return
	}

	uploader := s3manager.NewUploader(awsSession)

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(Settings.AwsBucket),
		Key:         aws.String("config.js"),
		ContentType: aws.String("application/javascript"),
		ACL:         aws.String("public-read"),
		Body:        bytes.NewReader(data),
	})
	if err != nil {
		onError(err)
		return
	}

	session.Callbacks <- then
}
