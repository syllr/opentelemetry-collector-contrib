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
	"encoding/hex"
	"encoding/json"
	"strconv"
	"time"

	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal/tracetranslator"
	"github.com/volcengine/volc-sdk-golang/service/tls/pb"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
	conventions "go.opentelemetry.io/collector/semconv/v1.6.1"
)

const (
	nameField              = "Name"
	traceIDField           = "TraceID"
	spanIDField            = "SpanID"
	traceStateField        = "TraceState"
	parentSpanIDField      = "ParentSpanID"
	kindField              = "Kind"
	startTimeField         = "Start"
	endTimeField           = "End"
	durationField          = "Duration"
	attributesField        = "Attributes"
	eventsField            = "Events"
	linksField             = "Links"
	timeField              = "Time"
	statusCodeField        = "StatusCode"
	statusDescriptionField = "StatusDescription"

	resourceField    = "Resource"
	hostField        = "Host"
	serviceNameField = "ServiceName"

	// shortcut for "InstrumentationLibrary.Name" "InstrumentationLibrary.Version"
	otlpNameField    = "OTLPName"
	otlpVersionField = "OTLPVersion"
)

// traceDataToLogService translates trace data into the LogService format.
func traceDataToTlsData(td ptrace.Traces) *pb.LogGroupList {
	var pbLogs []*pb.Log
	resourceSpansSlice := td.ResourceSpans()
	for i := 0; i < resourceSpansSlice.Len(); i++ {
		resourceSpans := resourceSpansSlice.At(i)
		tmpPbLogs := resourceSpansToLogServiceData(resourceSpans)
		pbLogs = append(pbLogs, tmpPbLogs...)
	}

	return &pb.LogGroupList{
		LogGroups: []*pb.LogGroup{{Logs: pbLogs}},
	}
}

func resourceSpansToLogServiceData(resourceSpans ptrace.ResourceSpans) []*pb.Log {
	var pbLogs []*pb.Log
	resourceLogContents := resourceToLogContents(resourceSpans.Resource())
	scopeSpansSlice := resourceSpans.ScopeSpans()
	for i := 0; i < scopeSpansSlice.Len(); i++ {
		scopeSpans := scopeSpansSlice.At(i)

		instrumentationScope := scopeSpans.Scope()
		instrumentationLibraryLogContents := instrumentationScopeToLogContents(instrumentationScope)

		spanSlice := scopeSpans.Spans()
		for j := 0; j < spanSlice.Len(); j++ {
			span := spanSlice.At(j)
			pbLog := spanToLogServiceData(span, resourceLogContents, instrumentationLibraryLogContents)
			if pbLog != nil {
				pbLogs = append(pbLogs, pbLog)
			}
		}
	}

	return pbLogs
}

func instrumentationScopeToLogContents(instrumentationScope pcommon.InstrumentationScope) []*pb.LogContent {
	logContents := make([]*pb.LogContent, 2)

	logContents[0] = &pb.LogContent{
		Key:   otlpNameField,
		Value: instrumentationScope.Name(),
	}
	logContents[1] = &pb.LogContent{
		Key:   otlpVersionField,
		Value: instrumentationScope.Version(),
	}

	return logContents
}

func spanToLogServiceData(span ptrace.Span, resourceLogContents, instrumentationLibraryLogContents []*pb.LogContent) *pb.Log {
	timeNano := int64(span.EndTimestamp())
	if timeNano == 0 {
		timeNano = time.Now().UnixNano()
	}

	preAllocCount := 16
	pbLog := pb.Log{
		Time:     timeNano / 1000 / 1000, // 毫秒时间戳
		Contents: make([]*pb.LogContent, 0, preAllocCount+len(resourceLogContents)+len(instrumentationLibraryLogContents)),
	}

	pbLog.Contents = append(pbLog.Contents, resourceLogContents...)
	pbLog.Contents = append(pbLog.Contents, instrumentationLibraryLogContents...)

	pbLog.Contents = append(pbLog.Contents, &pb.LogContent{
		Key:   nameField,
		Value: span.Name(),
	})
	pbLog.Contents = append(pbLog.Contents, &pb.LogContent{
		Key:   traceIDField,
		Value: TraceIDToHexOrEmptyString(span.TraceID()),
	})
	pbLog.Contents = append(pbLog.Contents, &pb.LogContent{
		Key:   spanIDField,
		Value: SpanIDToHexOrEmptyString(span.SpanID()),
	})
	pbLog.Contents = append(pbLog.Contents, &pb.LogContent{
		Key:   traceStateField,
		Value: span.TraceState().AsRaw(),
	})
	pbLog.Contents = append(pbLog.Contents, &pb.LogContent{
		Key:   parentSpanIDField,
		Value: SpanIDToHexOrEmptyString(span.ParentSpanID()),
	})
	pbLog.Contents = append(pbLog.Contents, &pb.LogContent{
		Key:   kindField,
		Value: spanKindToShortString(span.Kind()),
	})
	pbLog.Contents = append(pbLog.Contents, &pb.LogContent{
		Key:   startTimeField,
		Value: strconv.FormatUint(uint64(span.StartTimestamp()/1000), 10),
	})
	pbLog.Contents = append(pbLog.Contents, &pb.LogContent{
		Key:   endTimeField,
		Value: strconv.FormatUint(uint64(span.EndTimestamp()/1000), 10),
	})
	pbLog.Contents = append(pbLog.Contents, &pb.LogContent{
		Key:   durationField,
		Value: strconv.FormatUint(uint64((span.EndTimestamp()-span.StartTimestamp())/1000), 10),
	})
	attributesMap := span.Attributes().AsRaw()
	attributesJSONBytes, _ := json.Marshal(attributesMap)
	pbLog.Contents = append(pbLog.Contents, &pb.LogContent{
		Key:   attributesField,
		Value: string(attributesJSONBytes),
	})
	pbLog.Contents = append(pbLog.Contents, &pb.LogContent{
		Key:   eventsField,
		Value: eventsToString(span.Events()),
	})
	pbLog.Contents = append(pbLog.Contents, &pb.LogContent{
		Key:   linksField,
		Value: linksToString(span.Links()),
	})
	pbLog.Contents = append(pbLog.Contents, &pb.LogContent{
		Key:   statusCodeField,
		Value: statusCodeToShortString(span.Status().Code()),
	})
	pbLog.Contents = append(pbLog.Contents, &pb.LogContent{
		Key:   statusDescriptionField,
		Value: span.Status().Message(),
	})

	return &pbLog
}

