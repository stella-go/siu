// Copyright 2010-2025 the original author or authors.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

// 	http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package autoconfig

import (
	"reflect"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/stella-go/siu/config"
	"github.com/stella-go/siu/t"
)

const (
	OssKey               = "oss"
	OssDisableKey        = OssKey + ".disable"
	OssEndpointKey       = OssKey + ".endpoint"
	OssAkKey             = OssKey + ".ak"
	OssSkKey             = OssKey + ".sk"
	OssRegionKey         = OssKey + ".region"
	OssDdisableSSLKey    = OssKey + ".disable-ssl"
	OssForcePathStyleKey = OssKey + ".force-path-style"

	OssOrder = 40
)

type AutoOss struct {
	Conf   config.TypedConfig `@siu:"name='environment',default='type'"`
	client *s3.S3
}

func (p *AutoOss) Condition() bool {
	_, ok1 := p.Conf.Get(OssKey)
	v, ok2 := p.Conf.GetBool(OssDisableKey)

	if ok2 && v {
		return false
	}
	return ok1
}

func (p *AutoOss) OnStart() error {
	endpoint, ok := p.Conf.GetString(OssEndpointKey)
	if !ok {
		return t.Errorf("oss endpoint not found")
	}
	ak, ok := p.Conf.GetString(OssAkKey)
	if !ok {
		return t.Errorf("oss ak not found")
	}
	sk, ok := p.Conf.GetString(OssSkKey)
	if !ok {
		return t.Errorf("oss sk not found")
	}
	region := p.Conf.GetStringOr(OssRegionKey, "default")
	disableSSL := p.Conf.GetBoolOr(OssDdisableSSLKey, false)
	forcePathStyle := p.Conf.GetBoolOr(OssForcePathStyleKey, true)
	creds := credentials.NewStaticCredentials(ak, sk, "")
	_, err := creds.Get()
	if err != nil {
		return t.Error(err)
	}

	config := &aws.Config{
		Region:           aws.String(region),
		Endpoint:         aws.String(endpoint),
		DisableSSL:       aws.Bool(disableSSL),
		Credentials:      creds,
		S3ForcePathStyle: aws.Bool(forcePathStyle),
	}
	sess := session.Must(session.NewSession(config))
	client := s3.New(sess)
	_, err = client.ListBuckets(nil)
	if err != nil {
		return t.Error(err)
	}
	p.client = client
	return nil
}

func (p *AutoOss) OnStop() error {
	return nil
}

func (*AutoOss) Order() int {
	return OssOrder
}

func (*AutoOss) Name() string {
	return OssKey
}

func (p *AutoOss) Named() map[string]interface{} {
	return map[string]interface{}{
		OssKey: p.client,
	}
}

func (p *AutoOss) Typed() map[reflect.Type]interface{} {
	refType := reflect.TypeOf((*s3.S3)(nil))
	return map[reflect.Type]interface{}{
		refType: p.client,
	}
}
