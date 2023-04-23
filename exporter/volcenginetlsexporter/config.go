// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package volcenginetlsexporter

import "go.opentelemetry.io/collector/config/configopaque"

// Config defines configuration for volcengine tls exporter.

type Config struct {
	// TLS Endpoint, https://www.volcengine.com/docs/6470/73641
	Endpoint string `mapstructure:"endpoint"`
	// TLS topic id
	TopicID string `mapstructure:"topic_id"`
	// Volcengine access key
	AccessKey string `mapstructure:"access_key"`
	// Volcengine secret key
	SecretKey configopaque.String `mapstructure:"secret_key"`
	// Volcengine region
	Region string `mapstructure:"region"`
}
