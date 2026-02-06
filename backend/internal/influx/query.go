package influx

import (
	"context"
	"fmt"
	"time"
)

type StateSnapshot struct {
	Ts     int64
	Link   string
	IP     string
	Uptime int64
}

type TelemetrySnapshot struct {
	Ts      int64
	Metrics map[string]float64
}

type TelemetryPoint struct {
	Ts    int64
	Value float64
}

func (c *Client) GetLastState(ctx context.Context, deviceID string) (*StateSnapshot, error) {
	if c == nil || c.cli == nil || deviceID == "" {
		return nil, nil
	}

	query := fmt.Sprintf(
		`from(bucket: "%s")
  |> range(start: -30d)
  |> filter(fn: (r) => r._measurement == "state" and r.deviceId == "%s")
  |> last()`,
		c.bucket,
		deviceID,
	)

	q := c.cli.QueryAPI(c.org)
	result, err := q.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query last state: %w", err)
	}
	defer result.Close()

	var snapshot *StateSnapshot
	var lastTime time.Time

	for result.Next() {
		record := result.Record()
		if snapshot == nil {
			snapshot = &StateSnapshot{}
		}

		if record.Time().After(lastTime) {
			lastTime = record.Time()
		}

		if link, ok := record.ValueByKey("link").(string); ok && link != "" {
			snapshot.Link = link
		}

		if field, ok := record.ValueByKey("_field").(string); ok {
			switch field {
			case "ip":
				if ip, ok := record.Value().(string); ok {
					snapshot.IP = ip
				}
			case "uptimeSec":
				if uptime, ok := asInt64(record.Value()); ok {
					snapshot.Uptime = uptime
				}
			}
		}
	}

	if err := result.Err(); err != nil {
		return nil, fmt.Errorf("query last state result: %w", err)
	}
	if snapshot == nil {
		return nil, nil
	}
	if !lastTime.IsZero() {
		snapshot.Ts = lastTime.Unix()
	}
	return snapshot, nil
}

func (c *Client) GetLastTelemetry(ctx context.Context, deviceID string) (*TelemetrySnapshot, error) {
	if c == nil || c.cli == nil || deviceID == "" {
		return nil, nil
	}

	query := fmt.Sprintf(
		`from(bucket: "%s")
  |> range(start: -30d)
  |> filter(fn: (r) => r._measurement == "telemetry" and r.deviceId == "%s")
  |> group(columns: ["metric"])
  |> last()`,
		c.bucket,
		deviceID,
	)

	q := c.cli.QueryAPI(c.org)
	result, err := q.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query last telemetry: %w", err)
	}
	defer result.Close()

	var snapshot *TelemetrySnapshot
	var lastTime time.Time

	for result.Next() {
		record := result.Record()
		metric, ok := record.ValueByKey("metric").(string)
		if !ok || metric == "" {
			continue
		}

		value, ok := asFloat64(record.Value())
		if !ok {
			continue
		}

		if snapshot == nil {
			snapshot = &TelemetrySnapshot{Metrics: map[string]float64{}}
		}
		snapshot.Metrics[metric] = value

		if record.Time().After(lastTime) {
			lastTime = record.Time()
		}
	}

	if err := result.Err(); err != nil {
		return nil, fmt.Errorf("query last telemetry result: %w", err)
	}
	if snapshot == nil {
		return nil, nil
	}
	if !lastTime.IsZero() {
		snapshot.Ts = lastTime.Unix()
	}
	return snapshot, nil
}

func (c *Client) GetTelemetrySeries(ctx context.Context, deviceID, metric string, from, to time.Time, limit int) ([]TelemetryPoint, error) {
	if c == nil || c.cli == nil || deviceID == "" || metric == "" {
		return nil, nil
	}

	start := from.UTC().Format(time.RFC3339)
	stop := to.UTC().Format(time.RFC3339)

	query := fmt.Sprintf(
		`from(bucket: "%s")
  |> range(start: %s, stop: %s)
  |> filter(fn: (r) => r._measurement == "telemetry" and r.deviceId == %q and r.metric == %q and r._field == "value")
  |> sort(columns: ["_time"])
  |> limit(n: %d)`,
		c.bucket,
		start,
		stop,
		deviceID,
		metric,
		limit,
	)

	q := c.cli.QueryAPI(c.org)
	result, err := q.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query telemetry series: %w", err)
	}
	defer result.Close()

	points := make([]TelemetryPoint, 0)
	for result.Next() {
		record := result.Record()
		value, ok := asFloat64(record.Value())
		if !ok {
			continue
		}
		points = append(points, TelemetryPoint{
			Ts:    record.Time().Unix(),
			Value: value,
		})
	}

	if err := result.Err(); err != nil {
		return nil, fmt.Errorf("query telemetry series result: %w", err)
	}
	return points, nil
}

func asInt64(value any) (int64, bool) {
	switch v := value.(type) {
	case int64:
		return v, true
	case int:
		return int64(v), true
	case float64:
		return int64(v), true
	case uint64:
		return int64(v), true
	default:
		return 0, false
	}
}

func asFloat64(value any) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case int64:
		return float64(v), true
	case int:
		return float64(v), true
	case uint64:
		return float64(v), true
	default:
		return 0, false
	}
}