func linksToString(spanLinkSlice ptrace.SpanLinkSlice) string {
	linkSlice := make([]map[string]interface{}, 0, spanLinkSlice.Len())

	for i := 0; i < spanLinkSlice.Len(); i++ {
		spanLink := spanLinkSlice.At(i)

		linkMap := map[string]interface{}{}
		linkMap[spanIDField] = SpanIDToHexOrEmptyString(spanLink.SpanID())
		linkMap[traceIDField] = TraceIDToHexOrEmptyString(spanLink.TraceID())
		linkMap[attributesField] = spanLink.Attributes().AsRaw()

		linkSlice = append(linkSlice, linkMap)
	}

	linkSliceJsonBytes, _ := json.Marshal(&linkSlice)
	return string(linkSliceJsonBytes)
}

func resourceToLogContents(resource pcommon.Resource) []*pb.LogContent {
	logContents := make([]*pb.LogContent, 3)

	attributesMap := resource.Attributes()
	if hostName, ok := attributesMap.Get(conventions.AttributeHostName); ok {
		logContents[0] = &pb.LogContent{
			Key:   hostField,
			Value: hostName.AsString(),
		}
	} else {
		logContents[0] = &pb.LogContent{
			Key:   hostField,
			Value: "",
		}
	}

	if serviceName, ok := attributesMap.Get(conventions.AttributeServiceName); ok {
		logContents[1] = &pb.LogContent{
			Key:   serviceNameField,
			Value: serviceName.AsString(),
		}
	} else {
		logContents[1] = &pb.LogContent{
			Key:   serviceNameField,
			Value: "",
		}
	}

	otherResource := map[string]interface{}{}
	attributesMap.Range(func(k string, v pcommon.Value) bool {
		if k == conventions.AttributeServiceName ||
			k == conventions.AttributeHostName {
			return true
		}
		otherResource[k] = v.AsString()
		return true
	})
	otherResourceJsonBytes, _ := json.Marshal(otherResource)
	logContents[2] = &pb.LogContent{
		Key:   resourceField,
		Value: string(otherResourceJsonBytes),
	}

	return logContents
}

func TraceIDToHexOrEmptyString(id pcommon.TraceID) string {
	if id.IsEmpty() {
		return ""
	}
	return hex.EncodeToString(id[:])
}

func SpanIDToHexOrEmptyString(id pcommon.SpanID) string {
	if id.IsEmpty() {
		return ""
	}
	return hex.EncodeToString(id[:])
}

func spanKindToShortString(kind ptrace.SpanKind) string {
	switch kind {
	case ptrace.SpanKindInternal:
		return string(tracetranslator.OpenTracingSpanKindInternal)
	case ptrace.SpanKindClient:
		return string(tracetranslator.OpenTracingSpanKindClient)
	case ptrace.SpanKindServer:
		return string(tracetranslator.OpenTracingSpanKindServer)
	case ptrace.SpanKindProducer:
		return string(tracetranslator.OpenTracingSpanKindProducer)
	case ptrace.SpanKindConsumer:
		return string(tracetranslator.OpenTracingSpanKindConsumer)
	default:
		return string(tracetranslator.OpenTracingSpanKindUnspecified)
	}
}

func statusCodeToShortString(code ptrace.StatusCode) string {
	switch code {
	case ptrace.StatusCodeError:
		return "ERROR"
	case ptrace.StatusCodeOk:
		return "OK"
	default:
		return "UNSET"
	}
}

func eventsToString(events ptrace.SpanEventSlice) string {
	eventSlice := make([]map[string]interface{}, 0, events.Len())

	for i := 0; i < events.Len(); i++ {
		spanEvent := events.At(i)

		eventMap := map[string]interface{}{}
		eventMap[nameField] = spanEvent.Name()
		eventMap[timeField] = spanEvent.Timestamp()
		eventMap[attributesField] = spanEvent.Attributes().AsRaw()

		eventSlice = append(eventSlice, eventMap)
	}

	eventSliceJsonBytes, _ := json.Marshal(&eventSlice)
	return string(eventSliceJsonBytes)
}
