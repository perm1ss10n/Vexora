package mqtt

var (
    TopicTelemetry = "v1/dev/+/telemetry"
    TopicEvent     = "v1/dev/+/event"
    TopicState     = "v1/dev/+/state"
    TopicAck       = "v1/dev/+/ack"
    TopicCfgStatus = "v1/dev/+/cfg/status"
)

var AllTopics = []string{
    TopicTelemetry,
    TopicEvent,
    TopicState,
    TopicAck,
    TopicCfgStatus,
}