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

import (
	"context"
	"errors"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
	"go.uber.org/zap"
)

// newTracesExporter return a new volcengine tls trace exporter.
func newTracesExporter(set exporter.CreateSettings, cfg component.Config) (exporter.Traces, error) {

	tlsClient, err := NewTlsPusher(cfg.(*Config), set.Logger)
	if err != nil {
		return nil, err
	}

	return exporterhelper.NewTracesExporter(
		context.TODO(),
		set,
		cfg,
		tlsClient.pushTraceData)
}

func NewTlsPusher(config *Config, logger *zap.Logger) (*TlsPusher, error) {
	if config == nil || config.Endpoint == "" || config.TopicID == "" || config.AccessKey == "" || config.SecretKey == "" || config.Region == "" {
		return nil, errors.New("missing volcengine tls trace params: Endpoint, TopicID, AccessKey, SecretKey, Region")
	}

	pusher := &TlsPusher{
		EndPoint:      config.Endpoint,
		ConfigAK:      config.AccessKey,
		ConfigSK:      string(config.SecretKey),
		ConfigTopicID: config.TopicID,
		ConfigRegion:  config.Region,
	}
	return pusher, nil
}
