package telemetry

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"testing"

	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

func TestDeliveryStartsNewTraceAndLinksRealProducer(t *testing.T) {
	recorder := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	t.Cleanup(func() { _ = provider.Shutdown(context.Background()) })
	tracer := provider.Tracer("book-observability")
	producerCtx, producer := tracer.Start(context.Background(), "producer")
	parent := CaptureProducer(producerCtx)
	if parent == "" {
		t.Fatal("producer traceparent was not captured")
	}
	headers := make(http.Header)
	InjectInternal(producerCtx, headers)
	extracted := propagation.TraceContext{}.Extract(context.Background(), propagation.HeaderCarrier(headers))
	source := producer.SpanContext()
	if got := trace.SpanContextFromContext(extracted); got.TraceID() != source.TraceID() || got.SpanID() != source.SpanID() {
		t.Fatal("synchronous propagation lost the original span identity")
	}
	unrelatedCtx, unrelated := tracer.Start(context.Background(), "unrelated")
	deliveryCtx, delivery := StartDelivery(unrelatedCtx, tracer, parent)
	if got := trace.SpanContextFromContext(deliveryCtx); got.TraceID() != delivery.SpanContext().TraceID() ||
		got.SpanID() != delivery.SpanContext().SpanID() {
		t.Fatal("returned context did not carry the consumer span")
	}
	delivery.End()
	unrelated.End()
	producer.End()
	for _, span := range recorder.Ended() {
		if span.Name() != "outbox.deliver" {
			continue
		}
		if span.Parent().IsValid() || span.SpanKind() != trace.SpanKindConsumer ||
			span.SpanContext().TraceID() == source.TraceID() ||
			span.SpanContext().TraceID() == unrelated.SpanContext().TraceID() {
			t.Fatal("asynchronous delivery inherited a producer or incidental parent")
		}
		links := span.Links()
		if len(links) != 1 || links[0].SpanContext.TraceID() != source.TraceID() ||
			links[0].SpanContext.SpanID() != source.SpanID() {
			t.Fatal("asynchronous consumer lost its causal producer link")
		}
		return
	}
	t.Fatal("real SDK did not record the consumer span")
}

func TestMalformedProducerDoesNotBecomeParentOrLink(t *testing.T) {
	recorder := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	t.Cleanup(func() { _ = provider.Shutdown(context.Background()) })
	ctx, span := StartDelivery(context.Background(), provider.Tracer("book-observability"), "invalid-private-value")
	var output bytes.Buffer
	LogCorrelation(ctx, slog.New(slog.NewJSONHandler(&output, nil)), Correlation{
		PaymentIntentID: "pi_1", AttemptID: "attempt_1", DecisionID: "decision_1", OperationID: "op_1",
	})
	span.End()
	ended := recorder.Ended()
	if len(ended) != 1 || ended[0].Parent().IsValid() || len(ended[0].Links()) != 0 {
		t.Fatal("malformed traceparent was trusted")
	}
	var record map[string]any
	if err := json.Unmarshal(output.Bytes(), &record); err != nil {
		t.Fatal(err)
	}
	for key, value := range map[string]string{
		"payment_intent_id": "pi_1", "attempt_id": "attempt_1", "decision_id": "decision_1", "operation_id": "op_1",
		"trace_id": span.SpanContext().TraceID().String(), "span_id": span.SpanContext().SpanID().String(),
	} {
		if record[key] != value {
			t.Fatalf("correlation %s = %v, want %s", key, record[key], value)
		}
	}
	if bytes.Contains(output.Bytes(), []byte("invalid-private-value")) {
		t.Fatal("untrusted raw parent value leaked into logs")
	}
}
