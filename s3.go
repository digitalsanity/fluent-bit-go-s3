package main

import "github.com/aws/aws-sdk-go/aws"
import "github.com/aws/aws-sdk-go/aws/credentials"
import log "github.com/sirupsen/logrus"

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type format int

const (
	plainTextFormat format = iota
	gzipFormat
)

type s3Config struct {
	credentials      *credentials.Credentials
	bucket           *string
	s3prefix         *string
	region           *string
	compress         format
	endpoint         string
	logLevel         log.Level
	location         *time.Location
	autoCreateBucket bool
}

type S3Credential interface {
	GetCredentials(accessID, secretkey, credentials string) (*credentials.Credentials, error)
}

type s3PluginConfig struct{}

var s3Creds S3Credential = &s3PluginConfig{}

func (c *s3PluginConfig) GetCredentials(accessKeyID, secretKey, credential string) (*credentials.Credentials, error) {
	var creds *credentials.Credentials
	if credential != "" {
		creds = credentials.NewSharedCredentials(credential, "default")
		if _, err := creds.Get(); err != nil {
			fmt.Println("[SharedCredentials] ERROR:", err)
		} else {
			return creds, nil
		}
	} else if !(accessKeyID == "" && secretKey == "") {
		creds = credentials.NewStaticCredentials(accessKeyID, secretKey, "")
		if _, err := creds.Get(); err != nil {
			fmt.Println("[StaticCredentials] ERROR:", err)
		} else {
			return creds, nil
		}
	} else {
		creds = credentials.NewEnvCredentials()
		if _, err := creds.Get(); err != nil {
			fmt.Println("[EnvCredentials] ERROR:", err)
		} else {
			return creds, nil
		}
	}

	return nil, fmt.Errorf("Failed to create credentials")
}

func getS3Config(accessID, secretKey, credential, s3prefix, bucket, region, compress, endpoint, autoCreateBucket, logLevel, timeZone string) (*s3Config, error) {
	conf := &s3Config{}
	creds, err := s3Creds.GetCredentials(accessID, secretKey, credential)
	if err != nil {
		return nil, fmt.Errorf("Failed to create credentials")
	}
	conf.credentials = creds

	if bucket == "" {
		return nil, fmt.Errorf("Cannot specify empty string to bucket name")
	}
	conf.bucket = aws.String(bucket)

	if s3prefix == "" {
		return nil, fmt.Errorf("Cannot specify empty string to s3prefix")
	}
	conf.s3prefix = aws.String(s3prefix)

	if region == "" {
		return nil, fmt.Errorf("Cannot specify empty string to region")
	}
	conf.region = aws.String(region)

	switch compress {
	case "gzip":
		conf.compress = gzipFormat
	default:
		conf.compress = plainTextFormat
	}

	if endpoint != "" {
		if strings.HasSuffix(endpoint, "amazonaws.com") {
			return nil, fmt.Errorf("Endpoint is not supported for AWS S3. This parameter is intended for S3 compatible services. Use Region instead.")
		}
		conf.endpoint = endpoint
	}

	isAutoCreateBucket, err := strconv.ParseBool(autoCreateBucket)
	if err != nil {
		conf.autoCreateBucket = false
	} else {
		conf.autoCreateBucket = isAutoCreateBucket
	}

	if logLevel == "" {
		logLevel = "info"
	}
	var level log.Level
	if level, err = log.ParseLevel(logLevel); err != nil {
		return nil, fmt.Errorf("invalid log level: %v", logLevel)
	}
	conf.logLevel = level

	if timeZone != "" {
		loc, err := time.LoadLocation(timeZone)
		if err != nil {
			return nil, fmt.Errorf("invalid timeZone: %v", err)
		} else {
			conf.location = loc
		}
	} else {
		conf.location = time.Local
	}

	return conf, nil
}
