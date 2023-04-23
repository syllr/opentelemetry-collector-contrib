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

	"github.com/volcengine/volc-sdk-golang/service/tls"
	"github.com/volcengine/volc-sdk-golang/service/tls/pb"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

const (
	TlsTraceTopicKey = "tracetopic"
	TlsAKKey         = "ak"
	TlsSKKey         = "sk"
	TlsRegionKey     = "region"
)

type TlsPusher struct {
	EndPoint      string
	ConfigAK      string
	ConfigSK      string
	ConfigTopicID string
	ConfigRegion  string
}

func (c *TlsPusher) pushTraceData(ctx context.Context, trace ptrace.Traces) error {
	logPb := traceDataToTlsData(trace)
	return c.pushToTls(ctx, logPb)
}

func (c *TlsPusher) pushToTls(ctx context.Context, logPb *pb.LogGroupList) error {
	var topic, ak, sk, region = c.ConfigTopicID, c.ConfigAK, c.ConfigSK, c.ConfigRegion
	ctxTopic, topicOk := ctx.Value(TlsTraceTopicKey).(string)
	ctxAk, akOk := ctx.Value(TlsAKKey).(string)
	ctxSk, skOk := ctx.Value(TlsSKKey).(string)
	ctxRegion, regionOk := ctx.Value(TlsRegionKey).(string)
	if topicOk && akOk && skOk && regionOk {
		topic, ak, sk, region = ctxTopic, ctxAk, ctxSk, ctxRegion
	}
	cli := tls.NewClient(c.EndPoint, ak, sk, "", region)
	_, err := cli.PutLogs(buildTlsPutLogsReq(topic, logPb))
	return err
}

func buildTlsPutLogsReq(topicID string, logGroup *pb.LogGroupList) *tls.PutLogsRequest {
	req := &tls.PutLogsRequest{
		TopicID: topicID,
		LogBody: logGroup,
	}
	return req
}
